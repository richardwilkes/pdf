package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"time"

	"github.com/richardwilkes/pdf"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xmath/geom32"
)

func main() {
	startedAt := time.Now()
	data, err := os.ReadFile("/Users/rich/Documents/Gaming/Gaming - Shared/Roleplaying Games/GURPS/4th Edition/Basic Set/Basic Set - Characters.pdf")
	elapsed := time.Since(startedAt)
	jot.FatalIfErr(err)
	fmt.Printf("Loaded %d bytes from PDF file in %v\n", len(data), elapsed)
	startedAt = time.Now()
	var doc *pdf.Document
	doc, err = pdf.New(data, 0)
	elapsed = time.Since(startedAt)
	jot.FatalIfErr(err)
	defer doc.Release()
	pageCount := doc.PageCount()
	fmt.Printf("Loaded %d pages in %v\n", pageCount, elapsed)

	// Extract pages as images
	jot.FatalIfErr(os.RemoveAll("out"))
	jot.FatalIfErr(os.MkdirAll("out", 0o750))
	var maxElapsed time.Duration
	if pageCount > 3 {
		pageCount = 3
	}
	for n := 0; n < pageCount; n++ {
		startedAt = time.Now()
		var img draw.Image
		var boxes []geom32.Rect
		img, boxes, err = doc.RenderPage(n, 100, "GURPS", 20)
		jot.FatalIfErr(err)
		elapsed = time.Since(startedAt)
		fmt.Printf("Page %d converted in %v\n", n+1, elapsed)
		marker := image.NewUniform(color.NRGBA{
			R: 255,
			G: 255,
			B: 0,
			A: 96,
		})
		for _, box := range boxes {
			draw.Draw(img, image.Rect(int(box.X), int(box.Y), int(box.Right()), int(box.Bottom())), marker, image.Point{}, draw.Over)
		}
		if elapsed > maxElapsed {
			maxElapsed = elapsed
		}
		writePNG(img, n)
	}
	fmt.Printf("Maximum time to load a page: %v\n", maxElapsed)
	fmt.Println("Done!")
}

func writePNG(img draw.Image, pageNumber int) {
	f, err := os.Create(fmt.Sprintf("out/img-%d.png", pageNumber))
	jot.FatalIfErr(err)
	jot.FatalIfErr(png.Encode(f, img))
	jot.FatalIfErr(f.Close())
}
