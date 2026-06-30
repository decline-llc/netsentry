package stats

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

var defaultSeverities = []model.Severity{
	model.SeverityLow,
	model.SeverityMedium,
	model.SeverityHigh,
	model.SeverityCritical,
}

var matchDurationBuckets = [...]float64{
	0.0001,
	0.0005,
	0.001,
	0.005,
	0.01,
	0.025,
	0.05,
	0.1,
	0.25,
	0.5,
	1,
	2.5,
	5,
}

// Stats holds process-local counters exported through /api/metrics.
type Stats struct {
	framesTotal       atomic.Uint64
	controlFrames     atomic.Uint64
	packetsReceived   atomic.Uint64
	packetsProcessed  atomic.Uint64
	decodeErrors      atomic.Uint64
	alertsGenerated   atomic.Uint64
	matchCount        atomic.Uint64
	matchDurationNS   atomic.Uint64
	matchBuckets      [len(matchDurationBuckets)]atomic.Uint64
	workerPanics      atomic.Uint64
	alertWriteErrors  atomic.Uint64
	alertWriteCount   atomic.Uint64
	alertWriteNS      atomic.Uint64
	alertWriteBuckets [len(matchDurationBuckets)]atomic.Uint64

	mu               sync.RWMutex
	alertsBySeverity map[model.Severity]uint64
}

// New creates a Stats instance with stable severity labels preinitialized.
func New() *Stats {
	s := &Stats{alertsBySeverity: make(map[model.Severity]uint64)}
	for _, sev := range defaultSeverities {
		s.alertsBySeverity[sev] = 0
	}
	return s
}

func (s *Stats) IncFrame() {
	if s != nil {
		s.framesTotal.Add(1)
	}
}

func (s *Stats) IncControlFrame() {
	if s != nil {
		s.controlFrames.Add(1)
	}
}

func (s *Stats) IncPacketReceived() {
	if s != nil {
		s.packetsReceived.Add(1)
	}
}

func (s *Stats) IncPacketProcessed() {
	if s != nil {
		s.packetsProcessed.Add(1)
	}
}

func (s *Stats) IncDecodeError() {
	if s != nil {
		s.decodeErrors.Add(1)
	}
}

func (s *Stats) IncWorkerPanic() {
	if s != nil {
		s.workerPanics.Add(1)
	}
}

func (s *Stats) IncAlertWriteError() {
	if s != nil {
		s.alertWriteErrors.Add(1)
	}
}

func (s *Stats) ObserveMatchDuration(d time.Duration) {
	if s == nil {
		return
	}
	s.matchCount.Add(1)
	s.matchDurationNS.Add(uint64(d))
	observeDurationBucket(s.matchBuckets[:], d)
}

func (s *Stats) ObserveAlertWriteDuration(d time.Duration) {
	if s == nil {
		return
	}
	s.alertWriteCount.Add(1)
	s.alertWriteNS.Add(uint64(d))
	observeDurationBucket(s.alertWriteBuckets[:], d)
}

func observeDurationBucket(buckets []atomic.Uint64, d time.Duration) {
	seconds := d.Seconds()
	for i, bucket := range matchDurationBuckets {
		if seconds <= bucket {
			buckets[i].Add(1)
		}
	}
}

func (s *Stats) ObserveAlerts(alerts []*model.Alert) {
	if s == nil || len(alerts) == 0 {
		return
	}
	s.alertsGenerated.Add(uint64(len(alerts)))
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, alert := range alerts {
		if alert == nil {
			continue
		}
		sev := alert.Severity
		if sev == "" {
			sev = model.SeverityLow
		}
		s.alertsBySeverity[sev]++
	}
}

// Snapshot is a point-in-time view of exported counters.
type Snapshot struct {
	FramesTotal       uint64
	ControlFrames     uint64
	PacketsReceived   uint64
	PacketsProcessed  uint64
	DecodeErrors      uint64
	AlertsGenerated   uint64
	MatchCount        uint64
	MatchDurationNS   uint64
	MatchBuckets      []HistogramBucket
	WorkerPanics      uint64
	AlertWriteErrors  uint64
	AlertWriteCount   uint64
	AlertWriteNS      uint64
	AlertWriteBuckets []HistogramBucket
	AlertsBySeverity  map[model.Severity]uint64
}

type HistogramBucket struct {
	Le    float64
	Count uint64
}

