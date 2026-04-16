package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/config"
	"github.com/jaychinthrajah/ai-upskill/internal/enricher"
	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
	"github.com/jaychinthrajah/ai-upskill/internal/processor"
	"github.com/jaychinthrajah/ai-upskill/internal/renderer"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-report",
	Short: "Generate daily AI news reports",
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the daily AI news report",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("config")
		dateStr, _ := cmd.Flags().GetString("date")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		skipEnrich, _ := cmd.Flags().GetBool("skip-enrich")

		fmt.Println("Loading config...")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		var reportDate time.Time
		if dateStr != "" {
			reportDate, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("parsing date %q: %w", dateStr, err)
			}
		} else {
			reportDate = time.Now().UTC().Truncate(24 * time.Hour)
		}

		since := reportDate.Add(-24 * time.Hour)
		fmt.Printf("Fetching items since %s...\n", since.Format(time.RFC3339))

		items, errs := fetcher.FetchAll(&cfg.Sources, since, "", "")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "fetch warning: %v\n", e)
		}
		fmt.Printf("Fetched %d items\n", len(items))

		if !skipEnrich && cfg.Enrichment.Enabled {
			fmt.Println("Enriching items with metadata...")
			items = enricher.Enrich(items, cfg.Enrichment.Concurrency)
			fmt.Printf("Enrichment complete\n")
		}

		allKeywords := make([]string, 0)
		for _, kws := range cfg.Keywords {
			allKeywords = append(allKeywords, kws...)
		}

		fmt.Println("Classifying items...")
		classified := processor.Classify(items, allKeywords, cfg.Companies)

		fmt.Println("Deduplicating items per company...")
		deduplicated := make(map[string][]processor.DeduplicatedItem)
		for company, companyItems := range classified {
			deduplicated[company] = processor.Deduplicate(companyItems, cfg.Dedup.LevenshteinThreshold)
		}

		sourceNames := collectSourceNames(cfg)
		fmt.Println("Rendering report...")
		report := renderer.RenderMarkdown(deduplicated, reportDate, sourceNames)

		if dryRun {
			fmt.Println("--- DRY RUN OUTPUT ---")
			fmt.Println(report)
			return nil
		}

		outDir := cfg.Output.Dir
		if outDir == "" {
			outDir = "reports"
		}
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("creating output dir: %w", err)
		}

		outPath := filepath.Join(outDir, reportDate.Format("2006-01-02")+".md")
		if err := os.WriteFile(outPath, []byte(report), 0644); err != nil {
			return fmt.Errorf("writing report: %w", err)
		}
		fmt.Printf("Report written to %s\n", outPath)
		return nil
	},
}

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List configured news sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		fmt.Println("Reddit subreddits:")
		for _, sub := range cfg.Sources.Reddit.Subreddits {
			fmt.Printf("  r/%s\n", sub)
		}

		fmt.Printf("\nHacker News:\n")
		fmt.Printf("  min_score: %d\n", cfg.Sources.HackerNews.MinScore)

		fmt.Println("\nRSS feeds:")
		for _, feed := range cfg.Sources.RSS {
			fmt.Printf("  %s: %s\n", feed.Name, feed.URL)
		}

		return nil
	},
}

func collectSourceNames(cfg *config.Config) []string {
	var names []string
	for _, sub := range cfg.Sources.Reddit.Subreddits {
		names = append(names, "r/"+sub)
	}
	names = append(names, "HackerNews")
	for _, feed := range cfg.Sources.RSS {
		names = append(names, feed.Name)
	}
	return names
}

func init() {
	generateCmd.Flags().String("config", "config.yaml", "Path to config file")
	generateCmd.Flags().String("date", "", "Generate report for a specific date (YYYY-MM-DD)")
	generateCmd.Flags().Bool("dry-run", false, "Show report without writing to file")
	generateCmd.Flags().Bool("skip-enrich", false, "Skip metadata enrichment step")

	sourcesCmd.Flags().String("config", "config.yaml", "Path to config file")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(sourcesCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
