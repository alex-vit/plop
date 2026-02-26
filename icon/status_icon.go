package icon

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

type StatusLight uint8

const (
	StatusLightSynced StatusLight = iota
	StatusLightSyncing
	StatusLightAttention
)

// Traffic-light colors for tray icon (Windows ICO + macOS PNG).
var statusColors = map[StatusLight]color.NRGBA{
	StatusLightSynced:    {R: 0x22, G: 0xC5, B: 0x5E, A: 0xFF}, // green
	StatusLightSyncing:   {R: 0xEA, G: 0xB3, B: 0x08, A: 0xFF}, // yellow/amber
	StatusLightAttention: {R: 0xEF, G: 0x44, B: 0x44, A: 0xFF}, // red
}

type statusIconData struct {
	png []byte
	ico []byte
}

var generatedStatusIcons = map[StatusLight]statusIconData{
	StatusLightSynced:    buildStatusIcon(StatusLightSynced),
	StatusLightSyncing:   buildStatusIcon(StatusLightSyncing),
	StatusLightAttention: buildStatusIcon(StatusLightAttention),
}

func BytesForStatusLight(state StatusLight) (pngData, icoData []byte) {
	if iconData, ok := generatedStatusIcons[state]; ok {
		return iconData.png, iconData.ico
	}
	return Data, DataICO
}

func buildStatusIcon(state StatusLight) statusIconData {
	c := statusColors[state]

	// ICO: colorful filled circle for Windows (16+32 px).
	icoSizes := []int{16, 32}
	icoPngs := make([][]byte, 0, len(icoSizes))
	for _, size := range icoSizes {
		icoPngs = append(icoPngs, encodePNG(drawColorCircle(size, c)))
	}

	// PNG: colorful filled circle for macOS menu bar (22px per convention).
	return statusIconData{
		png: encodePNG(drawColorCircle(22, c)),
		ico: buildICO(icoSizes, icoPngs),
	}
}

// drawColorCircle draws a filled circle in the given color, centered in a
// size×size image. Hard pixel edges, no anti-aliasing.
func drawColorCircle(size int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	center := float64(size) / 2
	radius := center - 1 // 1px margin

	for y := range size {
		for x := range size {
			px := float64(x) + 0.5
			py := float64(y) + 0.5
			dx := px - center
			dy := py - center
			if dx*dx+dy*dy <= radius*radius {
				img.SetNRGBA(x, y, c)
			}
		}
	}
	return img
}

func encodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func buildICO(sizes []int, pngs [][]byte) []byte {
	n := len(sizes)
	dataOffset := 6 + n*16

	var buf bytes.Buffer
	mustWriteLE(&buf, [3]uint16{0, 1, uint16(n)})

	offset := uint32(dataOffset)
	for i, size := range sizes {
		w := uint8(size)
		if size >= 256 {
			w = 0
		}
		_, _ = buf.Write([]byte{w, w, 0, 0})
		mustWriteLE(&buf, uint16(1))
		mustWriteLE(&buf, uint16(32))
		mustWriteLE(&buf, uint32(len(pngs[i])))
		mustWriteLE(&buf, offset)
		offset += uint32(len(pngs[i]))
	}

	for _, p := range pngs {
		_, _ = buf.Write(p)
	}
	return buf.Bytes()
}

func mustWriteLE(buf *bytes.Buffer, data any) {
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		panic(err)
	}
}
