package tui

import (
	"fmt"
	"strings"
)

// Version is the current release version of jee-anki.
// It can be overridden at compile time via -ldflags="-X 'github.com/meetthehorizon/jee-anki/internal/tui.Version=v...'"
var Version = "v0.1.2"

// PrintBanner prints the application title header.
func PrintBanner() {
	title := TitleStyle.Render(fmt.Sprintf("jee-anki (%s)", Version))
	subtitle := SubtitleStyle.Render("High-Yield JEE & STEM Flashcard Extractor for Anki")
	fmt.Println(title)
	fmt.Println(subtitle)
}

// PrintAnkiInstructions displays detailed guidance when Anki Desktop is not reachable.
func PrintAnkiInstructions() {
	var sb strings.Builder

	sb.WriteString(WarningStyle.Render("Anki Desktop is not running or AnkiConnect is not installed.\n\n"))
	sb.WriteString("Follow these quick steps to connect Anki:\n")
	sb.WriteString("  1. Launch the " + HighlightStyle.Render("Anki Desktop") + " application.\n")
	sb.WriteString("  2. In Anki's top menu, click " + HighlightStyle.Render("Tools") + " -> " + HighlightStyle.Render("Add-ons") + " -> " + HighlightStyle.Render("Get Add-ons...") + "\n")
	sb.WriteString("  3. Paste the AnkiConnect code: " + CodeStyle.Render("2055492159") + " and click OK.\n")
	sb.WriteString("  4. (Recommended for Chemistry) Also install SMILES to SVG: " + CodeStyle.Render("1819582297") + "\n")
	sb.WriteString("  5. " + WarningStyle.Render("Restart Anki Desktop") + " after installing the add-ons.\n\n")
	sb.WriteString(SubtitleStyle.Render("Once Anki is running, select 'Retry' below."))

	fmt.Println(CardStyle.Render(sb.String()))
}

// PrintNoPDFsFound displays clear guidance when no PDF files are found in the working directory.
func PrintNoPDFsFound(currentDir string) {
	var sb strings.Builder

	sb.WriteString(ErrorStyle.Render("No PDF files found!\n\n"))
	sb.WriteString("Please place your PDF notes in the same folder as this application:\n")
	sb.WriteString("  Folder: " + HighlightStyle.Render(currentDir) + "\n\n")
	sb.WriteString("Supported files: *.pdf (maths, physics, chemistry lecture notes or cheatsheets)")

	fmt.Println(CardStyle.Render(sb.String()))
}
