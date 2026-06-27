package pdf

/*
#cgo CFLAGS: -Iinclude
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_darwin_amd64 -lm
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_darwin_arm64 -lm
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_linux_amd64 -lm
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_linux_arm64 -lm
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_windows_amd64 -lm -Wl,--allow-multiple-definition
#cgo windows,arm64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_windows_arm64 -lm -Wl,--allow-multiple-definition

#include <stdlib.h>
#include <mupdf/fitz.h>

// Wrappers for cases where "exceptions" can be thrown or where a macro is used

fz_context *wrapped_fz_new_context(const fz_alloc_context *alloc, const fz_locks_context *locks, size_t max_store) {
	return fz_new_context(alloc, locks, max_store);
}

fz_document *wrapped_fz_open_pdf_document_with_stream(fz_context *ctx, fz_stream *stream) {
	fz_document *doc = NULL;
	fz_var(doc);
	fz_try(ctx) {
		doc = fz_open_document_with_stream(ctx, "application/pdf", stream);
	}
	fz_catch(ctx) {
		doc = NULL;
	}
	return doc;
}

fz_stream *wrapped_fz_open_memory(fz_context *ctx, const unsigned char *data, size_t len) {
	fz_stream *stream = NULL;
	fz_var(stream);
	fz_try(ctx) {
		stream = fz_open_memory(ctx, data, len);
	}
	fz_catch(ctx) {
		stream = NULL;
	}
	return stream;
}

fz_display_list *wrapped_fz_new_display_list_from_page(fz_context *ctx, fz_page *page) {
	fz_display_list *list = NULL;
	fz_var(list);
	fz_try(ctx) {
		list = fz_new_display_list_from_page(ctx, page);
	}
	fz_catch(ctx) {
		list = NULL;
	}
	return list;
}

fz_pixmap *wrapped_fz_new_pixmap_from_display_list(fz_context *ctx, fz_display_list *list, fz_matrix ctm, fz_colorspace *cs, int alpha) {
	fz_pixmap *pixmap = NULL;
	fz_var(pixmap);
	fz_try(ctx) {
		pixmap = fz_new_pixmap_from_display_list(ctx, list, ctm, cs, alpha);
	}
	fz_catch(ctx) {
		pixmap = NULL;
	}
	return pixmap;
}

int wrapped_fz_search_display_list(fz_context *ctx, fz_display_list *list, const char *needle, int *hit_mark, fz_quad *hit_bbox, int hit_max) {
	int hits = 0;
	fz_var(hits);
	fz_try(ctx) {
		hits = fz_search_display_list(ctx, list, needle, hit_mark, hit_bbox, hit_max);
	}
	fz_catch(ctx) {
		hits = 0;
	}
	return hits;
}

fz_outline *wrapped_fz_load_outline(fz_context *ctx, fz_document *doc) {
	fz_outline *outline = NULL;
	fz_var(outline);
	fz_try(ctx) {
		outline = fz_load_outline(ctx, doc);
	}
	fz_catch(ctx) {
		outline = NULL;
	}
	return outline;
}

// Returns 1 on success, 0 if registration threw.
int wrapped_fz_register_document_handlers(fz_context *ctx) {
	int ok = 0;
	fz_var(ok);
	fz_try(ctx) {
		fz_register_document_handlers(ctx);
		ok = 1;
	}
	fz_catch(ctx) {
		ok = 0;
	}
	return ok;
}

// Returns the result of fz_needs_password, or 0 if it threw.
int wrapped_fz_needs_password(fz_context *ctx, fz_document *doc) {
	int needs = 0;
	fz_var(needs);
	fz_try(ctx) {
		needs = fz_needs_password(ctx, doc);
	}
	fz_catch(ctx) {
		needs = 0;
	}
	return needs;
}

// Returns the result of fz_authenticate_password, or 0 (failure) if it threw.
int wrapped_fz_authenticate_password(fz_context *ctx, fz_document *doc, const char *password) {
	int result = 0;
	fz_var(result);
	fz_try(ctx) {
		result = fz_authenticate_password(ctx, doc, password);
	}
	fz_catch(ctx) {
		result = 0;
	}
	return result;
}

// Returns the page count, or -1 if it threw.
int wrapped_fz_count_pages(fz_context *ctx, fz_document *doc) {
	int count = -1;
	fz_var(count);
	fz_try(ctx) {
		count = fz_count_pages(ctx, doc);
	}
	fz_catch(ctx) {
		count = -1;
	}
	return count;
}

fz_page *wrapped_fz_load_page(fz_context *ctx, fz_document *doc, int number) {
	fz_page *page = NULL;
	fz_var(page);
	fz_try(ctx) {
		page = fz_load_page(ctx, doc, number);
	}
	fz_catch(ctx) {
		page = NULL;
	}
	return page;
}

// Returns the page bounds, or a zero rect if it threw.
fz_rect wrapped_fz_bound_page(fz_context *ctx, fz_page *page) {
	fz_rect rect = fz_empty_rect;
	fz_var(rect);
	fz_try(ctx) {
		rect = fz_bound_page(ctx, page);
	}
	fz_catch(ctx) {
		rect = fz_empty_rect;
	}
	return rect;
}

fz_link *wrapped_fz_load_links(fz_context *ctx, fz_page *page) {
	fz_link *links = NULL;
	fz_var(links);
	fz_try(ctx) {
		links = fz_load_links(ctx, page);
	}
	fz_catch(ctx) {
		links = NULL;
	}
	return links;
}

// Returns 1 if the link URI is external, 0 if internal (or if uri is NULL or the check threw).
int wrapped_fz_is_external_link(fz_context *ctx, const char *uri) {
	int result = 0;
	if (uri == NULL) {
		return 0;
	}
	fz_var(result);
	fz_try(ctx) {
		result = fz_is_external_link(ctx, uri);
	}
	fz_catch(ctx) {
		result = 0;
	}
	return result;
}

typedef struct {
	int page;
	float x;
	float y;
} resolved_link;

// Resolves an internal link URI to a 0-based page number and a destination point on that page. page is -1 (and the
// point 0,0) if uri is NULL, it cannot be resolved, or it threw. When the destination carries no explicit coordinate
// (e.g. a /Fit destination) MuPDF leaves x/y non-finite (NaN); the Go caller maps that to 0,0 (see scaledFloor).
resolved_link wrapped_fz_resolve_link(fz_context *ctx, fz_document *doc, const char *uri) {
	resolved_link r = { -1, 0, 0 };
	if (uri == NULL) {
		return r;
	}
	fz_var(r);
	fz_try(ctx) {
		float x = 0, y = 0;
		fz_location loc = fz_resolve_link(ctx, doc, uri, &x, &y);
		r.page = fz_page_number_from_location(ctx, doc, loc);
		r.x = x;
		r.y = y;
	}
	fz_catch(ctx) {
		r.page = -1;
		r.x = 0;
		r.y = 0;
	}
	return r;
}
*/
import "C"

