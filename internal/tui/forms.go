package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/meetthehorizon/jee-anki/internal/config"
)

// PromptRetryAnki asks the user whether to retry connecting to Anki Desktop.
func PromptRetryAnki() (bool, error) {
	var retry bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Retry connecting to Anki?").
				Affirmative("Retry").
				Negative("Exit").
				Value(&retry),
		),
	)

	err := form.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}
		return false, err
	}

	return retry, nil
}

// PromptAPIKey prompts the user to enter their Gemini API key or confirm existing.
func PromptAPIKey(currentKey string) (string, error) {
	if currentKey != "" {
		maskedKey := currentKey
		if len(currentKey) > 8 {
			maskedKey = currentKey[:4] + "..." + currentKey[len(currentKey)-4:]
		}

		var useExisting bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Saved API Key found (%s). Use it?", maskedKey)).
					Affirmative("Yes, use saved key").
					Negative("No, enter a new key").
					Value(&useExisting),
			),
		)

		if err := confirmForm.Run(); err != nil {
			return "", err
		}

		if useExisting {
			return currentKey, nil
		}
	}

	var newKey string
	inputForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter your Google Gemini API Key").
				Description("Get your 100% free key at: https://aistudio.google.com/").
				Placeholder("AIzaSy...").
				Value(&newKey).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("API key cannot be empty")
					}
					if len(strings.TrimSpace(s)) < 15 {
						return errors.New("API key appears too short (must be valid Google API key)")
					}
					return nil
				}),
		),
	)

	if err := inputForm.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(newKey), nil
}

// PromptModel prompts the user to pick a model or specify a custom one.
func PromptModel(currentModel string) (string, error) {
	const customOption = "__custom__"

	if currentModel == "" {
		currentModel = config.ModelGemini20Flash
	}

	options := []huh.Option[string]{
		huh.NewOption("gemini-2.0-flash (Recommended: Free, Fast & Accurate)", config.ModelGemini20Flash),
		huh.NewOption("gemini-2.0 (Standard)", config.ModelGemini20),
		huh.NewOption("gemini-2.5-flash (Latest Flash Preview)", config.ModelGemini25Flash),
		huh.NewOption("gemini-2.5 (Pro reasoning)", config.ModelGemini25),
		huh.NewOption("Enter custom model name...", customOption),
	}

	// If currentModel is custom, make sure it's valid
	selected := currentModel
	isKnown := false
	for _, opt := range options {
		if opt.Value == currentModel {
			isKnown = true
			break
		}
	}
	if !isKnown {
		selected = customOption
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Gemini Model").
				Description("Flash models are fast and 100% free; Pro models offer deeper reasoning.").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	if selected == customOption {
		var customName string
		customForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter Custom Gemini Model Name").
					Description("e.g. gemini-3.0-flash or gemini-exp-1206").
					Placeholder("gemini-2.0-flash").
					Value(&customName).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("model name cannot be empty")
						}
						return nil
					}),
			),
		)

		if err := customForm.Run(); err != nil {
			return "", err
		}
		return strings.TrimSpace(customName), nil
	}

	return selected, nil
}

// PromptSelectPDF lets the user pick one PDF file from the list.
func PromptSelectPDF(pdfs []string) (string, error) {
	if len(pdfs) == 0 {
		return "", errors.New("no PDFs available")
	}

	if len(pdfs) == 1 {
		fmt.Printf("%s Auto-selected only PDF found: %s\n\n", SuccessStyle.Render("✓"), HighlightStyle.Render(pdfs[0]))
		return pdfs[0], nil
	}

	options := make([]huh.Option[string], len(pdfs))
	for i, p := range pdfs {
		options[i] = huh.NewOption(p, p)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select PDF Note to Process").
				Description("Use arrow keys or mouse to pick one file").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	return selected, nil
}

// PromptConfirmTags displays AI-suggested tags and allows the user to accept or edit them.
func PromptConfirmTags(suggestedTags []string) ([]string, error) {
	cleanSuggestions := make([]string, 0, len(suggestedTags))
	for _, t := range suggestedTags {
		norm := config.NormalizeTag(t)
		if norm != "" {
			cleanSuggestions = append(cleanSuggestions, norm)
		}
	}

	initialValue := strings.Join(cleanSuggestions, ", ")
	var userInput string = initialValue

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Review Flashcard Tags").
				Description("AI-suggested tags based on note content. Press Enter to accept, or edit.").
				Value(&userInput).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("at least one tag is required")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	rawParts := strings.Split(userInput, ",")
	tags := make([]string, 0, len(rawParts))
	seen := make(map[string]bool)

	for _, part := range rawParts {
		norm := config.NormalizeTag(part)
		if norm != "" && !seen[norm] {
			seen[norm] = true
			tags = append(tags, norm)
		}
	}

	if len(tags) == 0 {
		return []string{"jee", "notes"}, nil
	}

	return tags, nil
}
