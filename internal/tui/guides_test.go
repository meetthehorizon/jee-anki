package tui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = stdout
	return <-outC
}

func TestPrintBanner(t *testing.T) {
	out := captureOutput(PrintBanner)
	if !strings.Contains(out, "jee-anki") {
		t.Errorf("expected banner to contain 'jee-anki', got: %s", out)
	}
}

func TestPrintAnkiInstructions(t *testing.T) {
	out := captureOutput(PrintAnkiInstructions)
	if !strings.Contains(out, "2055492159") {
		t.Errorf("expected instructions to contain AnkiConnect code 2055492159, got: %s", out)
	}
	if !strings.Contains(out, "1819582297") {
		t.Errorf("expected instructions to contain SMILES code 1819582297, got: %s", out)
	}
}

func TestPrintNoPDFsFound(t *testing.T) {
	out := captureOutput(func() {
		PrintNoPDFsFound("/dummy/folder")
	})
	if !strings.Contains(out, "No PDF files found") {
		t.Errorf("expected 'No PDF files found', got: %s", out)
	}
	if !strings.Contains(out, "/dummy/folder") {
		t.Errorf("expected folder to be printed, got: %s", out)
	}
}
