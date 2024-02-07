package pdf

/*
#cgo CFLAGS: -Iinclude
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_darwin_amd64 -lm
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_darwin_arm64 -lm -ld_classic
#cgo linux LDFLAGS: -L${SRCDIR}/lib -lmupdf_linux_amd64 -lm
#cgo windows LDFLAGS: -L${SRCDIR}/lib -lmupdf_windows_amd64 -lm -Wl,--allow-multiple-definition

#include <stdlib.h>
#include <mupdf/fitz.h>

const char *version = FZ_VERSION;
const char *pdfMimeType = "application/pdf";

// Wrappers for cases where "exceptions" can be thrown

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
*/
import "C"

import (
	"bytes"
	"errors"
	"image"
	"math"
	"runtime"
	"strconv"
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
	ErrInvalidPageSize          = errors.New("invalid page size")
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

// Document represents PDF document.
type Document struct {
	ctx  *C.fz_context
	doc  *C.fz_document
	data *C.uchar
	lock sync.Mutex
}

// TOCEntry holds a single entry in the table of contents.
type TOCEntry struct {
	Title      string
	PageNumber int
	PageX      int
	PageY      int
	Children   []*TOCEntry
}

// PageLink holds a single link on a page. If PageNumber if >= 0, then this is an internal link and the URI will be
// empty.
type PageLink struct {
	PageNumber int
	URI        string
	Bounds     image.Rectangle
}

// RenderedPage holds the rendered page.
type RenderedPage struct {
	Image      *image.NRGBA
	SearchHits []image.Rectangle
	Links      []*PageLink
}

// New returns new PDF document from the provided raw bytes. Pass in 0 for maxCacheSize for no limit.
func New(buffer []byte, maxCacheSize uint64) (*Document, error) {
	if !bytes.HasPrefix(buffer, []byte("%PDF")) {
		return nil, ErrNotPDFData
	}
	var d Document
	d.ctx = C.fz_new_context_imp(nil, nil, C.size_t(maxCacheSize), C.version)
	if d.ctx == nil {
		return nil, ErrUnableToCreatePDFContext
	}
	C.fz_register_document_handlers(d.ctx)
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
	d.doc = C.fz_open_document_with_stream(d.ctx, C.pdfMimeType, stream)
	C.fz_drop_stream(d.ctx, stream)
	if d.doc == nil {
		d.Release()
		return nil, ErrUnableToOpenPDF
	}
	runtime.SetFinalizer(&d, func(obj *Document) { obj.Release() })
	return &d, nil
}

// RequiresAuthentication returns true if a password is required.
func (d *Document) RequiresAuthentication() bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	return C.fz_needs_password(d.ctx, d.doc) != 0
}

// Authenticate with either the user or owner password.
func (d *Document) Authenticate(password string) AuthenticationStatus {
	d.lock.Lock()
	defer d.lock.Unlock()
	pw := C.CString(password)
	defer C.free(unsafe.Pointer(pw))
	return AuthenticationStatus(C.fz_authenticate_password(d.ctx, d.doc, pw))
}

// TableOfContents returns the table of contents for this document, if any.
func (d *Document) TableOfContents(dpi int) []*TOCEntry {
	d.lock.Lock()
	defer d.lock.Unlock()
	outline := C.fz_load_outline(d.ctx, d.doc)
	if outline == nil {
		return nil
	}
	defer C.fz_drop_outline(d.ctx, outline)
	return buildTOCEntries(outline, float32(dpiToScale(dpi)))
}

func buildTOCEntries(outline *C.fz_outline, scale float32) []*TOCEntry {
	var entries []*TOCEntry
	for outline != nil {
		entry := &TOCEntry{
			PageNumber: int(outline.page.page),
			PageX:      int(math.Floor(float64(outline.x) * float64(scale))),
			PageY:      int(math.Floor(float64(outline.y) * float64(scale))),
		}
		if outline.title != nil {
			entry.Title = sanitizeString(outline.title)
		}
		entries = append(entries, entry)
		if outline.down != nil {
			entry.Children = buildTOCEntries(outline.down, scale)
		}
		outline = outline.next
	}
	return entries
}

func sanitizeString(in *C.char) string {
	str := C.GoString(in)
	sanitized := make([]rune, 0, len(str))
	for _, ch := range str {
		if !unicode.IsControl(ch) && unicode.IsPrint(ch) {
			sanitized = append(sanitized, ch)
		}
	}
	return strings.TrimSpace(string(sanitized))
}

// PageCount returns total number of pages in the document.
func (d *Document) PageCount() int {
	d.lock.Lock()
	defer d.lock.Unlock()
	return int(C.fz_count_pages(d.ctx, d.doc))
}

func dpiToScale(dpi int) float64 {
	scale := float64(dpi) / 72
	if scale > 10 {
		return 10 // Limit scaling to 10x; some displays report bad EDID data, causing the input DPI from programs to be wildly off
	}
	return scale
}

// RenderPage renders the specified page at the requested dpi. If search is not empty, then the bounding boxes of up to
// maxHits matching text on the page will be returned.
func (d *Document) RenderPage(pageNumber, dpi, maxHits int, search string) (*RenderedPage, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	pageCount := int(C.fz_count_pages(d.ctx, d.doc))
	if pageNumber >= pageCount {
		return nil, ErrInvalidPageNumber
	}
	page := C.fz_load_page(d.ctx, d.doc, C.int(pageNumber))
	if page == nil {
		return nil, ErrUnableToLoadPage
	}
	defer C.fz_drop_page(d.ctx, page)
	displayList := C.wrapped_fz_new_display_list_from_page(d.ctx, page)
	defer C.fz_drop_display_list(d.ctx, displayList)
	scale := dpiToScale(dpi)
	img := d.renderPage(displayList, scale)
	if img == nil {
		return nil, ErrUnableToCreateImage
	}
	return &RenderedPage{
		Image:      img,
		SearchHits: d.searchDisplayList(displayList, scale, search, maxHits),
		Links:      d.loadLinks(page, scale),
	}, nil
}

