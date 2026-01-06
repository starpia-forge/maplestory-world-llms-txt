package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"maplestory-world-llms-txt/internal/crawler"
)

var (
	targets = map[string]string{
		/* Reference docs */
		"https://maplestoryworlds-creators.nexon.com/ko/docs/?postId=472": "docs/kr/reference.md",
		"https://maplestoryworlds-creators.nexon.com/en/docs/?postId=472": "docs/en/reference.md",
		/* API docs */
		"https://maplestoryworlds-creators.nexon.com/ko/apiReference/How-to-use-API-Reference": "docs/kr/api.md",
		"https://maplestoryworlds-creators.nexon.com/en/apiReference/How-to-use-API-Reference": "docs/en/api.md",
	}
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

	for targetURL, outFileName := range targets {
		docs, err := c.Run(targetURL)
		if err != nil {
			log.Fatalf("crawler error: %v", err)
		}
		log.Printf("crawled %d documents from %q", len(docs), targetURL)

		htmlFileName := fmt.Sprintf("%s.html", outFileName)
		if err := crawler.SaveDocumentFile(docs, htmlFileName); err != nil {
			log.Fatalf("SaveDocumentFile error: %v", err)
		}
		log.Printf("document to %s", htmlFileName)

		if err := mdream(htmlFileName, outFileName); err != nil {
			log.Fatalf("mdream error: %v", err)
		}
		log.Printf("converted to %s", outFileName)
	}
}
func mdream(inputFileName, outFileName string) error {
	inputFile, err := os.Open(inputFileName)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	cmd := exec.Command("npx", "mdream", "--preset", "minimal")
	cmd.Stdin = inputFile
	cmd.Stdout = io.Writer(outputFile)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
