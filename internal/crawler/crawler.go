package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"maplestory-world-llms-txt/internal/logger"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

const (
	navContainerSel = "#App > main > div.contents_wrap > div.tree_view_container"
	// contentOuterSel targets the container div used when opening curURL in a separate context
	// per requirement: search for div elements with class text_content_container
	contentOuterSel     = "div.text_content_container"
	contentContainerSel = "#App > main > div.contents_wrap > div.renderContent > div.text_content_container > div.text_content"
	titleSel            = "#App > main > div.contents_wrap > div.renderContent h1"
)

// Crawler holds reusable configuration between runs.
type Crawler struct {
	ClickDelay     time.Duration
	Limit          int
	OverallTimeout time.Duration
	Headless       bool
}

// Option configures a Crawler.
type Option func(*Crawler)

// WithClickDelay sets the delay between clicks. Negative values are clamped to 0.
func WithClickDelay(d time.Duration) Option {
	if d < 0 {
		d = 0
	}
	return func(c *Crawler) { c.ClickDelay = d }
}

// WithLimit sets the maximum number of documents to crawl. Negative values become 0 (no limit).
func WithLimit(n int) Option {
	if n < 0 {
		n = 0
	}
	return func(c *Crawler) { c.Limit = n }
}

// WithOverallTimeout sets the overall crawling timeout. Negative values are clamped to 0 (no timeout).
func WithOverallTimeout(d time.Duration) Option {
	if d < 0 {
		d = 0
	}
	return func(c *Crawler) { c.OverallTimeout = d }
}

// WithHeadless sets whether to run Chrome in headless mode.
func WithHeadless(b bool) Option { return func(c *Crawler) { c.Headless = b } }