import (
	"bytes"
	"errors"
	"image"
	"math"
	"runtime"
	"strings"
	"sync"
	"unicode"
	"unsafe"
)

// Possible error values
var (
	ErrNotPDFData               = errors.New("only PDF documents are supported")
	ErrUnableToCreatePDFContext = errors.New("unable to create PDF context")
	ErrInternal                 = errors.New("internal error")
	ErrUnableToOpenPDF          = errors.New("unable to open PDF")
	ErrInvalidPageNumber        = errors.New("invalid page number")
	ErrUnableToLoadPage         = errors.New("unable to load page")
	ErrUnableToCreateImage      = errors.New("unable to create image")
	ErrImageTooLarge            = errors.New("rendered image would be too large")
	ErrInvalidPageSize          = errors.New("invalid page size")
	ErrDocumentReleased         = errors.New("document has been released")
)

// Each of these variables is global and are not safe to modify when other calls to this code are being made. Generally,
// they should be modified at startup before any other use of this package.
var (
	// OverallMaxHits is the maximum number of hits returned, even if the API is called with a larger value. This is
	// here to safeguard against untrusted input that might otherwise cause an out of memory error.
	OverallMaxHits = 1000
	// OverallMaxLinks is the maximum number of links returned. This is here to safeguard against untrusted input that
	// might otherwise cause an out of memory error.
	OverallMaxLinks = 1000
	// OverallMaxTOCEntries is the maximum number of TOC entries returned. This is here to safeguard against untrusted
	// input that might otherwise cause an out of memory error.
	OverallMaxTOCEntries = 1000
	// OverallMaxPixels is the maximum number of pixels (width × height) a rendered page image may contain. Requests
	// that would produce a larger image are rejected rather than attempting a very large allocation, safeguarding
	// against untrusted input or bad sizing parameters that might otherwise cause an out of memory error. The default
	// matches the largest image permitted by the internal 32-bit limit on the rendered buffer's byte size (4 bytes per
	// pixel).
	OverallMaxPixels = math.MaxInt32 / 4
)

