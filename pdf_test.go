package pdf_test

import (
	"crypto/sha1"
	"encoding/base64"
	"image"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/richardwilkes/pdf"
)

func TestPDF(t *testing.T) {
	// Load the data we are going to use
	data, err := os.ReadFile("testfiles/GLAIVE_Mini_v2_3_for_GURPS_4e.pdf")
	if err != nil {
		t.Fatal(err)
	}

	// Parse the data as a PDF document
	var doc *pdf.Document
	doc, err = pdf.New(data, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Release()

	// Count the pages
	if pageCount := doc.PageCount(); pageCount != 2 {
		t.Fatalf("expected 2 pages, got %d", pageCount)
	}

	// Load the table of contents
	toc := doc.TableOfContents(100)
	if len(toc) != 66 {
		t.Fatalf("expected 66 TOC entries, got %d", len(toc))
	}

	// This particular PDF has ridiculously long TOC headings, so just spot-check a few
	checkTOCEntry(t, toc, 0, "GLAIVE Mini (GMi) ", 0, 69, 163)
	checkTOCEntry(t, toc, 12, "Semibalanced: A new ", 0, 81, 680)
	checkTOCEntry(t, toc, 60, "What's that odd ", 1, 446, 691)

	// Render the first page
	var page *pdf.RenderedPage
	page, err = doc.RenderPage(0, 100, 20, "GURPS")
	if err != nil {
		t.Fatal(err)
	}

	// Verify the search hits match expectations
	if len(page.SearchHits) != 9 {
		t.Fatalf("expected 9 search hits, got %d", len(page.SearchHits))
	}
	for i, one := range []image.Rectangle{
		image.Rect(152, 180, 193, 194),
		image.Rect(162, 208, 204, 221),
		image.Rect(265, 684, 306, 698),
		image.Rect(484, 311, 526, 324),
		image.Rect(670, 384, 712, 398),
		image.Rect(600, 567, 660, 585),
		image.Rect(180, 1131, 226, 1145),
		image.Rect(69, 126, 125, 143),
		image.Rect(425, 86, 460, 97),
	} {
		if page.SearchHits[i] != one {
			t.Errorf("search hit rect %d doesn't match, expected %v, got %v", i, one, page.SearchHits[i])
		}
	}

	// Verify the links match expectations
	if len(page.Links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(page.Links))
	}
	for i, one := range []pdf.PageLink{
		{
			PageNumber: -1,
			URI:        "http://www.gamesdiner.com/glaive_mini",
			Bounds:     image.Rect(69, 163, 149, 180),
		},
		{
			PageNumber: -1,
			URI:        "http://www.gamesdiner.com",
			Bounds:     image.Rect(472, 1128, 604, 1145),
		},
	} {
		if *page.Links[i] != one {
			t.Errorf("link %d doesn't match, expected %#v, got %#v", i, one, *page.Links[i])
		}
	}

	// Verify the image
	if page.Image == nil {
		t.Fatal("expected image data, got nil")
	}
	if page.Image.Stride != 3308 {
		t.Errorf("expected an image stride of 3308, got %d", page.Image.Stride)
	}
	expectedBounds := image.Rect(0, 0, 827, 1170)
	if page.Image.Rect != expectedBounds {
		t.Errorf("expected an image bounds of %v, got %v", expectedBounds, page.Image.Rect)
	}
	sum := sha1.Sum(page.Image.Pix)
	var expected string
	if runtime.GOOS != "windows" {
		expected = "/zVQwB1j73JopAxP9db/qig1mU4"
	} else {
		expected = "1kyvOn48i4kVlKIfNyBoqb51uEQ" // Windows has a different value due to subtle differences in font display
	}
	if base64.RawStdEncoding.EncodeToString(sum[:]) != expected {
		t.Error("rendered image doesn't match expectation")
	}
}

func checkTOCEntry(t *testing.T, toc []*pdf.TOCEntry, index int, prefix string, pageNumber, pageX, pageY int) {
	t.Helper()
	if !strings.HasPrefix(toc[index].Title, prefix) {
		t.Errorf("TOC entry %d's Title does not start with %q, instead is %q", index, prefix, toc[index].Title)
	}
	if toc[index].PageNumber != pageNumber {
		t.Errorf("TOC entry %d's PageNumber is not %d, got %d", index, pageNumber, toc[index].PageNumber)
	}
	if toc[index].PageX != pageX {
		t.Errorf("TOC entry %d's PageX is not %d, got %d", index, pageX, toc[index].PageX)
	}
	if toc[index].PageY != pageY {
		t.Errorf("TOC entry %d's PageY is not %d, got %d", index, pageY, toc[index].PageY)
	}
	if toc[index].Children != nil {
		t.Errorf("TOC entry %d's Children expected to be nil", index)
	}
}
