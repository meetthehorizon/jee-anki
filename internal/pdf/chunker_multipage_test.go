package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func createTestMultiPagePDF(t *testing.T, targetPath string, numPages int) {
	tmpDir := t.TempDir()
	imgPaths := make([]string, numPages)
	for i := 0; i < numPages; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				img.Set(x, y, color.RGBA{uint8(i * 20), 100, 200, 255})
			}
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatalf("png encode error: %v", err)
		}
		imgPath := filepath.Join(tmpDir, filepath.Clean(string(rune('a'+i))+".png"))
		if err := os.WriteFile(imgPath, buf.Bytes(), 0644); err != nil {
			t.Fatalf("write png error: %v", err)
		}
		imgPaths[i] = imgPath
	}

	conf := model.NewDefaultConfiguration()
	imp, err := api.Import("form:A4, pos:c, sc:1.0", types.POINTS)
	if err != nil {
		t.Fatalf("api.Import config failed: %v", err)
	}

	if err := api.ImportImagesFile(imgPaths, targetPath, imp, conf); err != nil {
		t.Fatalf("api.ImportImagesFile failed: %v", err)
	}
}

func TestChunkPDF_RealMultiPage(t *testing.T) {
	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "multipage.pdf")
	createTestMultiPagePDF(t, pdfPath, 7) // 7 pages

	count, err := GetPageCount(pdfPath)
	if err != nil {
		t.Fatalf("GetPageCount failed: %v", err)
	}
	if count != 7 {
		t.Fatalf("expected 7 pages, got %d", count)
	}

	// Chunk into 3 pages per chunk -> 7 pages = 3 chunks (3 + 3 + 1)
	chunks, err := ChunkPDF(pdfPath, 3)
	if err != nil {
		t.Fatalf("ChunkPDF failed: %v", err)
	}

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	if chunks[0].StartPage != 1 || chunks[0].EndPage != 3 {
		t.Errorf("chunk 0 wrong range: %d-%d", chunks[0].StartPage, chunks[0].EndPage)
	}
	if chunks[1].StartPage != 4 || chunks[1].EndPage != 6 {
		t.Errorf("chunk 1 wrong range: %d-%d", chunks[1].StartPage, chunks[1].EndPage)
	}
	if chunks[2].StartPage != 7 || chunks[2].EndPage != 7 {
		t.Errorf("chunk 2 wrong range: %d-%d", chunks[2].StartPage, chunks[2].EndPage)
	}

	// Verify each chunk is valid PDF bytes by checking its page count
	for i, c := range chunks {
		cPath := filepath.Join(tmpDir, filepath.Clean(string(rune('0'+i))+".pdf"))
		os.WriteFile(cPath, c.Data, 0644)
		cCount, err := GetPageCount(cPath)
		if err != nil {
			t.Fatalf("chunk %d is not a valid PDF: %v", i, err)
		}
		expectedPages := (c.EndPage - c.StartPage) + 1
		if cCount != expectedPages {
			t.Errorf("chunk %d expected %d pages, got %d", i, expectedPages, cCount)
		}
	}
}