func (s *Stats) Snapshot() Snapshot {
	if s == nil {
		return Snapshot{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	bySeverity := make(map[model.Severity]uint64, len(s.alertsBySeverity))
	for sev, count := range s.alertsBySeverity {
		bySeverity[sev] = count
	}
	buckets := make([]HistogramBucket, 0, len(matchDurationBuckets))
	alertWriteBuckets := make([]HistogramBucket, 0, len(matchDurationBuckets))
	for i, upperBound := range matchDurationBuckets {
		buckets = append(buckets, HistogramBucket{
			Le:    upperBound,
			Count: s.matchBuckets[i].Load(),
		})
		alertWriteBuckets = append(alertWriteBuckets, HistogramBucket{
			Le:    upperBound,
			Count: s.alertWriteBuckets[i].Load(),
		})
	}
	return Snapshot{
		FramesTotal:       s.framesTotal.Load(),
		ControlFrames:     s.controlFrames.Load(),
		PacketsReceived:   s.packetsReceived.Load(),
		PacketsProcessed:  s.packetsProcessed.Load(),
		DecodeErrors:      s.decodeErrors.Load(),
		AlertsGenerated:   s.alertsGenerated.Load(),
		MatchCount:        s.matchCount.Load(),
		MatchDurationNS:   s.matchDurationNS.Load(),
		MatchBuckets:      buckets,
		WorkerPanics:      s.workerPanics.Load(),
		AlertWriteErrors:  s.alertWriteErrors.Load(),
		AlertWriteCount:   s.alertWriteCount.Load(),
		AlertWriteNS:      s.alertWriteNS.Load(),
		AlertWriteBuckets: alertWriteBuckets,
		AlertsBySeverity:  bySeverity,
	}
}

// RenderPrometheus emits a minimal Prometheus text exposition without adding a
// runtime dependency before the API package is split out.
func RenderPrometheus(snapshot Snapshot, gauges map[string]float64) string {
	var b strings.Builder
	writeCounter(&b, "netsentry_frames_total", "UDS frames received.", snapshot.FramesTotal)
	writeCounter(&b, "netsentry_control_frames_total", "UDS control frames received.", snapshot.ControlFrames)
	writeCounter(&b, "netsentry_packets_received_total", "Packet frames accepted by the receiver.", snapshot.PacketsReceived)
	writeCounter(&b, "netsentry_packets_processed_total", "Packets processed by the pipeline.", snapshot.PacketsProcessed)
	writeCounter(&b, "netsentry_decode_errors_total", "UDS frames rejected during decode or validation.", snapshot.DecodeErrors)
	writeCounter(&b, "netsentry_alerts_generated_total", "Alerts generated by the pipeline.", snapshot.AlertsGenerated)
	writeCounter(&b, "netsentry_rule_match_total", "Rule match operations executed.", snapshot.MatchCount)
	writeCounter(&b, "netsentry_rule_match_duration_seconds_total", "Total time spent matching packets.", float64(snapshot.MatchDurationNS)/float64(time.Second))
	writeHistogram(&b, "netsentry_rule_match_duration_seconds", "Rule match duration distribution.", snapshot.MatchBuckets, snapshot.MatchCount, float64(snapshot.MatchDurationNS)/float64(time.Second))
	writeCounter(&b, "netsentry_worker_panics_total", "Pipeline worker panics recovered.", snapshot.WorkerPanics)
	writeCounter(&b, "netsentry_alert_write_errors_total", "Alert write failures.", snapshot.AlertWriteErrors)
	writeCounter(&b, "netsentry_alert_write_duration_seconds_total", "Total time spent writing alert batches.", float64(snapshot.AlertWriteNS)/float64(time.Second))
	writeHistogram(&b, "netsentry_alert_write_duration_seconds", "Alert write duration distribution.", snapshot.AlertWriteBuckets, snapshot.AlertWriteCount, float64(snapshot.AlertWriteNS)/float64(time.Second))

	severities := make([]string, 0, len(snapshot.AlertsBySeverity))
	for sev := range snapshot.AlertsBySeverity {
		severities = append(severities, string(sev))
	}
	sort.Strings(severities)
	b.WriteString("# HELP netsentry_alerts_by_severity_total Alerts generated by severity.\n")
	b.WriteString("# TYPE netsentry_alerts_by_severity_total counter\n")
	for _, sev := range severities {
		fmt.Fprintf(&b, "netsentry_alerts_by_severity_total{severity=%q} %d\n", sev, snapshot.AlertsBySeverity[model.Severity(sev)])
	}

	names := make([]string, 0, len(gauges))
	for name := range gauges {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		writeGauge(&b, name, gauges[name])
	}
	return b.String()
}

func writeCounter[T ~uint64 | ~float64](b *strings.Builder, name, help string, value T) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s counter\n", name)
	fmt.Fprintf(b, "%s %v\n", name, value)
}

func writeGauge(b *strings.Builder, name string, value float64) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, gaugeHelp(name))
	fmt.Fprintf(b, "# TYPE %s gauge\n", name)
	fmt.Fprintf(b, "%s %v\n", name, value)
}

func writeHistogram(b *strings.Builder, name, help string, buckets []HistogramBucket, count uint64, sum float64) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s histogram\n", name)
	for _, bucket := range buckets {
		fmt.Fprintf(b, "%s_bucket{le=%q} %d\n", name, formatBucket(bucket.Le), bucket.Count)
	}
	fmt.Fprintf(b, "%s_bucket{le=\"+Inf\"} %d\n", name, count)
	fmt.Fprintf(b, "%s_sum %v\n", name, sum)
	fmt.Fprintf(b, "%s_count %d\n", name, count)
}

func formatBucket(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", value), "0"), ".")
}

func gaugeHelp(name string) string {
	switch name {
	case "netsentry_alerts_current":
		return "Current number of aggregated alerts in storage."
	case "netsentry_capture_avg_json_serialize_seconds":
		return "Latest capture-reported average JSON serialization time."
	case "netsentry_capture_connected":
		return "Whether the capture heartbeat is currently fresh."
	case "netsentry_capture_heartbeat_age_seconds":
		return "Age of the latest capture heartbeat."
	case "netsentry_capture_packets_dropped":
		return "Latest capture-reported dropped packet count."
	case "netsentry_capture_packets_sent":
		return "Latest capture-reported sent packet count."
	case "netsentry_capture_parse_errors":
		return "Latest capture-reported parse error count."
	case "netsentry_capture_uds_write_errors":
		return "Latest capture-reported UDS write error count."
	case "netsentry_packet_queue_depth":
		return "Current packet queue depth."
	case "netsentry_rules_loaded":
		return "Current number of loaded rules."
	case "netsentry_storage_available_bytes":
		return "Available bytes on the alert storage filesystem."
	default:
		return "Gauge value."
	}
}
