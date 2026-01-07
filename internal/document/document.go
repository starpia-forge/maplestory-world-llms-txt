package document

// Document represents a crawled document item.
type Document struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	InnerHTML string `json:"innerHTML"`
	Content   string `json:"content"`
}
