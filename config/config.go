package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ContentExclusionRule defines a rule for excluding content from files.
type ContentExclusionRule struct {
	Type        string `yaml:"type"`              // "delimiters" or "regexp"
	FilePattern string `yaml:"file_pattern"`      // Glob pattern for files (e.g., "*.vue", "vue")
	Start       string `yaml:"start,omitempty"`   // Start delimiter
	End         string `yaml:"end,omitempty"`     // End delimiter
	Pattern     string `yaml:"pattern,omitempty"` // Regex pattern
}

// Config holds the application configuration.
type Config struct {
	Root              string                 `yaml:"root"`
	Include           []string               `yaml:"include"` // Each entry can be "path" or "path:mode" (path, content, both)
	Formats           []string               `yaml:"formats"`
	Output            string                 `yaml:"output"`
	ExcludePatterns   []string               `yaml:"exclude_patterns,omitempty"`
	ContentExclusions []ContentExclusionRule `yaml:"content_exclusions,omitempty"`
}

// ParsedIncludeEntry represents a parsed include item with its mode.
type ParsedIncludeEntry struct {
	Path           string
	Mode           string // "path", "content", or "both"
	IsDirOnlyFiles bool
}

// NewDefaultConfig creates a config with some default values.
func NewDefaultConfig() *Config {
	return &Config{
		Output:            "output.json",
		Include:           []string{},
		Formats:           []string{},
		ExcludePatterns:   []string{},
		ContentExclusions: []ContentExclusionRule{},
	}
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(filePath string) (*Config, error) {
	cfg := NewDefaultConfig() // Start with defaults
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	// Ensure essential fields are initialized if not in YAML
	if cfg.Include == nil {
		cfg.Include = []string{}
	}
	if cfg.Formats == nil {
		cfg.Formats = []string{}
	}
	if cfg.ExcludePatterns == nil {
		cfg.ExcludePatterns = []string{}
	}
	if cfg.ContentExclusions == nil {
		cfg.ContentExclusions = []ContentExclusionRule{}
	}
	return cfg, nil
}

// SaveConfig saves configuration to a YAML file.
func (c *Config) SaveConfig(filePath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Root == "" {
		return errors.New("config error: 'root' directory not specified")
	}
	if _, err := os.Stat(c.Root); os.IsNotExist(err) {
		return errors.New("config error: 'root' directory does not exist: " + c.Root)
	}
	if len(c.Formats) == 0 {
		return errors.New("config error: no file formats specified")
	}
	if c.Output == "" {
		return errors.New("config error: output path not specified")
	}
	// Ensure output directory exists, or can be created
	outputDir := filepath.Dir(c.Output)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return errors.New("config error: could not create output directory: " + outputDir)
		}
	}
	return nil
}
