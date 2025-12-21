package crawler

import (
	"testing"

	"github.com/chromedp/cdproto/cdp"
)

func Test_getAttr(t *testing.T) {
	n := &cdp.Node{Attributes: []string{"id", "foo", "class", "a b c"}}
	if v, ok := getAttr(n, "id"); !ok || v != "foo" {
		t.Fatalf("expected id=foo, got ok=%v v=%q", ok, v)
	}
	if v, ok := getAttr(n, "class"); !ok || v != "a b c" {
		t.Fatalf("expected class=\"a b c\", got ok=%v v=%q", ok, v)
	}
	if _, ok := getAttr(n, "href"); ok {
		t.Fatalf("expected no href attr")
	}
}

func Test_hasClass(t *testing.T) {
	n := &cdp.Node{Attributes: []string{"class", "inactiveDot isHavingChildren"}}
	if !hasClass(n, "inactiveDot") {
		t.Fatalf("inactiveDot should be present")
	}
	if !hasClass(n, "isHavingChildren") {
		t.Fatalf("isHavingChildren should be present")
	}
	if hasClass(n, "isHavingChildrenAndOpen") {
		t.Fatalf("isHavingChildrenAndOpen should NOT be present")
	}
}

func Test_hasAllClasses(t *testing.T) {
	n := &cdp.Node{Attributes: []string{"class", "inactiveDot isHavingChildren"}}
	if !hasAllClasses(n, "inactiveDot", "isHavingChildren") {
		t.Fatalf("expected hasAllClasses to be true")
	}
	if hasAllClasses(n, "inactiveDot", "isHavingChildren", "isHavingChildrenAndOpen") {
		t.Fatalf("expected hasAllClasses to be false due to missing class")
	}
}

func Test_hasAnyTextInHTML(t *testing.T) {
	cases := []struct {
		html string
		want bool
	}{
		{"<div></div>", false},
		{"<div> </div>", false},
		{"<div>text</div>", true},
		{"<div><span> t </span></div>", true},
		{"<div><span></span>\n\t</div>", false},
		{"<div>한</div>", true},
	}
	for i, c := range cases {
		if got := hasAnyTextInHTML(c.html); got != c.want {
			t.Fatalf("case %d: expected %v, got %v for %q", i, c.want, got, c.html)
		}
	}
}