// RenderPageForSize renders the specified page to fit within the requested size. If search is not empty, then the
// bounding boxes of up to maxHits matching text on the page will be returned.
func (d *Document) RenderPageForSize(pageNumber, maxWidth, maxHeight, maxHits int, search string) (*RenderedPage, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	pageCount := int(C.fz_count_pages(d.ctx, d.doc))
	if pageNumber >= pageCount {
		return nil, ErrInvalidPageNumber
	}
	page := C.fz_load_page(d.ctx, d.doc, C.int(pageNumber))
	if page == nil {
		return nil, ErrUnableToLoadPage
	}
	defer C.fz_drop_page(d.ctx, page)
	displayList := C.wrapped_fz_new_display_list_from_page(d.ctx, page)
	defer C.fz_drop_display_list(d.ctx, displayList)
	r := C.fz_bound_page(d.ctx, page)
	w := float64(r.x1 - r.x0)
	h := float64(r.y1 - r.y0)
	if w <= 0 || h <= 0 {
		return nil, ErrInvalidPageSize
	}
	scale := float64(maxWidth) / w
	ratio := float64(maxHeight) / h
	if scale > ratio {
		scale = ratio
	}
	if scale <= 0 {
		return nil, ErrInvalidPageSize
	}
	img := d.renderPage(displayList, scale)
	if img == nil {
		return nil, ErrUnableToCreateImage
	}
	return &RenderedPage{
		Image:      img,
		SearchHits: d.searchDisplayList(displayList, scale, search, maxHits),
		Links:      d.loadLinks(page, scale),
	}, nil
}

func (d *Document) renderPage(displayList *C.fz_display_list, scale float64) *image.NRGBA {
	ctm := C.fz_scale(C.float(scale), C.float(scale))
	cs := C.fz_device_rgb(d.ctx)
	pixmap := C.wrapped_fz_new_pixmap_from_display_list(d.ctx, displayList, ctm, cs, 1)
	if pixmap == nil {
		return nil
	}
	defer C.fz_drop_pixmap(d.ctx, pixmap)
	pixels := C.fz_pixmap_samples(d.ctx, pixmap)
	if pixels == nil {
		return nil
	}
	size := int(pixmap.stride) * int(pixmap.h)
	if size <= 0 || size > math.MaxInt32 {
		return nil
	}
	return &image.NRGBA{
		Pix:    C.GoBytes(unsafe.Pointer(pixels), C.int(size)),
		Stride: int(pixmap.stride),
		Rect:   image.Rect(0, 0, int(pixmap.w), int(pixmap.h)),
	}
}

func (d *Document) searchDisplayList(displayList *C.fz_display_list, scale float64, search string, maxHits int) []image.Rectangle {
	var boxes []image.Rectangle
	if search != "" {
		searchText := C.CString(search)
		defer C.free(unsafe.Pointer(searchText))
		quads := make([]C.fz_quad, maxHits)
		hits := C.wrapped_fz_search_display_list(d.ctx, displayList, searchText, nil, (*C.fz_quad)(unsafe.Pointer(&quads[0])), C.int(len(quads)))
		if hits > 0 {
			boxes = make([]image.Rectangle, hits)
			for i := range boxes {
				boxes[i] = image.Rect(int(math.Floor(math.Min(float64(quads[i].ul.x), float64(quads[i].ll.x))*scale)),
					int(math.Floor(math.Min(float64(quads[i].ul.y), float64(quads[i].ur.y))*scale)),
					int(math.Ceil(math.Max(float64(quads[i].ur.x), float64(quads[i].lr.x))*scale)),
					int(math.Ceil(math.Max(float64(quads[i].ll.y), float64(quads[i].lr.y))*scale)),
				)
			}
		}
	}
	return boxes
}

func (d *Document) loadLinks(page *C.fz_page, scale float64) []*PageLink {
	var links []*PageLink
	if link := C.fz_load_links(d.ctx, page); link != nil {
		firstLink := link
		for link != nil {
			pageLink := &PageLink{
				PageNumber: -1,
				URI:        sanitizeString(link.uri),
				Bounds: image.Rect(int(math.Floor(float64(link.rect.x0)*scale)),
					int(math.Floor(float64(link.rect.y0)*scale)),
					int(math.Ceil(float64(link.rect.x1)*scale)),
					int(math.Ceil(float64(link.rect.y1)*scale)),
				),
			}
			if strings.HasPrefix(pageLink.URI, "#") {
				const pagePrefix = "#page="
				if i := strings.Index(pageLink.URI, pagePrefix); i != -1 {
					pageLink.URI = pageLink.URI[i+len(pagePrefix):]
					if i = strings.Index(pageLink.URI, "&"); i != -1 {
						pageLink.URI = pageLink.URI[:i]
					}
					pageLink.PageNumber, _ = strconv.Atoi(pageLink.URI) //nolint:errcheck // Failure here results in 0, which is acceptable
					pageLink.PageNumber--                               // Page numbers in links seem to be 1-based, but we use 0-based internally
				}
				pageLink.URI = ""
			}
			if pageLink.PageNumber != -1 || pageLink.URI != "" {
				links = append(links, pageLink)
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
