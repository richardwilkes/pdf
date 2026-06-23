package pdf_test

import (
	"errors"
	"image"
	"os"
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

	// A negative page number must be rejected rather than crashing in MuPDF
	if _, err = doc.RenderPage(-1, 100, 20, ""); !errors.Is(err, pdf.ErrInvalidPageNumber) {
		t.Errorf("expected ErrInvalidPageNumber for a negative page, got %v", err)
	}
	if _, err = doc.RenderPageForSize(-1, 800, 800, 20, ""); !errors.Is(err, pdf.ErrInvalidPageNumber) {
		t.Errorf("expected ErrInvalidPageNumber for a negative page, got %v", err)
	}

	// A search with maxHits <= 0 must not panic and must yield no hits
	page, err = doc.RenderPage(0, 100, 0, "GURPS")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.SearchHits) != 0 {
		t.Errorf("expected 0 search hits with maxHits of 0, got %d", len(page.SearchHits))
	}
}

func TestMalformedPDF(t *testing.T) {
	// A buffer with a valid %PDF prefix but garbage contents passes the prefix check and then causes MuPDF to throw
	// while opening the document. This must surface as ErrUnableToOpenPDF rather than crashing the process.
	if _, err := pdf.New([]byte("%PDF-1.7\nnot a real pdf"), 0); !errors.Is(err, pdf.ErrUnableToOpenPDF) {
		t.Fatalf("expected ErrUnableToOpenPDF for a malformed document, got %v", err)
	}
}

