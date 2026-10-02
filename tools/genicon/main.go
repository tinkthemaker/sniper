// Command genicon draws the Sniper app icon and writes cmd/sniper/Icon.png.
package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
)

const size = 512

var (
	blue  = color.RGBA{0x2B, 0x6C, 0xB0, 0xFF}
	white = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)

	fill(img, blue, func(x, y float64) bool { return inRoundedRect(x, y, 0, 0, size, size, 96) })

	c := float64(size) / 2
	fill(img, white, func(x, y float64) bool {
		d := math.Hypot(x-c, y-c)
		return d >= 118 && d <= 138 // circle outline
	})
	const tickW, tickIn, tickOut = 20.0, 80.0, 190.0
	fill(img, white, func(x, y float64) bool {
		dx, dy := math.Abs(x-c), math.Abs(y-c)
		vertical := dx <= tickW/2 && dy >= tickIn && dy <= tickOut
		horizontal := dy <= tickW/2 && dx >= tickIn && dx <= tickOut
		return vertical || horizontal
	})

	f, err := os.Create("cmd/sniper/Icon.png")
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

// fill paints col on every pixel where in reports true, sampling a 3×3 grid per
// pixel for smooth edges.
func fill(img *image.RGBA, col color.RGBA, in func(x, y float64) bool) {
	const n = 3
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			hits := 0
			for sy := 0; sy < n; sy++ {
				for sx := 0; sx < n; sx++ {
					if in(float64(px)+(float64(sx)+0.5)/n, float64(py)+(float64(sy)+0.5)/n) {
						hits++
					}
				}
			}
			if hits == 0 {
				continue
			}
			a := float64(hits) / (n * n)
			bg := img.RGBAAt(px, py)
			img.SetRGBA(px, py, color.RGBA{
				R: mix(bg.R, col.R, a), G: mix(bg.G, col.G, a), B: mix(bg.B, col.B, a),
				A: uint8(math.Max(float64(bg.A), a*255)),
			})
		}
	}
}

func mix(bg, fg uint8, a float64) uint8 {
	return uint8(math.Round(float64(bg)*(1-a) + float64(fg)*a))
}

func inRoundedRect(x, y, x0, y0, x1, y1, r float64) bool {
	if x < x0 || x > x1 || y < y0 || y > y1 {
		return false
	}
	cx := math.Min(math.Max(x, x0+r), x1-r)
	cy := math.Min(math.Max(y, y0+r), y1-r)
	return math.Hypot(x-cx, y-cy) <= r
}
