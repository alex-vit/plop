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

// 16x16 source grid for the default app icon (green ^_^ synced blob).
var grid16 = [16]string{
	".....OOOOOO.....",
	"...OOBBBBBBOO...",
	"..OBBBBBBBBBBO..",
	".OBBBBBBBBBBBBO.",
	".OBBBBBBBBBBBBO.",
	"OBBBBBBBBBBBBBBO",
	"OBBBEBBBBBBEBBBO", // ^_^ peaks
	"OBBEBEBBBBEBEBBO", // ^_^ mids
	"OBEBBBEBBEBBBEBO", // ^_^ feet
	"OBBBBBBBBBBBBBBO",
	"OBBBBBBBBBBBBBBO",
	".OBBBBBBBBBBBBO.",
	".OBBBBBBBBBBBBO.",
	"..OBBBBBBBBBBO..",
	"...OOBBBBBBOO...",
	".....OOOOOO.....",
}

var (
	bodyColor    = color.NRGBA{0x22, 0xC5, 0x5E, 0xFF} // green
	outlineColor = color.NRGBA{0x15, 0x80, 0x3D, 0xFF}
	eyeColor     = color.NRGBA{0x1E, 0x1E, 0x1E, 0xFF}
)

func main() {
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

	f, err := os.Create("icon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, drawIcon(22))
}

func drawIcon(size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	for y := range size {
		srcY := y * 16 / size
		row := grid16[srcY]
		for x := range size {
			srcX := x * 16 / size
			switch row[srcX] {
			case 'O':
				img.SetNRGBA(x, y, outlineColor)
			case 'B':
				img.SetNRGBA(x, y, bodyColor)
			case 'E':
				img.SetNRGBA(x, y, eyeColor)
			}
		}
	}
	return img
}

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
