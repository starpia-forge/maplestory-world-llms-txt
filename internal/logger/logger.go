package logger

import "log/slog"

// LogParsedDoc emits an info-level structured log for a newly parsed document.
// It logs message "parsed_doc" with attrs: title, url.
// If l is nil, slog.Default() is used.
func LogParsedDoc(l *slog.Logger, title, url string) {
	if l == nil {
		l = slog.Default()
	}
	l.Info("parsed_doc",
		slog.String("title", title),
		slog.String("url", url),
	)
}
