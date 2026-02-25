//go:build ignore

package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

const size = 22

func main() {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	black := color.NRGBA{0, 0, 0, 255}

	set := func(x, y int) { img.Set(x, y, black) }

	// Up arrow (left side, centered at x=7)
	// Shaft: 2px wide, y=7..17
	for y := 7; y <= 17; y++ {
		set(6, y)
		set(7, y)
	}
	// Arrowhead pointing up
	set(6, 6)
	set(7, 6)
	set(5, 7)
	set(8, 7)
	set(4, 8)
	set(9, 8)
	set(3, 9)
	set(10, 9)

	// Down arrow (right side, centered at x=14)
	// Shaft: 2px wide, y=4..14
	for y := 4; y <= 14; y++ {
		set(13, y)
		set(14, y)
	}
	// Arrowhead pointing down
	set(13, 15)
	set(14, 15)
	set(12, 14)
	set(15, 14)
	set(11, 13)
	set(16, 13)
	set(10, 12)
	set(17, 12)

	f, err := os.Create("icon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
