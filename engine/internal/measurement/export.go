// Package measurement exports opt-in engine lifecycle evidence, never SLO passes.
package measurement

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

// Exporter serializes writes across workers without retaining packet identities.
// Duplicate identities are rejected later by the disk-backed ledger adapter.
type Exporter struct {
	mu     sync.Mutex
	file   *os.File
	dir    string
	runID  string
	origin time.Time
	digest hash.Hash
	rows   int64
	bytes  int64
	err    error
	closed bool
	cancel func()
}

type event struct {
	Kind     string `json:"kind"`
	PacketID string `json:"packet_id,omitempty"`
	EventID  string `json:"event_id,omitempty"`
	AtNS     int64  `json:"at_ns"`
}

// Open exclusively creates a private directory; the parent must already exist.
func Open(dir, runID string, origin time.Time, cancel func()) (*Exporter, error) {
	if !model.ValidSLOIdentifier(runID) || origin.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) ||
		!origin.Before(time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return nil, fmt.Errorf("measurement requires a bounded run ID and origin in [2000,2100)")
	}
	if err := os.Mkdir(dir, 0o700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(dir, "events.jsonl"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, err
	}
	return &Exporter{file: file, dir: dir, runID: runID, origin: origin, digest: sha256.New(), cancel: cancel}, nil
}

// EventID is a measurement correlation ID, separate from storage event/row IDs.
// The fixture oracle must use this same UTF-8 packet + NUL + rule derivation.
func EventID(packetID, ruleID string) string {
	sum := sha256.Sum256([]byte(packetID + "\x00" + ruleID))
	return "slo_" + hex.EncodeToString(sum[:])
}

func (e *Exporter) fail(err error) error {
	if e.err == nil {
		e.err = err
		if e.cancel != nil {
			e.cancel()
		}
	}
	return e.err
}

func (e *Exporter) emit(pkt *model.PacketInfo, kind string, alerts []*model.Alert, at time.Time) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.err != nil {
		return e.err
	}
	if e.closed {
		return e.fail(fmt.Errorf("measurement exporter already closed"))
	}
	if pkt == nil {
		return e.fail(fmt.Errorf("measurement packet is nil"))
	}
	if err := pkt.SLO.Validate(); err != nil {
		return e.fail(err)
	}
	if pkt.SLO.RunID != e.runID {
		return e.fail(fmt.Errorf("measurement run ID mismatch"))
	}
	arrival := time.Unix(0, pkt.SLO.ArrivalUnixNS)
	if arrival.Before(e.origin) || arrival.After(at) {
		return e.fail(fmt.Errorf("measurement arrival outside run origin/current observation"))
	}
	if kind == "arrival" {
		at = arrival
	}
	ns := at.Sub(e.origin).Nanoseconds()
	if ns < 0 || !e.origin.Add(time.Duration(ns)).Equal(at) {
		return e.fail(fmt.Errorf("measurement timestamp outside representable run interval"))
	}
	rows := []event{{Kind: kind, PacketID: pkt.SLO.PacketID, AtNS: ns}}
	if kind == "durable" {
		rows = nil
		seen := make(map[string]bool, len(alerts))
		for _, alert := range alerts {
			if alert == nil {
				continue
			}
			if !model.ValidSLOIdentifier(alert.RuleID) || seen[alert.RuleID] {
				return e.fail(fmt.Errorf("invalid or repeated measurement rule ID"))
			}
			seen[alert.RuleID] = true
			rows = append(rows, event{Kind: kind, EventID: EventID(pkt.SLO.PacketID, alert.RuleID), AtNS: ns})
		}
	}
	for _, row := range rows {
		raw, err := json.Marshal(row)
		if err != nil {
			return e.fail(err)
		}
		raw = append(raw, '\n')
		n, err := e.file.Write(raw)
		if err != nil {
			return e.fail(err)
		}
		if n != len(raw) {
			return e.fail(io.ErrShortWrite)
		}
		_, _ = e.digest.Write(raw)
		e.bytes += int64(n)
		e.rows++
	}
	return nil
}

func (e *Exporter) Arrival(pkt *model.PacketInfo) error {
	return e.emit(pkt, "arrival", nil, time.Now())
}

func (e *Exporter) Durable(pkt *model.PacketInfo, alerts []*model.Alert, at time.Time) error {
	return e.emit(pkt, "durable", alerts, at)
}

func (e *Exporter) Processed(pkt *model.PacketInfo, at time.Time) error {
	return e.emit(pkt, "processed", nil, at)
}

func syncDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return errors.Join(f.Sync(), f.Close())
}

// Close must follow worker termination. A close receipt only establishes that
// exported bytes were synchronized; it never establishes complete offered load.
func (e *Exporter) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return e.err
	}
	e.closed = true
	err := errors.Join(e.file.Sync(), e.file.Close())
	if err != nil {
		e.fail(err)
	}
	if e.err != nil {
		return e.err
	}
	receipt := map[string]any{
		"schema_version": 1, "run_id": e.runID, "origin": e.origin.UTC().Format(time.RFC3339Nano),
		"export_closed": true, "execution_complete": false, "slo_compliance_asserted": false,
		"departmental_review_required": true, "file": "events.jsonl", "rows": e.rows,
		"bytes": e.bytes, "sha256": hex.EncodeToString(e.digest.Sum(nil)),
		"clock":            "supplied_live_arrival_and_engine_unix_wall_clock",
		"durable_boundary": "successful_full_synchronous_WAL_WriteBatch_return_upper_bound",
		"event_identity":   "slo_ plus SHA256(UTF8(packet_id) + NUL + UTF8(rule_id))",
	}
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return e.fail(err)
	}
	raw = append(raw, '\n')
	f, err := os.CreateTemp(e.dir, ".close-")
	if err != nil {
		return e.fail(err)
	}
	name := f.Name()
	defer os.Remove(name)
	n, writeErr := f.Write(raw)
	if n != len(raw) && writeErr == nil {
		writeErr = io.ErrShortWrite
	}
	if err := errors.Join(writeErr, f.Sync(), f.Close()); err != nil {
		return e.fail(err)
	}
	if err := os.Link(name, filepath.Join(e.dir, "close.json")); err != nil {
		return e.fail(err)
	}
	if err := syncDirectory(e.dir); err != nil {
		return e.fail(err)
	}
	if err := syncDirectory(filepath.Dir(e.dir)); err != nil {
		return e.fail(err)
	}
	return nil
}