func TestUseAfterRelease(t *testing.T) {
	data, err := os.ReadFile("testfiles/GLAIVE_Mini_v2_3_for_GURPS_4e.pdf")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := pdf.New(data, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Releasing and then calling methods must not crash; it must return safe zero values / ErrDocumentReleased.
	doc.Release()

	// Calling Release again must be a safe no-op.
	doc.Release()

	if got := doc.PageCount(); got != 0 {
		t.Errorf("expected PageCount of 0 after release, got %d", got)
	}
	if got := doc.RequiresAuthentication(); got {
		t.Errorf("expected RequiresAuthentication of false after release, got %v", got)
	}
	if got := doc.Authenticate(""); got != 0 {
		t.Errorf("expected Authenticate status of 0 after release, got %v", got)
	}
	if got := doc.TableOfContents(100); got != nil {
		t.Errorf("expected nil TableOfContents after release, got %v", got)
	}
	if _, err = doc.RenderPage(0, 100, 20, ""); !errors.Is(err, pdf.ErrDocumentReleased) {
		t.Errorf("expected ErrDocumentReleased from RenderPage after release, got %v", err)
	}
	if _, err = doc.RenderPageForSize(0, 800, 800, 20, ""); !errors.Is(err, pdf.ErrDocumentReleased) {
		t.Errorf("expected ErrDocumentReleased from RenderPageForSize after release, got %v", err)
	}
}

func TestRenderPageForSizeLimits(t *testing.T) {
	data, err := os.ReadFile("testfiles/GLAIVE_Mini_v2_3_for_GURPS_4e.pdf")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := pdf.New(data, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Release()

	// A normal request renders successfully and fits within the requested box.
	page, err := doc.RenderPageForSize(0, 800, 800, 0, "")
	if err != nil {
		t.Fatalf("unexpected error rendering for size: %v", err)
	}
	if page.Image == nil {
		t.Fatal("expected image data, got nil")
	}
	if b := page.Image.Bounds(); b.Dx() <= 0 || b.Dy() <= 0 || b.Dx() > 800 || b.Dy() > 800 {
		t.Errorf("rendered image %v does not fit within 800x800", b)
	}

	// A non-positive target size must be rejected up front with ErrInvalidPageSize.
	for _, sz := range []struct{ w, h int }{{0, 800}, {800, 0}, {-1, 800}, {800, -1}} {
		if _, err = doc.RenderPageForSize(0, sz.w, sz.h, 0, ""); !errors.Is(err, pdf.ErrInvalidPageSize) {
			t.Errorf("expected ErrInvalidPageSize for target size %dx%d, got %v", sz.w, sz.h, err)
		}
	}

	// A request whose output would exceed OverallMaxPixels must be rejected with ErrImageTooLarge rather than
	// attempting a huge allocation. Both render paths enforce the same limit and report the same sentinel.
	defer func(prev int) { pdf.OverallMaxPixels = prev }(pdf.OverallMaxPixels)
	pdf.OverallMaxPixels = 100
	if _, err = doc.RenderPageForSize(0, 800, 800, 0, ""); !errors.Is(err, pdf.ErrImageTooLarge) {
		t.Errorf("expected ErrImageTooLarge from RenderPageForSize when exceeding OverallMaxPixels, got %v", err)
	}
	if _, err = doc.RenderPage(0, 100, 0, ""); !errors.Is(err, pdf.ErrImageTooLarge) {
		t.Errorf("expected ErrImageTooLarge from RenderPage when exceeding OverallMaxPixels, got %v", err)
	}
}

// internalLinkPDF is a minimal two-page document with two internal links on page 0, both targeting the second page:
// one via an explicit /XYZ destination ([4 0 R /XYZ 30 150 0]) and one via a named destination
// (/A /GoTo /D (Chapter2), which resolves to a /Fit destination with no point). No xref is supplied (startxref 0) so
// MuPDF rebuilds it; only the link resolution matters here.
const internalLinkPDF = `%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R /Names << /Dests 6 0 R >> >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R 4 0 R] /Count 2 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Annots [5 0 R 7 0 R] >>
endobj
4 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] >>
endobj
5 0 obj
<< /Type /Annot /Subtype /Link /Rect [10 10 90 30] /Border [0 0 0] /Dest [4 0 R /XYZ 30 150 0] >>
endobj
6 0 obj
<< /Names [(Chapter2) [4 0 R /Fit]] >>
endobj
7 0 obj
<< /Type /Annot /Subtype /Link /Rect [10 40 90 60] /Border [0 0 0] /A << /S /GoTo /D (Chapter2) >> >>
endobj
trailer
<< /Root 1 0 R /Size 8 >>
startxref
0
%%EOF
`

func TestInternalLinks(t *testing.T) {
	doc, err := pdf.New([]byte(internalLinkPDF), 0)
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Release()

	page, err := doc.RenderPage(0, 72, 0, "") // 72 dpi => scale 1.0, so DestPoint values are page points
	if err != nil {
		t.Fatal(err)
	}

	// Page 0 carries two internal links — one explicit /XYZ destination and one named destination — both pointing at
	// the second page (0-based index 1). Each must resolve to PageNumber 1 with an empty URI. The named destination in
	// particular was silently dropped by the previous "#page=" string parsing, and the page index must be 0-based: the
	// target is the second page object, so 1 rather than 0.
	if len(page.Links) != 2 {
		t.Fatalf("expected 2 internal links, got %d", len(page.Links))
	}
	// The /XYZ destination (left 30, top 150 on a 200-tall page) resolves to (30, 50) in top-left/y-down image space;
	// the named /Fit destination has no explicit point and so resolves to (0, 0). Match by DestPoint rather than order.
	var sawXYZ, sawFit bool
	for i, l := range page.Links {
		if l.PageNumber != 1 {
			t.Errorf("link %d: expected PageNumber 1 (0-based second page), got %d", i, l.PageNumber)
		}
		if l.URI != "" {
			t.Errorf("link %d: expected empty URI for an internal link, got %q", i, l.URI)
		}
		switch l.DestPoint {
		case image.Pt(30, 50):
			sawXYZ = true
		case image.Pt(0, 0):
			sawFit = true
		default:
			t.Errorf("link %d: unexpected DestPoint %v", i, l.DestPoint)
		}
	}
	if !sawXYZ {
		t.Error("expected a link with the /XYZ DestPoint (30, 50)")
	}
	if !sawFit {
		t.Error("expected a link with the /Fit DestPoint (0, 0)")
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