// NewCrawler constructs a Crawler using the provided functional options.
func NewCrawler(opts ...Option) *Crawler {
	c := &Crawler{}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

// Run crawls the documentation starting at startURL and writes results to outPath
// using the given format.
func (c *Crawler) Run(url string) ([]Document, error) {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", c.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36"),
	)

	ctx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	if c.OverallTimeout > 0 {
		var toCancel context.CancelFunc
		ctx, toCancel = context.WithTimeout(ctx, c.OverallTimeout)
		defer toCancel()
	}

	// Navigate to start URL
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		return []Document{}, err
	}
	if err := waitVisible(ctx, navContainerSel, 30*time.Second); err != nil {
		return []Document{}, fmt.Errorf("navigation container not visible: %w", err)
	}

	visited := make(map[string]bool)
	var docs []Document
	backoff := NewBackoff(500*time.Millisecond, 20*time.Second, 2.0, 0.2)

	// 1) Expansion phase: click any closed node that has children until none remain.
	for {
		_ = scrollMenuToEnd(ctx)

		var nodes []*cdp.Node
		if err := chromedp.Run(ctx, chromedp.Nodes(navContainerSel+" *", &nodes, chromedp.ByQueryAll)); err != nil {
			return []Document{}, fmt.Errorf("query nodes: %w", err)
		}

		expanded := false
		for _, n := range nodes {
			// Condition: span.inactiveDot.isHavingChildren and NOT .isHavingChildrenAndOpen
			if strings.EqualFold(n.LocalName, "span") && hasAllClasses(n, "inactiveDot", "isHavingChildren") && !hasClass(n, "isHavingChildrenAndOpen") {
				_ = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					return dom.ScrollIntoViewIfNeeded().WithNodeID(n.NodeID).Do(ctx)
				}))
				if err := chromedp.Run(ctx, chromedp.MouseClickNode(n)); err != nil {
					continue
				}
				time.Sleep(c.ClickDelay)
				expanded = true
			}
		}
		if !expanded {
			break
		}
	}

	// 2) Collection phase: among div.inactiveDepth, target only nodes that have any visible text
	// and are NOT .activeParent; click and collect.
	_ = scrollMenuToEnd(ctx)
	var leafNodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(navContainerSel+" div.inactiveDepth", &leafNodes, chromedp.ByQueryAll)); err != nil {
		return []Document{}, fmt.Errorf("query leaf nodes: %w", err)
	}

	for _, n := range leafNodes {
		if hasClass(n, "activeParent") {
			continue
		}
		// Check for visible text: inspect outerHTML and see if there is any visible text after stripping tags.
		hasText := false
		_ = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			html, err := dom.GetOuterHTML().WithNodeID(n.NodeID).Do(ctx)
			if err != nil {
				return nil
			}
			if hasAnyTextInHTML(html) {
				hasText = true
			}
			return nil
		}))
		if !hasText {
			continue
		}

		// Click the node
		_ = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return dom.ScrollIntoViewIfNeeded().WithNodeID(n.NodeID).Do(ctx)
		}))
		if err := chromedp.Run(ctx, chromedp.MouseClickNode(n)); err != nil {
			continue
		}
		time.Sleep(c.ClickDelay)

		// Determine document URL
		var curURL string
		_ = chromedp.Run(ctx, chromedp.Location(&curURL))
		if curURL == "" || visited[curURL] {
			continue
		}

		// Collect content (with backoff)
		var title string
		err := withRetry(backoff, 5, func() error {
			if err := waitVisible(ctx, contentContainerSel, 30*time.Second); err != nil {
				return err
			}
			if err := waitVisible(ctx, titleSel, 10*time.Second); err != nil {
				return err
			}
			var t string
			if err := chromedp.Run(ctx,
				chromedp.Text(titleSel, &t, chromedp.NodeVisible, chromedp.ByQuery),
			); err != nil {
				return err
			}
			title = strings.TrimSpace(t)
			return nil
		})
		if err != nil {
			continue
		}

		if !isAllowedDocURL(curURL) {
			_ = chromedp.Run(ctx, chromedp.Navigate(url))
			_ = waitVisible(ctx, navContainerSel, 15*time.Second)
			continue
		}

		// Fetch innerHTML in a separate context (per requirement)
		var innerHTML string
		_ = withRetry(backoff, 3, func() error {
			ih, e := fetchInnerHTMLWithNewContext(ctx, curURL, 30*time.Second)
			if e != nil {
				return e
			}
			innerHTML = ih
			return nil
		})

		doc := Document{Title: title, URL: curURL, InnerHTML: innerHTML, Content: ""}
		docs = append(docs, doc)
		visited[curURL] = true
		logger.LogParsedDoc(nil, "", doc.Title, doc.URL)

		if c.Limit > 0 && len(visited) >= c.Limit {
			break
		}
	}
	return docs, nil
}
func waitVisible(ctx context.Context, sel string, timeout time.Duration) error {
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return chromedp.Run(c, chromedp.WaitVisible(sel, chromedp.ByQuery))
}

func scrollMenuToEnd(ctx context.Context) error {
	// Scroll repeatedly until no progress
	const js = `(() => {
        const el = document.querySelector("#App > main > div.contents_wrap > div.tree_view_container");
        if (!el) return {ok:false, top:0, height:0};
        const before = el.scrollTop;
        el.scrollBy(0, 1000);
        return {ok:true, top: el.scrollTop, height: el.scrollHeight};
    })()`
	var lastTop int64 = -1
	for i := 0; i < 20; i++ { // safety bound
		var res struct {
			OK     bool  `json:"ok"`
			Top    int64 `json:"top"`
			Height int64 `json:"height"`
		}
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, &res)); err != nil {
			return err
		}
		if !res.OK {
			return errors.New("nav container not found")
		}
		if res.Top == lastTop {
			break
		}
		lastTop = res.Top
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

func withRetry(b *Backoff, maxTries int, fn func() error) error {
	b.Reset()
	var err error
	for i := 0; i < maxTries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(b.Next())
	}
	return err
}

func isAllowedDocURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if !strings.HasSuffix(u.Hostname(), "nexon.com") {
		return false
	}
	if !strings.Contains(u.Path, "/docs") || !strings.Contains(u.Path, "/apiReference") {
		return false
	}
	return true
}

func saveOutput(path, format string, docs []Document) error {
	switch format {
	case "json":
		return SaveJSON(path, docs)
	case "csv":
		return SaveCSV(path, docs)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}
