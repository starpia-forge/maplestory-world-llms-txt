package crawler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/chromedp/chromedp"
)

// collectLeafXPaths builds a list of absolute XPaths for clickable leaf nodes
// under the navigation container after it has been fully expanded.
// Filtering rules:
// - Only div.inactiveDepth elements
// - Exclude elements having class activeParent
// - Exclude elements without any visible (innerText) content
func collectLeafXPaths(ctx context.Context) ([]string, error) {
	js := fmt.Sprintf(`(() => {
	  const container = document.querySelector(%s);
	  if (!container) return [];
	  const nodes = container.querySelectorAll('div.inactiveDepth');
	  const res = [];
	  function hasText(el){ return (el.innerText || '').trim().length > 0; }
	  function xpathFor(el){
	    function idx(e){ let i=1; for(let s=e.previousSibling; s; s=s.previousSibling){ if(s.nodeType===1 && s.nodeName===e.nodeName) i++; } return i; }
	    const seg=[]; for(let e=el; e && e.nodeType===1; e=e.parentNode){ seg.unshift(e.nodeName.toLowerCase()+'['+idx(e)+']'); }
	    return '/'+seg.join('/');
	  }
	  nodes.forEach(el => { if (el.classList.contains('activeParent')) return; if (!hasText(el)) return; res.push(xpathFor(el)); });
	  return res;
	})()`, strconv.Quote(navContainerSel))
	var list []string
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &list)); err != nil {
		return nil, err
	}
	return list, nil
}

// clickByXPathJS finds an element by absolute XPath, scrolls it into view and
// clicks it. Returns whether an element was found and attempted to click.
func clickByXPathJS(ctx context.Context, xp string) (bool, error) {
	js := fmt.Sprintf(`(() => {
	  const xp = %s;
	  const r = document.evaluate(xp, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue;
	  if (r) { try { r.scrollIntoView({block:'center'}); } catch(e){}; try { r.click(); } catch(e){}; return true; }
	  return false;
	})()`, strconv.Quote(xp))
	var ok bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &ok)); err != nil {
		return false, err
	}
	return ok, nil
}
