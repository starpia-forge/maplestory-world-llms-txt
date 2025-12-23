package crawler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveInnerHTMLLines_WritesEachDocumentLine(t *testing.T) {
	docs := []Document{
		{InnerHTML: "<div>first</div>"},
		{InnerHTML: "<p>second</p>"},
		{InnerHTML: "<span>third</span>"},
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "inner.txt")
	if err := SaveDocumentFile(docs, out); err != nil {
		t.Fatalf("SaveDocumentFile error: %v", err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	expected := "<div>first</div>\n<p>second</p>\n<span>third</span>\n"
	if string(b) != expected {
		t.Fatalf("unexpected content.\nwant: %q\n got: %q", expected, string(b))
	}
}
