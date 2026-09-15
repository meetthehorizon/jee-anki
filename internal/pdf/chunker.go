package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

const (
	DefaultPagesPerChunk = 5
)

// Chunk represents a segmented slice of a PDF document.
type Chunk struct {
	Index      int    // 1-based index of this chunk
	Total      int    // Total number of chunks
	StartPage  int    // Starting page number (1-based)
	EndPage    int    // Ending page number (1-based)
	TotalPages int    // Total pages in the original document
	Data       []byte // Raw PDF bytes for this chunk
}

// ListPDFs scans the specified directory for PDF files.
func ListPDFs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var pdfs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".pdf" {
			pdfs = append(pdfs, entry.Name())
		}
	}

	sort.Strings(pdfs)
	return pdfs, nil
}

// GetPageCount returns the number of pages in a PDF file.
func GetPageCount(filePath string) (int, error) {
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer f.Close()

	count, err := api.PageCount(f, conf)
	if err != nil {
		return 0, fmt.Errorf("failed to read PDF page count: %w", err)
	}
	return count, nil
}

// ChunkPDF splits a PDF file into chunks of pagesPerChunk.
// If the PDF has fewer or equal pages to pagesPerChunk, it returns the whole PDF as a single chunk.
func ChunkPDF(filePath string, pagesPerChunk int) ([]Chunk, error) {
	if pagesPerChunk <= 0 {
		pagesPerChunk = DefaultPagesPerChunk
	}

	rawBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF file: %w", err)
	}

	totalPages, err := GetPageCount(filePath)
	if err != nil {
		// If page count determination fails, fallback to passing the whole file as a single chunk
		return []Chunk{
			{
				Index:      1,
				Total:      1,
				StartPage:  1,
				EndPage:    1,
				TotalPages: 1,
				Data:       rawBytes,
			},
		}, nil
	}

	// If file has fewer or equal pages to the chunk size, send the raw bytes directly
	if totalPages <= pagesPerChunk {
		return []Chunk{
			{
				Index:      1,
				Total:      1,
				StartPage:  1,
				EndPage:    totalPages,
				TotalPages: totalPages,
				Data:       rawBytes,
			},
		}, nil
	}

	// Calculate chunk ranges
	type pageRange struct {
		start int
		end   int
	}
	var ranges []pageRange
	for start := 1; start <= totalPages; start += pagesPerChunk {
		end := start + pagesPerChunk - 1
		if end > totalPages {
			end = totalPages
		}
		ranges = append(ranges, pageRange{start: start, end: end})
	}

	totalChunks := len(ranges)
	chunks := make([]Chunk, 0, totalChunks)

	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	for i, r := range ranges {
		var chunkBuf bytes.Buffer
		pageSelection := []string{fmt.Sprintf("%d-%d", r.start, r.end)}

		reader := bytes.NewReader(rawBytes)
		err := api.Trim(reader, &chunkBuf, pageSelection, conf)
		if err != nil {
			return nil, fmt.Errorf("failed to extract pages %d-%d: %w", r.start, r.end, err)
		}

		chunks = append(chunks, Chunk{
			Index:      i + 1,
			Total:      totalChunks,
			StartPage:  r.start,
			EndPage:    r.end,
			TotalPages: totalPages,
			Data:       chunkBuf.Bytes(),
		})
	}

	return chunks, nil
}
