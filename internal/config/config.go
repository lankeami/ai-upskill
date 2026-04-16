package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type RSSSource struct {
	Name string `mapstructure:"name"`
	URL  string `mapstructure:"url"`
}

type RedditConfig struct {
	Subreddits []string `mapstructure:"subreddits"`
}

type HackerNewsConfig struct {
	MinScore int `mapstructure:"min_score"`
}

type SourcesConfig struct {
	Reddit     RedditConfig     `mapstructure:"reddit"`
	HackerNews HackerNewsConfig `mapstructure:"hackernews"`
	RSS        []RSSSource      `mapstructure:"rss"`
}

type DedupConfig struct {
	LevenshteinThreshold float64 `mapstructure:"levenshtein_threshold"`
}

type OutputConfig struct {
	Dir    string `mapstructure:"dir"`
	Format string `mapstructure:"format"`
}

type EnrichmentConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	Concurrency    int  `mapstructure:"concurrency"`
	TimeoutSeconds int  `mapstructure:"timeout_seconds"`
}

type Config struct {
	Sources    SourcesConfig       `mapstructure:"sources"`
	Keywords   map[string][]string `mapstructure:"keywords"`
	Companies  map[string][]string `mapstructure:"companies"`
	Dedup      DedupConfig         `mapstructure:"dedup"`
	Output     OutputConfig        `mapstructure:"output"`
	Enrichment EnrichmentConfig    `mapstructure:"enrichment"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}
