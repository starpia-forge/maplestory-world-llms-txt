package logger

import (
	"context"
	"log/slog"
	"testing"
)

// capHandler is a simple slog.Handler that captures records for assertions.
type capHandler struct{ recs []slog.Record }

func (h *capHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }
func (h *capHandler) Handle(ctx context.Context, r slog.Record) error {
	h.recs = append(h.recs, r.Clone())
	return nil
}
func (h *capHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *capHandler) WithGroup(name string) slog.Handler       { return h }

func attrsToMap(r slog.Record) map[string]any {
	m := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		m[a.Key] = a.Value.Any()
		return true
	})
	return m
}

func TestLogParsedDoc_EmitsInfoWithExpectedAttrs(t *testing.T) {
	h := &capHandler{}
	l := slog.New(h)

	LogParsedDoc(l, "Hello", "https://example.com/docs?postId=123")

	if len(h.recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(h.recs))
	}
	rec := h.recs[0]
	if rec.Level != slog.LevelInfo {
		t.Fatalf("expected info level, got %v", rec.Level)
	}
	if rec.Message != "parsed_doc" {
		t.Fatalf("expected message 'parsed_doc', got %q", rec.Message)
	}
	got := attrsToMap(rec)
	if got["title"] != "Hello" || got["url"] != "https://example.com/docs?postId=123" {
		t.Fatalf("unexpected attrs: %+v", got)
	}
}
