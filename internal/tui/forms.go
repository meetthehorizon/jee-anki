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
		var useExisting bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Saved Google Gemini API Key found. Use it?").
					Description("Key is safely loaded from local ./jee-anki.config.json (hidden for screen recording)").
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
				Description("Get your free key at: https://aistudio.google.com/ (Input is masked for recording privacy)").
				Placeholder("Paste API key here...").
				EchoMode(huh.EchoModePassword).
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
		currentModel = config.ModelGeminiFlashLatest
	}

	options := []huh.Option[string]{
		huh.NewOption("gemini-flash-latest (Recommended: Free, Fastest & Always Up-to-Date)", config.ModelGeminiFlashLatest),
		huh.NewOption("gemini-3.8-flash (Latest Flash 3.8)", config.ModelGemini38Flash),
		huh.NewOption("gemini-3.6-flash (Fast & Stable)", config.ModelGemini36Flash),
		huh.NewOption("gemini-3.5-flash (Fast)", config.ModelGemini35Flash),
		huh.NewOption("gemini-pro-latest (Pro Reasoning)", config.ModelGeminiProLatest),
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

	options := make([]huh.Option[string], len(pdfs))
	for i, p := range pdfs {
		options[i] = huh.NewOption(p, p)
	}

	selected := pdfs[0]
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select PDF Note to Process").
				Description("Choose which PDF note to parse into flashcards").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	return selected, nil
}

// PromptConfirmStart asks the user to confirm before starting extraction.
func PromptConfirmStart(fileName string, totalPages, chunkCount int, modelName string) (bool, error) {
	title := fmt.Sprintf("Ready to extract flashcards from '%s'?", fileName)
	desc := fmt.Sprintf("Document: %d page(s) | Chunks: %d | Model: %s", totalPages, chunkCount, modelName)

	var proceed bool = true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(desc).
				Affirmative("Start Extraction").
				Negative("Cancel").
				Value(&proceed),
		),
	)

	if err := form.Run(); err != nil {
		return false, err
	}

	return proceed, nil
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
