package crawler

import (
	"strings"
	"unicode"

	"github.com/chromedp/cdproto/cdp"
)

// getAttr returns the value of the given attribute from the node, if present.
// cdp.Node.Attributes is a flat list: name1, value1, name2, value2, ...
func getAttr(n *cdp.Node, name string) (string, bool) {
	if n == nil {
		return "", false
	}
	for i := 0; i+1 < len(n.Attributes); i += 2 {
		if strings.EqualFold(n.Attributes[i], name) {
			return n.Attributes[i+1], true
		}
	}
	return "", false
}

// hasClass reports whether the node's class attribute includes the given class name.
func hasClass(n *cdp.Node, className string) bool {
	v, ok := getAttr(n, "class")
	if !ok || v == "" {
		return false
	}
	for _, c := range strings.Fields(v) {
		if c == className {
			return true
		}
	}
	return false
}

// hasAllClasses reports whether the node has all of the provided class names.
func hasAllClasses(n *cdp.Node, classes ...string) bool {
	for _, c := range classes {
		if !hasClass(n, c) {
			return false
		}
	}
	return true
}

// hasAnyTextInHTML reports whether the given HTML string contains any
// non-whitespace character outside of angle-bracketed tags.
func hasAnyTextInHTML(s string) bool {
	inTag := false
	for _, r := range s {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag && !unicode.IsSpace(r) {
				return true
			}
		}
	}
	return false
}
