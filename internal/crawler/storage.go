package crawler

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
)

// EncodeJSON encodes the docs slice as a pretty JSON array.
func EncodeJSON(docs []Document) ([]byte, error) {
	return json.MarshalIndent(docs, "", "  ")
}

// SaveJSON writes the docs slice as JSON to the given file path.
func SaveJSON(path string, docs []Document) error {
	data, err := EncodeJSON(docs)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// EncodeCSV encodes the docs slice as CSV with header.
func EncodeCSV(docs []Document) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	// header
	if err := w.Write([]string{"postId", "title", "url", "content"}); err != nil {
		return nil, err
	}
	for _, d := range docs {
		if err := w.Write([]string{d.PostID, d.Title, d.URL, d.Content}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SaveCSV writes the docs slice as CSV to the given file path.
func SaveCSV(path string, docs []Document) error {
	data, err := EncodeCSV(docs)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