// AuthenticationStatus holds the result of an authentication attempt. A non-zero value indicates success and the masks
// can be used to determine further details.
type AuthenticationStatus byte

// Masks that can be used to examine AuthenticationStatus for additional details.
const (
	NoAuthenticationRequiredMask AuthenticationStatus = 1 << iota
	UserAuthenticatedMask
	OwnerAuthenticatedMask
)

type document struct {
	ctx  *C.fz_context
	doc  *C.fz_document
	data *C.uchar
	lock sync.Mutex
}

// Document represents PDF document. Page numbers for the exposed API are zero-based. Methods on this are safe to use
// from multiple goroutines. Calls into the underlying MuPDF library are serialized internally, so they execute one at a
// time.
type Document struct {
	// document is held by pointer so it lives in its own heap allocation, separate from the Document wrapper. This is
	// required by runtime.AddCleanup(): the cleanup arg must not point into the same allocation as the tracked pointer,
	// otherwise the tracked object can never become unreachable and the cleanup would never run.
	*document
}

// TOCEntry holds a single entry in the table of contents.
type TOCEntry struct {
	Title      string
	Children   []*TOCEntry
	PageNumber int
	PageX      int
	PageY      int
}

// PageLink holds a single link on a page. If PageNumber is >= 0, then this is an internal link and the URI will be
// empty.
type PageLink struct {
	URI        string
	PageNumber int
	// Bounds is the clickable hot-zone of the link on the page it appears on, in rendered-image pixel space.
	Bounds image.Rectangle
	// DestPoint is the point on the destination page (PageNumber) that an internal link targets, in rendered-image
	// pixel space. It is the zero value (0,0) for external links and for internal links whose destination has no
	// explicit coordinate (such as a /Fit destination).
	DestPoint image.Point
}

// RenderedPage holds the rendered page.
type RenderedPage struct {
	// Image is the rendered page. It is rendered with an alpha channel, and most PDF pages do not paint their own
	// background, so areas with no content are transparent rather than white. Callers that want an opaque page (for
	// example, when encoding to a format without alpha) should composite the image onto their desired background color.
	Image      *image.NRGBA
	SearchHits []image.Rectangle
	Links      []*PageLink
}

// New returns new PDF document from the provided raw bytes. Pass in 0 for maxCacheSize for no limit.
func New(buffer []byte, maxCacheSize uint64) (*Document, error) {
	// Allow some garbage to be before the PDF content, as Acrobat and MuPDF itself allow it
	if !bytes.Contains(buffer[:min(1024, len(buffer))], []byte("%PDF")) {
		return nil, ErrNotPDFData
	}
	d := Document{
		document: &document{
			ctx: C.wrapped_fz_new_context(nil, nil, C.size_t(maxCacheSize)),
		},
	}
	if d.ctx == nil {
		return nil, ErrUnableToCreatePDFContext
	}
	if C.wrapped_fz_register_document_handlers(d.ctx) == 0 {
		d.Release()
		return nil, ErrUnableToCreatePDFContext
	}
	d.data = (*C.uchar)(C.CBytes(buffer))
	if d.data == nil {
		d.Release()
		return nil, ErrInternal
	}
	stream := C.wrapped_fz_open_memory(d.ctx, d.data, C.size_t(len(buffer)))
	if stream == nil {
		d.Release()
		return nil, ErrInternal
	}
	d.doc = C.wrapped_fz_open_pdf_document_with_stream(d.ctx, stream)
	C.fz_drop_stream(d.ctx, stream)
	if d.doc == nil {
		d.Release()
		return nil, ErrUnableToOpenPDF
	}
	runtime.AddCleanup(&d, func(doc *document) { doc.release() }, d.document)
	return &d, nil
}

// released reports whether the underlying document has been released. The caller must hold d.lock.
func (d *document) released() bool {
	return d.ctx == nil || d.doc == nil
}

// RequiresAuthentication returns true if a password is required. Returns false if the document has been released.
func (d *Document) RequiresAuthentication() bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.released() {
		return false
	}
	return C.wrapped_fz_needs_password(d.ctx, d.doc) != 0
}

