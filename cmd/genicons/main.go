package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

var (
	midGold = color.RGBA{0xff, 0xd6, 0x0a, 0xff} // base gold
	hiGold  = color.RGBA{0xff, 0xf0, 0x6e, 0xff} // highlight (top)
	loGold  = color.RGBA{0xb5, 0x8a, 0x00, 0xff} // shadow (bottom)
)

func lerpU8(a, b int, t float64) uint8 {
	v := float64(a) + t*float64(b-a)
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// inRoundedRect tests whether pixel (x, y) falls inside the rounded rectangle.
func inRoundedRect(x, y, x1, y1, x2, y2, r int) bool {
	if x < x1 || x >= x2 || y < y1 || y >= y2 {
		return false
	}
	r = min(r, (x2-x1)/2, (y2-y1)/2)
	cx := clampi(x, x1+r, x2-r)
	cy := clampi(y, y1+r, y2-r)
	dx, dy := float64(x-cx), float64(y-cy)
	return dx*dx+dy*dy <= float64(r*r)
}

// drawElement draws one barbell piece with a 3-tone vertical gradient.
func drawElement(img *image.RGBA, x1, y1, x2, y2, r int) {
	h := float64(y2 - y1)
	for px := x1; px < x2; px++ {
		for py := y1; py < y2; py++ {
			if !inRoundedRect(px, py, x1, y1, x2, y2, r) {
				continue
			}
			t := float64(py-y1) / h // 0 = top, 1 = bottom
			var c color.RGBA
			switch {
			case t < 0.28: // highlight → mid
				tt := t / 0.28
				c = color.RGBA{
					lerpU8(int(hiGold.R), int(midGold.R), tt),
					lerpU8(int(hiGold.G), int(midGold.G), tt),
					lerpU8(int(hiGold.B), int(midGold.B), tt),
					0xff,
				}
			case t > 0.78: // mid → shadow
				tt := (t - 0.78) / 0.22
				c = color.RGBA{
					lerpU8(int(midGold.R), int(loGold.R), tt),
					lerpU8(int(midGold.G), int(loGold.G), tt),
					lerpU8(int(midGold.B), int(loGold.B), tt),
					0xff,
				}
			default:
				c = midGold
			}
			img.Set(px, py, c)
		}
	}
}

func sc(v, size int) int { return v * size / 192 }

func drawBarbell(img *image.RGBA, size int) {
	pr := max(1, size*10/192) // plate corner radius
	cr := max(1, size*5/192)  // collar corner radius
	br := max(0, size*4/192)  // bar corner radius

	el := func(x1, y1, x2, y2, r int) {
		drawElement(img, sc(x1, size), sc(y1, size), sc(x2, size), sc(y2, size), r)
	}
	el(16, 40, 58, 152, pr)   // left plate
	el(58, 62, 76, 130, cr)   // left collar
	el(76, 84, 116, 108, br)  // bar / grip
	el(116, 62, 134, 130, cr) // right collar
	el(134, 40, 176, 152, pr) // right plate
}

func genIcon(size int, path string) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Dark navy → near-black gradient background
	for y := 0; y < size; y++ {
		t := float64(y) / float64(size)
		bg := color.RGBA{
			lerpU8(0x22, 0x10, t),
			lerpU8(0x22, 0x10, t),
			lerpU8(0x38, 0x20, t),
			0xff,
		}
		for x := 0; x < size; x++ {
			img.Set(x, y, bg)
		}
	}

	drawBarbell(img, size)

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
	base := "web/static/icons/"
	genIcon(512, base+"icon-512.png")
	genIcon(192, base+"icon-192.png")
	genIcon(180, base+"apple-touch-icon.png")
	genIcon(32, base+"favicon-32.png")
}
