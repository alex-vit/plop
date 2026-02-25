//go:build ignore

package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	// Generate ICO with 16 and 32 px sizes.
	sizes := []int{16, 32}
	var pngs [][]byte
	for _, size := range sizes {
		var buf bytes.Buffer
		png.Encode(&buf, drawIcon(size))
		pngs = append(pngs, buf.Bytes())
	}

	ico, err := os.Create("icon.ico")
	if err != nil {
		panic(err)
	}
	defer ico.Close()
	ico.Write(buildICO(sizes, pngs))

	// Also write a 22px PNG for macOS template icon.
	f, err := os.Create("icon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, drawIcon(22))
}

func drawIcon(size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	black := color.NRGBA{0, 0, 0, 255}

	set := func(x, y int) {
		if x >= 0 && x < size && y >= 0 && y < size {
			img.Set(x, y, black)
		}
	}

	// Scale factor relative to the original 22px design.
	s := float64(size) / 22.0

	si := func(v int) int { return int(float64(v)*s + 0.5) }

	// Up arrow (left side, centered at x=7)
	for y := si(7); y <= si(17); y++ {
		set(si(6), y)
		set(si(7), y)
	}
	set(si(6), si(6))
	set(si(7), si(6))
	set(si(5), si(7))
	set(si(8), si(7))
	set(si(4), si(8))
	set(si(9), si(8))
	set(si(3), si(9))
	set(si(10), si(9))

	// Down arrow (right side, centered at x=14)
	for y := si(4); y <= si(14); y++ {
		set(si(13), y)
		set(si(14), y)
	}
	set(si(13), si(15))
	set(si(14), si(15))
	set(si(12), si(14))
	set(si(15), si(14))
	set(si(11), si(13))
	set(si(16), si(13))
	set(si(10), si(12))
	set(si(17), si(12))

	return img
}

// buildICO assembles an ICO file from PNG-encoded images.
func buildICO(sizes []int, pngs [][]byte) []byte {
	n := len(sizes)
	dataOffset := 6 + n*16

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, [3]uint16{0, 1, uint16(n)})

	offset := uint32(dataOffset)
	for i, size := range sizes {
		w := uint8(size)
		if size >= 256 {
			w = 0
		}
		buf.Write([]byte{w, w, 0, 0})
		binary.Write(&buf, binary.LittleEndian, uint16(1))
		binary.Write(&buf, binary.LittleEndian, uint16(32))
		binary.Write(&buf, binary.LittleEndian, uint32(len(pngs[i])))
		binary.Write(&buf, binary.LittleEndian, offset)
		offset += uint32(len(pngs[i]))
	}

	for _, p := range pngs {
		buf.Write(p)
	}
	return buf.Bytes()
}
