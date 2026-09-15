package config

import (
	"path/filepath"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	// 1. Loading non-existent file returns default config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected nil error on missing config, got %v", err)
	}
	if cfg.SelectedModel != ModelGemini20Flash {
		t.Errorf("expected default model %s, got %s", ModelGemini20Flash, cfg.SelectedModel)
	}

	// 2. Modify and save
	cfg.APIKey = "AIzaSyTestKey123"
	cfg.SelectedModel = ModelGemini25Flash
	cfg.AddKnownTags([]string{"calculus", "integration-tricks", "new-tag"})

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// 3. Reload and verify
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if loaded.APIKey != "AIzaSyTestKey123" {
		t.Errorf("expected APIKey to match, got %s", loaded.APIKey)
	}
	if loaded.SelectedModel != ModelGemini25Flash {
		t.Errorf("expected SelectedModel %s, got %s", ModelGemini25Flash, loaded.SelectedModel)
	}

	// Verify new tag exists
	found := false
	for _, tag := range loaded.KnownTags {
		if tag == "integration-tricks" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'integration-tricks' in known tags")
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  Maths  ", "maths"},
		{"Differential_Calculus", "differential-calculus"},
		{"Current   Electricity", "current-electricity"},
		{"Organic---Chemistry!!", "organic-chemistry"},
		{"--test--", "test"},
	}

	for _, tt := range tests {
		actual := NormalizeTag(tt.input)
		if actual != tt.expected {
			t.Errorf("NormalizeTag(%q) = %q, expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestMatchKnownTag(t *testing.T) {
	cfg := NewDefaultConfig()
	// Default known tags include "maths"

	// "math" should match "maths"
	matched := cfg.MatchKnownTag("math")
	if matched != "maths" {
		t.Errorf("expected 'math' to match 'maths', got %q", matched)
	}

	// Unknown tag should return normalized form
	matched2 := cfg.MatchKnownTag("rotational motion")
	if matched2 != "rotational-motion" {
		t.Errorf("expected 'rotational-motion', got %q", matched2)
	}
}
