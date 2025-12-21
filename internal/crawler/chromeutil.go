package crawler

import (
	"context"

	"github.com/chromedp/chromedp"
)

// newChildBrowserContext creates a new child chromedp context derived from the
// given parent context. This allows navigation to occur without altering the
// state of the parent (current) browsing context.
// The caller is responsible for calling the returned cancel function.
func newChildBrowserContext(parent context.Context) (context.Context, context.CancelFunc) {
	return chromedp.NewContext(parent)
}
