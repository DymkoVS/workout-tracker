package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

var (
	yellow = color.RGBA{0xd7, 0xff, 0x1a, 0xff}
	black  = color.RGBA{0x00, 0x00, 0x00, 0xff}
	dark   = color.RGBA{0x0a, 0x0a, 0x0a, 0xff}
)

func fillRect(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			img.Set(x, y, c)
		}
	}
}

// drawBarbell draws a bold barbell icon scaled to given size (base coords: 192x192)
func drawBarbell(img *image.RGBA, size int, fg, bg color.RGBA) {
	s := float64(size)

	fill := func(x1, y1, x2, y2 float64) {
		fillRect(img,
			int(x1/192*s), int(y1/192*s),
			int(x2/192*s), int(y2/192*s),
			fg)
	}

	// Left plate
	fill(18, 52, 62, 140)
	// Left collar
	fill(62, 68, 78, 124)
	// Bar
	fill(78, 84, 114, 108)
	// Right collar
	fill(114, 68, 130, 124)
	// Right plate
	fill(130, 52, 174, 140)

	// Holes in plates (bg color to add detail)
	fillHole := func(x1, y1, x2, y2 float64) {
		fillRect(img,
			int(x1/192*s), int(y1/192*s),
			int(x2/192*s), int(y2/192*s),
			bg)
	}
	fillHole(28, 76, 52, 116)
	fillHole(140, 76, 164, 116)
}

func genIcon(size int, path string) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{yellow}, image.Point{}, draw.Src)
	drawBarbell(img, size, black, yellow)

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func main() {
	base := "/Users/vladimir/Documents/Claude Projects/workout-tracker/web/static/icons/"
	genIcon(512, base+"icon-512.png")
	genIcon(192, base+"icon-192.png")
	genIcon(180, base+"apple-touch-icon.png")
	genIcon(32, base+"favicon-32.png")
}
