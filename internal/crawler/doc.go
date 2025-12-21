package crawler

// Document represents a crawled document item.
type Document struct {
	PostID  string `json:"postId"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}
