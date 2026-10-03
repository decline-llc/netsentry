package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"reflect"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestAuditStatusMatchesCommittedHTTPResponse(t *testing.T) {
	for _, tc := range []struct {
		name      string
		headers   []int
		bodyFirst bool
		bodyAfter bool
		status    int
		infos     []int
		get       bool
		noLogger  bool
	}{
		{name: "first_unauthorized", headers: []int{401, 200}, bodyAfter: true, status: 401},
		{name: "first_success", headers: []int{201, 401}, bodyAfter: true, status: 201},
		{name: "first_failure", headers: []int{503, 201}, bodyAfter: true, status: 503},
		{name: "no_content", headers: []int{204, 500}, status: 204},
		{name: "implicit_body_then_header", headers: []int{401}, bodyFirst: true, status: 200},
		{name: "no_write", status: 200},
		{name: "informational_then_explicit", headers: []int{100, 102, 103, 401, 200}, bodyAfter: true, status: 401, infos: []int{100, 102, 103}},
		{name: "informational_then_body", headers: []int{103, 102}, bodyAfter: true, status: 200, infos: []int{103, 102}},
		{name: "informational_then_return", headers: []int{103, 103}, status: 200, infos: []int{103, 103}},
		{name: "informational_then_created", headers: []int{103, 201, 500}, bodyAfter: true, status: 201, infos: []int{103}},
		{name: "get_is_not_logged", headers: []int{103, 401, 200}, bodyAfter: true, status: 401, infos: []int{103}, get: true},
		{name: "absent_logger", headers: []int{401, 200}, bodyAfter: true, status: 401, noLogger: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core, observed := observer.New(zap.InfoLevel)
			api := &Server{opts: Options{AuditLogger: zap.New(core)}}
			if tc.noLogger {
				api.opts.AuditLogger = nil
			}
			const body = "mutation result"
			handler := api.audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
				writeBody := func() {
					if n, err := w.Write([]byte(body)); err != nil || n != len(body) {
						t.Errorf("body write=%d error=%v", n, err)
					}
				}
				if tc.bodyFirst {
					writeBody()
				}
				for _, status := range tc.headers {
					w.WriteHeader(status)
				}
				if tc.bodyAfter {
					writeBody()
				}
			}))
			// Actual net/http is required here: ResponseRecorder does not model
			// informational responses. Completion also makes log reads deterministic.
			done := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(done)
				handler.ServeHTTP(w, r)
			}))
			defer server.Close()
			client := server.Client()
			client.Timeout = 5 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var infos []int
			ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
				Got1xxResponse: func(code int, _ textproto.MIMEHeader) error {
					infos = append(infos, code)
					return nil
				},
			})
			method := http.MethodPost
			if tc.get {
				method = http.MethodGet
			}
			req, err := http.NewRequestWithContext(ctx, method, server.URL+"/api/rules/reload", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Request-ID", "audit-status-"+tc.name)
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			gotBody, readErr := io.ReadAll(resp.Body)
			closeErr := resp.Body.Close()
			if readErr != nil || closeErr != nil {
				t.Fatalf("response read=%v close=%v", readErr, closeErr)
			}
			select {
			case <-done:
			case <-ctx.Done():
				t.Fatalf("audit handler did not complete: %v", ctx.Err())
			}
			wantBody := ""
			if tc.bodyFirst || tc.bodyAfter {
				wantBody = body
			}
			if resp.StatusCode != tc.status || string(gotBody) != wantBody || !reflect.DeepEqual(infos, tc.infos) {
				t.Fatalf("wire status=%d body=%q informational=%v, want %d/%q/%v", resp.StatusCode, gotBody, infos, tc.status, wantBody, tc.infos)
			}
			if resp.Header.Get("X-Request-ID") != req.Header.Get("X-Request-ID") {
				t.Fatalf("wire request ID=%q", resp.Header.Get("X-Request-ID"))
			}
			entries := observed.FilterMessage("api audit").All()
			if tc.get || tc.noLogger {
				if len(entries) != 0 {
					t.Fatalf("unexpected audit entries: %+v", entries)
				}
				return
			}
			if len(entries) != 1 {
				t.Fatalf("audit entries=%d, want one", len(entries))
			}
			fields := entries[0].ContextMap()
			if fields["status"] != int64(resp.StatusCode) || fields["authorized"] != (resp.StatusCode != http.StatusUnauthorized) ||
				fields["request_id"] != req.Header.Get("X-Request-ID") || fields["method"] != method ||
				fields["path"] != "/api/rules/reload" || fields["target"] != "rules" {
				t.Fatalf("audit differs from committed response: %+v", fields)
			}
		})
	}
}

// Only the forwarding assertion uses a spy; the status/log contract above
// reaches a real server and client instead of substituting HTTP semantics.
type auditHeaderCalls struct {
	header http.Header
	codes  []int
}

func (w *auditHeaderCalls) Header() http.Header         { return w.header }
func (w *auditHeaderCalls) Write(p []byte) (int, error) { return len(p), nil }
func (w *auditHeaderCalls) WriteHeader(code int)        { w.codes = append(w.codes, code) }

func TestAuditResponseWriterHeaderForwardingAndSwitchingProtocols(t *testing.T) {
	for _, tc := range []struct {
		name    string
		headers []int
		want    []int
		status  int
	}{
		{"switching_is_terminal", []int{100, 103, 101, 200, 102, 99}, []int{100, 103, 101}, 101},
		{"unauthorized_is_terminal", []int{103, 102, 401, 200, 100}, []int{103, 102, 401}, 401},
		{"created_is_terminal", []int{201, 500}, []int{201}, 201},
		{"informational_is_not_terminal", []int{100, 102, 103}, []int{100, 102, 103}, 0},
		{"no_header", nil, nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			underlying := &auditHeaderCalls{header: make(http.Header)}
			writer := &auditResponseWriter{ResponseWriter: underlying}
			for _, code := range tc.headers {
				writer.WriteHeader(code)
			}
			if writer.status != tc.status || !reflect.DeepEqual(underlying.codes, tc.want) {
				t.Fatalf("recorded=%d forwarded=%v, want %d/%v", writer.status, underlying.codes, tc.status, tc.want)
			}
		})
	}
}

func TestAuditResponseWriterDoesNotRecordRejectedHeader(t *testing.T) {
	for _, invalid := range []int{99, 1000} {
		underlying := httptest.NewRecorder()
		writer := &auditResponseWriter{ResponseWriter: underlying}
		var rejected any
		func() {
			// This recovery asserts the underlying standard writer's deliberate
			// invalid-status panic; runtime middleware must not recover it.
			defer func() { rejected = recover() }()
			writer.WriteHeader(invalid)
		}()
		if rejected == nil || writer.status != 0 {
			t.Fatalf("invalid %d: rejection=%v recorded=%d, want panic and no commitment", invalid, rejected, writer.status)
		}
		writer.WriteHeader(http.StatusUnauthorized)
		writer.WriteHeader(http.StatusOK)
		if writer.status != http.StatusUnauthorized || underlying.Code != http.StatusUnauthorized {
			t.Fatalf("after rejected %d: recorded=%d wire=%d, want 401", invalid, writer.status, underlying.Code)
		}
	}
}
