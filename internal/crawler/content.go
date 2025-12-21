package crawler

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// fetchInnerHTMLWithNewContext opens the given URL in a new child chromedp context
// derived from parent, waits for the content container to be visible, and returns
// its innerHTML. The parent context state is preserved.
func fetchInnerHTMLWithNewContext(parent context.Context, url string, timeout time.Duration) (string, error) {
	// create a child context so that navigating to curURL does not affect the parent
	child, cancelChild := newChildBrowserContext(parent)
	defer cancelChild()

	if timeout > 0 {
		var toCancel context.CancelFunc
		child, toCancel = context.WithTimeout(child, timeout)
		defer toCancel()
	}

	if err := chromedp.Run(child, chromedp.Navigate(url)); err != nil {
		return "", err
	}
	if err := chromedp.Run(child, chromedp.WaitVisible(contentOuterSel, chromedp.ByQuery)); err != nil {
		return "", err
	}
	var inner string
	if err := chromedp.Run(child, chromedp.InnerHTML(contentOuterSel, &inner, chromedp.ByQuery)); err != nil {
		return "", err
	}
	return inner, nil
}
