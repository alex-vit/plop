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

type statusIconData struct {
	png []byte
	ico []byte
}

var generatedStatusIcons = map[StatusLight]statusIconData{
	StatusLightSynced:    buildStatusIcon(2), // Bottom lamp.
	StatusLightSyncing:   buildStatusIcon(1), // Middle lamp.
	StatusLightAttention: buildStatusIcon(0), // Top lamp.
}

func BytesForStatusLight(state StatusLight) (pngData, icoData []byte) {
	if iconData, ok := generatedStatusIcons[state]; ok {
		return iconData.png, iconData.ico
	}
	return Data, DataICO
}

func buildStatusIcon(activeLamp int) statusIconData {
	icoSizes := []int{32}
	pngs := make([][]byte, 0, len(icoSizes))
	for _, size := range icoSizes {
		pngs = append(pngs, encodePNG(drawTrafficLight(size, activeLamp)))
	}

	return statusIconData{
		png: encodePNG(drawTrafficLight(24, activeLamp)),
		ico: buildICO(icoSizes, pngs),
	}
}

func drawTrafficLight(size, activeLamp int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	black := color.NRGBA{0, 0, 0, 255}

	set := func(x, y int) {
		if x >= 0 && x < size && y >= 0 && y < size {
			img.Set(x, y, black)
		}
	}

	radius := size / 7
	if radius < 2 {
		radius = 2
	}

	margin := (size - 6*radius) / 4
	if margin < 1 {
		margin = 1
	}

	cx := size / 2
	step := 2*radius + margin
	centers := [3]int{
		margin + radius,
		margin + radius + step,
		margin + radius + step*2,
	}

	left := cx - radius - 2
	right := cx + radius + 2
	top := centers[0] - radius - 2
	bottom := centers[2] + radius + 2
	drawRectOutline(set, left, top, right, bottom)

	for i, cy := range centers {
		drawCircle(set, cx, cy, radius, false)
		if i == activeLamp {
			drawCircle(set, cx, cy, radius-1, true)
		}
	}

	return img
}

func drawRectOutline(set func(x, y int), left, top, right, bottom int) {
	for x := left; x <= right; x++ {
		set(x, top)
		set(x, bottom)
	}
	for y := top; y <= bottom; y++ {
		set(left, y)
		set(right, y)
	}
}

func drawCircle(set func(x, y int), cx, cy, r int, fill bool) {
	if r <= 0 {
		return
	}

	outer := r * r
	inner := (r - 1) * (r - 1)
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist2 := dx*dx + dy*dy
			if fill {
				if dist2 <= outer {
					set(cx+dx, cy+dy)
				}
				continue
			}
			if dist2 <= outer && dist2 >= inner {
				set(cx+dx, cy+dy)
			}
		}
	}
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
