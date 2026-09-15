package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultConfigFileName = "jee-anki.config.json"

	ModelGemini20Flash = "gemini-2.0-flash"
	ModelGemini20      = "gemini-2.0"
	ModelGemini25Flash = "gemini-2.5-flash"
	ModelGemini25      = "gemini-2.5"
)

// SupportedModels lists the standard choices shown in the UI.
var SupportedModels = []string{
	ModelGemini20Flash,
	ModelGemini20,
	ModelGemini25Flash,
	ModelGemini25,
}

// Config represents persistent application settings saved in the local folder.
type Config struct {
	APIKey        string   `json:"api_key"`
	SelectedModel string   `json:"selected_model"`
	KnownTags     []string `json:"known_tags"`
}

// NewDefaultConfig returns a config populated with safe defaults.
func NewDefaultConfig() *Config {
	return &Config{
		APIKey:        "",
		SelectedModel: ModelGemini20Flash,
		KnownTags: []string{
			"physics",
			"chemistry",
			"maths",
			"organic-chemistry",
			"inorganic-chemistry",
			"physical-chemistry",
			"calculus",
			"algebra",
			"mechanics",
			"electrodynamics",
		},
	}
}

// DefaultPath returns the path to jee-anki.config.json in the current working directory.
func DefaultPath() string {
	return filepath.Clean(DefaultConfigFileName)
}

// Load loads the configuration from disk, or returns default config if the file doesn't exist.
func Load(optionalPath ...string) (*Config, error) {
	targetPath := DefaultPath()
	if len(optionalPath) > 0 && optionalPath[0] != "" {
		targetPath = optionalPath[0]
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewDefaultConfig(), nil
		}
		return nil, err
	}

	cfg := NewDefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.SelectedModel == "" {
		cfg.SelectedModel = ModelGemini20Flash
	}

	return cfg, nil
}

// Save writes the configuration to disk in formatted JSON.
func (c *Config) Save(optionalPath ...string) error {
	targetPath := DefaultPath()
	if len(optionalPath) > 0 && optionalPath[0] != "" {
		targetPath = optionalPath[0]
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(targetPath, data, 0600)
}

// NormalizeTag cleans up a tag string to standard kebab-case format.
func NormalizeTag(tag string) string {
	tag = strings.TrimSpace(strings.ToLower(tag))
	// Replace spaces and underscores with hyphens
	tag = strings.ReplaceAll(tag, " ", "-")
	tag = strings.ReplaceAll(tag, "_", "-")

	// Filter allowed characters: alphanumeric and hyphen
	var b strings.Builder
	for _, r := range tag {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	clean := b.String()
	// Collapse consecutive hyphens
	for strings.Contains(clean, "--") {
		clean = strings.ReplaceAll(clean, "--", "-")
	}
	clean = strings.Trim(clean, "-")
	return clean
}

// MatchKnownTag checks if a candidate tag matches any known tag (exact or singular/plural).
// Returns the existing known tag if matched, or the normalized candidate.
func (c *Config) MatchKnownTag(candidate string) string {
	normCandidate := NormalizeTag(candidate)
	if normCandidate == "" {
		return ""
	}

	for _, existing := range c.KnownTags {
		normExisting := NormalizeTag(existing)
		if normExisting == normCandidate {
			return existing
		}
		// Handle common singular/plural matching like math <-> maths
		if normCandidate+"s" == normExisting || normExisting+"s" == normCandidate {
			return existing
		}
	}

	return normCandidate
}

// AddKnownTags adds new tags to the list of known tags without duplicates.
func (c *Config) AddKnownTags(tags []string) {
	tagSet := make(map[string]bool)
	for _, t := range c.KnownTags {
		cleaned := NormalizeTag(t)
		if cleaned != "" {
			tagSet[cleaned] = true
		}
	}

	for _, t := range tags {
		cleaned := NormalizeTag(t)
		if cleaned != "" && !tagSet[cleaned] {
			tagSet[cleaned] = true
			c.KnownTags = append(c.KnownTags, cleaned)
		}
	}
}
