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

type blobStyle struct {
	body    color.NRGBA
	outline color.NRGBA
	grid    [16]string // 16x16 bitmap: '.'=transparent, 'O'=outline, 'B'=body, 'E'=eye
}

var dark = color.NRGBA{0x1E, 0x1E, 0x1E, 0xFF}

var statusStyles = map[StatusLight]blobStyle{
	StatusLightSynced: {
		body:    color.NRGBA{0x22, 0xC5, 0x5E, 0xFF}, // green
		outline: color.NRGBA{0x15, 0x80, 0x3D, 0xFF},
		grid: [16]string{
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
		},
	},
	StatusLightSyncing: {
		body:    color.NRGBA{0xF5, 0x9E, 0x0B, 0xFF}, // amber
		outline: color.NRGBA{0xB4, 0x53, 0x09, 0xFF},
		grid: [16]string{
			".....OOOOOO.....",
			"...OOBBBBBBOO...",
			"..OBBBBBBBBBBO..",
			".OBBBBBBBBBBBBO.",
			".OBBBBBBBBBBBBO.",
			"OBBEEEBBBBEEEBBO", // o_o top arc
			"OBEBBBEBBEBBBEBO", // o_o sides
			"OBEBBBEBBEBBBEBO",
			"OBEBBBEBBEBBBEBO",
			"OBBEEEBBBBEEEBBO", // o_o bottom arc
			"OBBBBBBBBBBBBBBO",
			".OBBBBBBBBBBBBO.",
			".OBBBBBBBBBBBBO.",
			"..OBBBBBBBBBBO..",
			"...OOBBBBBBOO...",
			".....OOOOOO.....",
		},
	},
	StatusLightAttention: {
		body:    color.NRGBA{0xEF, 0x44, 0x44, 0xFF}, // red
		outline: color.NRGBA{0xDC, 0x26, 0x26, 0xFF},
		grid: [16]string{
			".....OOOOOO.....",
			"...OOBBBBBBOO...",
			"..OBBBBBBBBBBO..",
			".OBBBBBBBBBBBBO.",
			".OBBBBBBBBBBBBO.",
			"OBEBBBEBBEBBBEBO", // >_< corners
			"OBBEBEBBBBEBEBBO", // >_< inner
			"OBBBEBBBBBBEBBBO", // >_< center
			"OBBEBEBBBBEBEBBO", // >_< inner
			"OBEBBBEBBEBBBEBO", // >_< corners
			"OBBBBBBBBBBBBBBO",
			".OBBBBBBBBBBBBO.",
			".OBBBBBBBBBBBBO.",
			"..OBBBBBBBBBBO..",
			"...OOBBBBBBOO...",
			".....OOOOOO.....",
		},
	},
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
	style := statusStyles[state]

	icoSizes := []int{16, 32}
	icoPngs := make([][]byte, 0, len(icoSizes))
	for _, size := range icoSizes {
		icoPngs = append(icoPngs, encodePNG(renderBlob(style, size)))
	}

	return statusIconData{
		png: encodePNG(renderBlob(style, 22)),
		ico: buildICO(icoSizes, icoPngs),
	}
}

// renderBlob draws the blob icon at the given size by nearest-neighbor
// scaling from the 16x16 source grid.
func renderBlob(style blobStyle, size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	for y := range size {
		srcY := y * 16 / size
		row := style.grid[srcY]
		for x := range size {
			srcX := x * 16 / size
			switch row[srcX] {
			case 'O':
				img.SetNRGBA(x, y, style.outline)
			case 'B':
				img.SetNRGBA(x, y, style.body)
			case 'E':
				img.SetNRGBA(x, y, dark)
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
