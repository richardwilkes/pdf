package pdf

/*
#cgo CFLAGS: -Iinclude
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_darwin_amd64 -lm
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib -lmupdf_darwin_arm64 -lm
#cgo linux LDFLAGS: -L${SRCDIR}/lib -lmupdf_linux_amd64 -lm
#cgo windows LDFLAGS: -L${SRCDIR}/lib -lmupdf_windows_amd64 -lm

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

fz_display_list *wrapped_fz_new_display_list_from_page_number(fz_context *ctx, fz_document *doc, int number) {
	fz_display_list *list = NULL;
	fz_var(list);
	fz_try(ctx) {
		list = fz_new_display_list_from_page_number(ctx, doc, number);
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

int wrapped_fz_search_display_list(fz_context *ctx, fz_display_list *list, const char *needle, fz_quad *hit_bbox, int hit_max) {
	int hits = 0;
	fz_var(hits);
	fz_try(ctx) {
		hits = fz_search_display_list(ctx, list, needle, hit_bbox, hit_max);
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
	"image"
	"image/draw"
	"runtime"
	"sync"
	"unsafe"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xmath/geom32"
)

// Document represents PDF document.
type Document struct {
	ctx  *C.fz_context
	doc  *C.fz_document
	data *C.uchar
	lock sync.Mutex
}

// New returns new PDF document from the provided raw bytes. Pass in 0 for maxCacheSize for no limit.
func New(buffer []byte, maxCacheSize uint64) (*Document, error) {
	if !bytes.HasPrefix(buffer, []byte("%PDF")) {
		return nil, errs.New("only PDF documents are supported")
	}
	var d Document
	d.ctx = C.fz_new_context_imp(nil, nil, C.size_t(maxCacheSize), C.version)
	if d.ctx == nil {
		return nil, errs.New("unable to allocate PDF context")
	}
	C.fz_register_document_handlers(d.ctx)
	d.data = (*C.uchar)(C.CBytes(buffer))
	if d.data == nil {
		d.Release()
		return nil, errs.New("unable to allocate internal buffer")
	}
	stream := C.wrapped_fz_open_memory(d.ctx, d.data, C.size_t(len(buffer)))
	if stream == nil {
		d.Release()
		return nil, errs.New("unable to allocate internal stream")
	}
	d.doc = C.fz_open_document_with_stream(d.ctx, C.pdfMimeType, stream)
	C.fz_drop_stream(d.ctx, stream)
	if d.doc == nil {
		d.Release()
		return nil, errs.New("unable to open PDF")
	}
	if C.fz_needs_password(d.ctx, d.doc) != 0 {
		d.Release()
		return nil, errs.New("unable to open password-protected PDF")
	}
	runtime.SetFinalizer(&d, func(obj *Document) { obj.Release() })
	return &d, nil
}

// PageCount returns total number of pages in the document.
func (d *Document) PageCount() int {
	return int(C.fz_count_pages(d.ctx, d.doc))
}

// RenderPage renders the specified page at the requested dpi. If search is not empty, then the bounding boxes of up to
// maxHits matching text on the page will be returned.
func (d *Document) RenderPage(pageNumber int, dpi float32, search string, maxHits int) (draw.Image, []geom32.Rect, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	pageCount := d.PageCount()
	if pageNumber >= d.PageCount() {
		return nil, nil, errs.Newf("page number %d is out of range (0-%d)", pageNumber, pageCount)
	}
	displayList := C.wrapped_fz_new_display_list_from_page_number(d.ctx, d.doc, C.int(pageNumber))
	defer C.fz_drop_display_list(d.ctx, displayList)
	scale := dpi / 72
	ctm := C.fz_scale(C.float(scale), C.float(scale))
	pixmap := C.wrapped_fz_new_pixmap_from_display_list(d.ctx, displayList, ctm, C.fz_device_rgb(d.ctx), 1)
	pixels := C.fz_pixmap_samples(d.ctx, pixmap)
	if pixels == nil {
		return nil, nil, errs.New("unable to obtain pixels")
	}
	var boxes []geom32.Rect
	if search != "" {
		searchText := C.CString(search)
		defer C.free(unsafe.Pointer(searchText))
		hitBoxes := make([]C.fz_quad, maxHits)
		hits := C.wrapped_fz_search_display_list(d.ctx, displayList, searchText, (*C.fz_quad)(unsafe.Pointer(&hitBoxes[0])), C.int(len(hitBoxes)))
		if hits > 0 {
			boxes = make([]geom32.Rect, hits)
			for i := range boxes {
				boxes[i].X = float32(hitBoxes[i].ul.x) * scale
				boxes[i].Y = float32(hitBoxes[i].ul.y) * scale
				boxes[i].Width = (1 + float32(hitBoxes[i].lr.x-hitBoxes[i].ul.x)) * scale
				boxes[i].Height = (1 + float32(hitBoxes[i].lr.y-hitBoxes[i].ul.y)) * scale
			}
		}
	}
	return &image.NRGBA{
		Pix:    C.GoBytes(unsafe.Pointer(pixels), C.int(4*pixmap.w*pixmap.h)),
		Stride: int(pixmap.stride),
		Rect:   image.Rect(0, 0, int(pixmap.w), int(pixmap.h)),
	}, boxes, nil
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