// Authenticate with either the user or owner password. Returns a zero status if the document has been released.
func (d *Document) Authenticate(password string) AuthenticationStatus {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.released() {
		return 0
	}
	pw := C.CString(password)
	defer C.free(unsafe.Pointer(pw))
	return AuthenticationStatus(C.wrapped_fz_authenticate_password(d.ctx, d.doc, pw))
}

// TableOfContents returns the table of contents for this document, if any.
func (d *Document) TableOfContents(dpi int) []*TOCEntry {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.released() {
		return nil
	}
	outline := C.wrapped_fz_load_outline(d.ctx, d.doc)
	if outline == nil {
		return nil
	}
	defer C.fz_drop_outline(d.ctx, outline)
	entries, _ := buildTOCEntries(outline, float32(dpiToScale(dpi)), OverallMaxTOCEntries)
	return entries
}

func buildTOCEntries(outline *C.fz_outline, scale float32, maxAllowed int) (entries []*TOCEntry, remaining int) {
	if maxAllowed < 1 {
		return nil, 0
	}
	for outline != nil {
		entry := &TOCEntry{
			PageNumber: int(outline.page.page),
			PageX:      scaledFloor(float64(outline.x), float64(scale)),
			PageY:      scaledFloor(float64(outline.y), float64(scale)),
		}
		if outline.title != nil {
			entry.Title = sanitizeString(outline.title)
		}
		entries = append(entries, entry)
		maxAllowed--
		if maxAllowed <= 0 {
			break
		}
		if outline.down != nil {
			if entry.Children, maxAllowed = buildTOCEntries(outline.down, scale, maxAllowed); maxAllowed <= 0 {
				break
			}
		}
		outline = outline.next
	}
	return entries, max(maxAllowed, 0)
}

func sanitizeString(in *C.char) string {
	str := C.GoString(in)
	sanitized := make([]rune, 0, len(str))
	for _, ch := range str {
		// U+FFFD (the Unicode replacement character) stands in for bytes that could not be decoded as valid UTF-8,
		// such as the unmappable dot-leader glyphs some PDFs place in outline titles. It is printable and non-control,
		// so it would otherwise survive the filter below; drop it explicitly to keep those spurious characters out.
		if ch != unicode.ReplacementChar && !unicode.IsControl(ch) && unicode.IsPrint(ch) {
			sanitized = append(sanitized, ch)
		}
	}
	return strings.TrimSpace(string(sanitized))
}

// PageCount returns total number of pages in the document.
func (d *Document) PageCount() int {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.released() {
		return 0
	}
	if count := int(C.wrapped_fz_count_pages(d.ctx, d.doc)); count > 0 {
		return count
	}
	return 0
}

func dpiToScale(dpi int) float64 {
	// Limit scaling to 10x; some displays report bad EDID data, causing the input DPI from programs to be wildly off.
	return min(float64(max(dpi, 1))/72, 10)
}

// RenderPage renders the specified page at the requested dpi. If search is not empty, then the bounding boxes of up to
// maxHits matching text on the page will be returned.
func (d *Document) RenderPage(pageNumber, dpi, maxHits int, search string) (*RenderedPage, error) {
	return d.render(pageNumber, maxHits, search, func(*C.fz_page) (float64, error) {
		return dpiToScale(dpi), nil
	})
}

// RenderPageForSize renders the specified page to fit within the requested size. If search is not empty, then the
// bounding boxes of up to maxHits matching text on the page will be returned.
func (d *Document) RenderPageForSize(pageNumber, maxWidth, maxHeight, maxHits int, search string) (*RenderedPage, error) {
	return d.render(pageNumber, maxHits, search, func(page *C.fz_page) (float64, error) {
		if maxWidth <= 0 || maxHeight <= 0 {
			return 0, ErrInvalidPageSize
		}
		r := C.wrapped_fz_bound_page(d.ctx, page)
		w := float64(r.x1 - r.x0)
		h := float64(r.y1 - r.y0)
		if w <= 0 || h <= 0 {
			return 0, ErrInvalidPageSize
		}
		scale := float64(maxWidth) / w
		if ratio := float64(maxHeight) / h; ratio < scale {
			scale = ratio
		}
		// The rendered image is scaled to fit within maxWidth×maxHeight, so its pixel count is bounded by the
		// requested box. Reject an over-large request here, before building the display list or asking MuPDF to
		// allocate the pixmap.
		if (w*scale)*(h*scale) > float64(OverallMaxPixels) {
			return 0, ErrImageTooLarge
		}
		return scale, nil
	})
}

