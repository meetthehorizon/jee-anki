package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListPDFs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some files
	os.WriteFile(filepath.Join(tmpDir, "notes1.pdf"), []byte("dummy pdf content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "notes2.PDF"), []byte("dummy pdf content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not a pdf"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subfolder.pdf"), 0755)

	pdfs, err := ListPDFs(tmpDir)
	if err != nil {
		t.Fatalf("ListPDFs failed: %v", err)
	}

	if len(pdfs) != 2 {
		t.Fatalf("expected 2 PDFs, got %d: %v", len(pdfs), pdfs)
	}

	if pdfs[0] != "notes1.pdf" || pdfs[1] != "notes2.PDF" {
		t.Errorf("unexpected PDF list: %v", pdfs)
	}
}

func TestChunkPDF_SmallFallback(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sample.pdf")
	dummyBytes := []byte("Not a valid PDF header, will trigger fallback")
	os.WriteFile(filePath, dummyBytes, 0644)

	chunks, err := ChunkPDF(filePath, 5)
	if err != nil {
		t.Fatalf("unexpected error on fallback: %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if string(chunks[0].Data) != string(dummyBytes) {
		t.Errorf("expected chunk data to match original bytes")
	}
}
