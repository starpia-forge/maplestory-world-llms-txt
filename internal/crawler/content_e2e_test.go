//go:build e2e

package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestFetchInnerHTMLWithNewContext_E2E(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body><div class="text_content_container"><div class="text_content">Hello</div></div></body></html>`))
	}))
	defer srv.Close()

	// Create a parent browser context (allocator + context) like the app would do.
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), chromedp.DefaultExecAllocatorOptions[:]...)
	defer cancelAlloc()
	parent, cancelParent := chromedp.NewContext(allocCtx)
	defer cancelParent()

	inner, err := fetchInnerHTMLWithNewContext(parent, srv.URL, 10*time.Second)
	if err != nil {
		t.Fatalf("fetchInnerHTMLWithNewContext error: %v", err)
	}
	if inner == "" {
		t.Fatalf("expected non-empty innerHTML")
	}
	if !strings.Contains(inner, "Hello") {
		t.Fatalf("expected innerHTML to contain 'Hello', got: %q", inner)
	}
}