// render is the shared body of RenderPage and RenderPageForSize. It validates the page number, loads the page, asks
// scaleFor to compute the render scale (which may inspect the page bounds and reject the request), builds the display
// list, renders, and assembles the result. The document lock is held throughout so the underlying C calls are
// serialized.
func (d *Document) render(pageNumber, maxHits int, search string, scaleFor func(page *C.fz_page) (float64, error)) (*RenderedPage, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.released() {
		return nil, ErrDocumentReleased
	}
	pageCount := int(C.wrapped_fz_count_pages(d.ctx, d.doc))
	if pageNumber < 0 || pageNumber >= pageCount {
		return nil, ErrInvalidPageNumber
	}
	page := C.wrapped_fz_load_page(d.ctx, d.doc, C.int(pageNumber))
	if page == nil {
		return nil, ErrUnableToLoadPage
	}
	defer C.fz_drop_page(d.ctx, page)
	scale, err := scaleFor(page)
	if err != nil {
		return nil, err
	}
	displayList := C.wrapped_fz_new_display_list_from_page(d.ctx, page)
	if displayList == nil {
		return nil, ErrUnableToCreateImage
	}
	defer C.fz_drop_display_list(d.ctx, displayList)
	img, err := d.renderPage(displayList, scale)
	if err != nil {
		return nil, err
	}
	return &RenderedPage{
		Image:      img,
		SearchHits: d.searchDisplayList(displayList, scale, search, maxHits),
		Links:      d.loadLinks(page, scale),
	}, nil
}

func (d *Document) renderPage(displayList *C.fz_display_list, scale float64) (*image.NRGBA, error) {
	ctm := C.fz_scale(C.float(scale), C.float(scale))
	cs := C.fz_device_rgb(d.ctx)
	pixmap := C.wrapped_fz_new_pixmap_from_display_list(d.ctx, displayList, ctm, cs, 1)
	if pixmap == nil {
		return nil, ErrUnableToCreateImage
	}
	defer C.fz_drop_pixmap(d.ctx, pixmap)
	if pixmap.w <= 0 || pixmap.h <= 0 {
		return nil, ErrUnableToCreateImage
	}
	if int64(pixmap.w)*int64(pixmap.h) > int64(OverallMaxPixels) {
		return nil, ErrImageTooLarge
	}
	pixels := C.fz_pixmap_samples(d.ctx, pixmap)
	if pixels == nil {
		return nil, ErrUnableToCreateImage
	}
	size := int(pixmap.stride) * int(pixmap.h)
	if size <= 0 {
		return nil, ErrUnableToCreateImage
	}
	if size > math.MaxInt32 {
		return nil, ErrImageTooLarge
	}
	pix := C.GoBytes(unsafe.Pointer(pixels), C.int(size))
	// MuPDF renders with premultiplied alpha, but image.NRGBA expects non-premultiplied (straight) alpha, so undo the
	// premultiplication. Fully opaque (a == 255) and fully transparent (a == 0) pixels need no adjustment.
	for i := 0; i+3 < len(pix); i += 4 {
		switch a := pix[i+3]; a {
		case 0, 255:
		default:
			pix[i] = unpremultiply(pix[i], a)
			pix[i+1] = unpremultiply(pix[i+1], a)
			pix[i+2] = unpremultiply(pix[i+2], a)
		}
	}
	return &image.NRGBA{
		Pix:    pix,
		Stride: int(pixmap.stride),
		Rect:   image.Rect(0, 0, int(pixmap.w), int(pixmap.h)),
	}, nil
}

// unpremultiply converts a single premultiplied color component back to its straight-alpha value, rounding to nearest
// and clamping to 0xff. The caller guarantees a is neither 0 nor 0xff.
func unpremultiply(c, a uint8) uint8 {
	v := (int(c)*0xff + int(a)/2) / int(a)
	if v > 0xff {
		return 0xff
	}
	return uint8(v)
}

func (d *Document) searchDisplayList(displayList *C.fz_display_list, scale float64, search string, maxHits int) []image.Rectangle {
	var boxes []image.Rectangle
	if search != "" && maxHits > 0 && OverallMaxHits > 0 {
		searchText := C.CString(search)
		defer C.free(unsafe.Pointer(searchText))
		quads := make([]C.fz_quad, min(maxHits, OverallMaxHits))
		hits := C.wrapped_fz_search_display_list(d.ctx, displayList, searchText, nil, (*C.fz_quad)(unsafe.Pointer(&quads[0])), C.int(len(quads)))
		if hits > 0 {
			boxes = make([]image.Rectangle, hits)
			for i := range boxes {
				boxes[i] = quadToRect(quads[i], scale)
			}
		}
	}
	return boxes
}

