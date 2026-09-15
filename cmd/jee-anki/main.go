package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/meetthehorizon/jee-anki/internal/anki"
	"github.com/meetthehorizon/jee-anki/internal/config"
	"github.com/meetthehorizon/jee-anki/internal/gemini"
	"github.com/meetthehorizon/jee-anki/internal/pdf"
	"github.com/meetthehorizon/jee-anki/internal/tui"
)

func waitForExit() {
	fmt.Println()
	fmt.Println(tui.HighlightStyle.Render("Press Enter to exit..."))
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadBytes('\n')
}

func main() {
	// Guaranteed pause so Windows console never closes abruptly on completion or error
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\n%s Unexpected runtime error: %v\n", tui.ErrorStyle.Render("[!]"), r)
		}
		waitForExit()
	}()

	tui.PrintBanner()

	ctx := context.Background()
	ankiClient := anki.NewClient()

	// 1. Verify AnkiConnect connectivity
	for {
		err := ankiClient.Ping(ctx)
		if err == nil {
			fmt.Printf("%s Connected to Anki Desktop (AnkiConnect active)\n\n", tui.SuccessStyle.Render("✓"))
			break
		}

		tui.PrintAnkiInstructions()
		retry, promptErr := tui.PromptRetryAnki()
		if promptErr != nil || !retry {
			fmt.Println(tui.WarningStyle.Render("\nExiting without making changes."))
			return
		}
	}

	// 2. Load / Configure Settings
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("%s Failed to load config: %v\n", tui.WarningStyle.Render("[!]"), err)
		cfg = config.NewDefaultConfig()
	}

	apiKey, err := tui.PromptAPIKey(cfg.APIKey)
	if err != nil {
		fmt.Println(tui.WarningStyle.Render("\nOperation canceled."))
		return
	}
	cfg.APIKey = apiKey

	model, err := tui.PromptModel(cfg.SelectedModel)
	if err != nil {
		fmt.Println(tui.WarningStyle.Render("\nOperation canceled."))
		return
	}
	cfg.SelectedModel = model

	// Save updated credentials/settings to local config
	if err := cfg.Save(); err != nil {
		fmt.Printf("%s Warning: failed to save local config: %v\n", tui.WarningStyle.Render("[!]"), err)
	}

	// 3. Find and select PDF in current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	pdfFiles, err := pdf.ListPDFs(currentDir)
	if err != nil {
		fmt.Printf("%s Error reading directory: %v\n", tui.ErrorStyle.Render("[!]"), err)
		return
	}

	if len(pdfFiles) == 0 {
		tui.PrintNoPDFsFound(currentDir)
		return
	}

	selectedPDF, err := tui.PromptSelectPDF(pdfFiles)
	if err != nil {
		fmt.Println(tui.WarningStyle.Render("\nOperation canceled."))
		return
	}

	// 4. Chunk & Extract Cards
	fmt.Printf("%s Analyzing %s...\n", tui.HighlightStyle.Render("►"), tui.HighlightStyle.Render(selectedPDF))
	chunks, err := pdf.ChunkPDF(filepath.Join(currentDir, selectedPDF), pdf.DefaultPagesPerChunk)
	if err != nil {
		fmt.Printf("%s Failed to read PDF pages: %v\n", tui.ErrorStyle.Render("[!]"), err)
		return
	}

	totalPages := chunks[0].TotalPages
	fmt.Printf("%s Document has %d page(s). Processing in %d chunk(s) with model %s...\n\n",
		tui.SuccessStyle.Render("✓"),
		totalPages,
		len(chunks),
		tui.CodeStyle.Render(cfg.SelectedModel),
	)

	geminiClient := gemini.NewClient(cfg.APIKey, cfg.SelectedModel)

	var allCards []anki.Note
	suggestedTagSet := make(map[string]bool)
	var rawSuggestedTags []string

	for _, chunk := range chunks {
		fmt.Printf("  Processing chunk %d of %d (Pages %d-%d)... ",
			chunk.Index, chunk.Total, chunk.StartPage, chunk.EndPage)

		res, err := geminiClient.ExtractFromPDFChunk(ctx, chunk.Data)
		if err != nil {
			fmt.Printf("%s\n", tui.ErrorStyle.Render("FAILED"))
			fmt.Printf("\n%s %v\n", tui.ErrorStyle.Render("[Extraction Error]"), err)
			if len(allCards) > 0 {
				fmt.Printf("%s Continuing with %d cards extracted from earlier chunks...\n\n",
					tui.WarningStyle.Render("!"), len(allCards))
				break
			}
			return
		}

		fmt.Printf("%s (found %d cards)\n",
			tui.SuccessStyle.Render("DONE"),
			len(res.Cards),
		)

		allCards = append(allCards, res.Cards...)
		for _, tag := range res.SuggestedTags {
			matched := cfg.MatchKnownTag(tag)
			if matched != "" && !suggestedTagSet[matched] {
				suggestedTagSet[matched] = true
				rawSuggestedTags = append(rawSuggestedTags, matched)
			}
		}
	}

	if len(allCards) == 0 {
		fmt.Printf("\n%s No flashcards could be extracted from this PDF.\n", tui.WarningStyle.Render("[!]"))
		return
	}

	fmt.Printf("\n%s Extracted a total of %d flashcard(s)!\n\n",
		tui.SuccessStyle.Render("★"), len(allCards))

	// 5. Review & Confirm Tags
	// If no tags were suggested by the model, provide default based on filename
	if len(rawSuggestedTags) == 0 {
		base := strings.TrimSuffix(filepath.Base(selectedPDF), filepath.Ext(selectedPDF))
		rawSuggestedTags = []string{config.NormalizeTag(base)}
	}

	confirmedTags, err := tui.PromptConfirmTags(rawSuggestedTags)
	if err != nil {
		fmt.Println(tui.WarningStyle.Render("\nOperation canceled."))
		return
	}

	// Update known tags in local config
	cfg.AddKnownTags(confirmedTags)
	_ = cfg.Save()

	// 6. Deliver to Anki
	fmt.Printf("\n%s Pushing flashcards to Anki deck '%s'...\n",
		tui.HighlightStyle.Render("►"), anki.DefaultDeckName)

	if err := ankiClient.EnsureDeck(ctx, anki.DefaultDeckName); err != nil {
		fmt.Printf("%s Failed to create or verify deck '%s': %v\n",
			tui.ErrorStyle.Render("[!]"), anki.DefaultDeckName, err)
		return
	}

	modelName, err := ankiClient.GetBasicModelName(ctx)
	if err != nil {
		modelName = "Basic"
	}

	result, err := ankiClient.AddNotes(ctx, anki.DefaultDeckName, modelName, allCards, confirmedTags)
	if err != nil {
		fmt.Printf("%s Error adding notes to Anki: %v\n", tui.ErrorStyle.Render("[!]"), err)
		return
	}

	// 7. Completion Summary
	fmt.Println()
	fmt.Println(tui.SuccessStyle.Render("=================================================="))
	fmt.Println(tui.SuccessStyle.Render("                 SUCCESS!                         "))
	fmt.Println(tui.SuccessStyle.Render("=================================================="))
	fmt.Printf("• Target Deck:             %s\n", tui.HighlightStyle.Render(anki.DefaultDeckName))
	fmt.Printf("• Cards Successfully Added: %s\n", tui.SuccessStyle.Render(fmt.Sprintf("%d", result.Added)))
	if result.Duplicate > 0 {
		fmt.Printf("• Duplicate Cards Skipped:  %s\n", tui.WarningStyle.Render(fmt.Sprintf("%d", result.Duplicate)))
	}
	fmt.Printf("• Tags Applied:            %s\n\n", tui.HighlightStyle.Render(strings.Join(confirmedTags, ", ")))
	fmt.Println("All cards are ready for your review in Anki Desktop under the 'Inbox' deck.")
	fmt.Println("You can inspect them, edit them, and move them into your main subject decks.")
}
