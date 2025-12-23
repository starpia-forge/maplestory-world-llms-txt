package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"time"

	"maplestory-world-llms-txt/internal/crawler"
)

const (
	ReferenceDocumentKR = "https://maplestoryworlds-creators.nexon.com/ko/docs/?postId=472"
	ReferenceDocumentEN = "https://maplestoryworlds-creators.nexon.com/en/docs/?postId=472"
)

const (
	APIDocumentKR = ""
	APIDocumentEN = ""
)

func main() {
	var (
		head    bool
		delay   time.Duration
		limit   int
		timeout time.Duration
	)

	flag.BoolVar(&head, "headless", true, "run headless Chrome")
	flag.DurationVar(&delay, "delay", 150*time.Millisecond, "delay between clicks")
	flag.IntVar(&limit, "limit", 0, "max number of documents to crawl (0 = no limit)")
	flag.DurationVar(&timeout, "timeout", 120*time.Second, "overall timeout for crawling")
	flag.Parse()

	// Configure default slog logger (text to stderr, Info level)
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	c := crawler.NewCrawler(
		crawler.WithClickDelay(delay),
		crawler.WithLimit(limit),
		crawler.WithOverallTimeout(timeout),
		crawler.WithHeadless(head),
	)

	docs, err := c.Run(ReferenceDocumentKR)
	if err != nil {
		log.Fatalf("crawler error: %v", err)
	}

	if err := crawler.SaveDocumentFile(docs, "docs/kr/reference.raw.txt"); err != nil {
		log.Fatalf("SaveDocumentFile error: %v", err)
	}
}