// quadToRect computes the scaled, axis-aligned bounding rectangle that encloses all four corners of a search-hit quad.
// Considering every corner (rather than assuming an axis-aligned quad) keeps the box correct for rotated or skewed text.
func quadToRect(q C.fz_quad, scale float64) image.Rectangle {
	minX := math.Min(math.Min(float64(q.ul.x), float64(q.ur.x)), math.Min(float64(q.ll.x), float64(q.lr.x)))
	minY := math.Min(math.Min(float64(q.ul.y), float64(q.ur.y)), math.Min(float64(q.ll.y), float64(q.lr.y)))
	maxX := math.Max(math.Max(float64(q.ul.x), float64(q.ur.x)), math.Max(float64(q.ll.x), float64(q.lr.x)))
	maxY := math.Max(math.Max(float64(q.ul.y), float64(q.ur.y)), math.Max(float64(q.ll.y), float64(q.lr.y)))
	return scaleRect(minX, minY, maxX, maxY, scale)
}

// scaleRect scales an axis-aligned rectangle by scale and converts it to integer pixel space, expanding outward so the
// box never clips its content: the min corner is floored and the max corner is ceiled.
func scaleRect(x0, y0, x1, y1, scale float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(x0*scale)),
		int(math.Floor(y0*scale)),
		int(math.Ceil(x1*scale)),
		int(math.Ceil(y1*scale)),
	)
}

// scaledFloor multiplies v by scale, floors the result, and converts it to an int. MuPDF reports a destination that
// carries no explicit coordinate (e.g. a /Fit destination, in both link targets and TOC entries) as a non-finite
// value; Go's conversion of a non-finite (or out-of-range) float to int is architecture-defined — 0 on arm64 but
// math.MinInt64 on amd64 — so those values are mapped to 0 here to keep the returned coordinates deterministic across
// architectures.
func scaledFloor(v, scale float64) int {
	r := math.Floor(v * scale)
	if math.IsNaN(r) || r < math.MinInt || r > math.MaxInt {
		return 0
	}
	return int(r)
}

func (d *Document) loadLinks(page *C.fz_page, scale float64) []*PageLink {
	if OverallMaxLinks < 1 {
		return nil
	}
	var links []*PageLink
	if link := C.wrapped_fz_load_links(d.ctx, page); link != nil {
		firstLink := link
		for link != nil {
			pageLink := &PageLink{
				PageNumber: -1,
				Bounds: scaleRect(float64(link.rect.x0), float64(link.rect.y0),
					float64(link.rect.x1), float64(link.rect.y1), scale),
			}
			// External links keep their URI; internal links are resolved to a page. fz_resolve_link returns the
			// target location and fz_page_number_from_location turns it into MuPDF's 0-based page number, which is the
			// same numbering fz_load_page and this package's API use, so no adjustment is needed. Internal links that
			// cannot be resolved come back as -1 and, with an empty URI, are dropped by the test below.
			if C.wrapped_fz_is_external_link(d.ctx, link.uri) != 0 {
				pageLink.URI = sanitizeString(link.uri)
			} else {
				res := C.wrapped_fz_resolve_link(d.ctx, d.doc, link.uri)
				pageLink.PageNumber = int(res.page)
				pageLink.DestPoint = image.Pt(
					scaledFloor(float64(res.x), scale),
					scaledFloor(float64(res.y), scale),
				)
			}
			if pageLink.PageNumber >= 0 || pageLink.URI != "" {
				if links = append(links, pageLink); len(links) >= OverallMaxLinks {
					break
				}
			}
			link = link.next
		}
		C.fz_drop_link(d.ctx, firstLink)
	}
	return links
}

// Release the underlying PDF document, releasing any resources. It is not necessary to call this, as garbage collection
// will eventually do this for you, however, doing so explicitly will cause an immediate reclamation of any used memory.
func (d *Document) Release() {
	d.release()
}

func (d *document) release() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.doc != nil {
		C.fz_drop_document(d.ctx, d.doc)
		d.doc = nil
	}
	if d.data != nil {
		C.free(unsafe.Pointer(d.data))
		d.data = nil
	}
	if d.ctx != nil {
		C.fz_drop_context(d.ctx)
		d.ctx = nil
	}
}
