// Command genicon draws the app icon and writes assets/icon.png and
// assets/icon.ico (PNG-compressed entries). Run with `go generate ./assets`.
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

var (
	bg     = color.RGBA{20, 20, 24, 255}
	yellow = color.RGBA{255, 204, 0, 255}
	white  = color.RGBA{235, 235, 235, 255}
)

// draw renders a clock face with a yellow ring and three list bars.
func draw(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	s := float64(size)
	c := s / 2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			d := math.Hypot(fx-c, fy-c)
			switch {
			case d > s*0.48:
				continue
			case d > s*0.40:
				img.Set(x, y, yellow)
			default:
				img.Set(x, y, bg)
			}
		}
	}
	// Three "build steps": short bar, long highlighted bar, bar.
	bar := func(y0, x0, x1 float64, col color.RGBA) {
		for y := int(s * y0); y < int(s*(y0+0.09)); y++ {
			for x := int(s * x0); x < int(s*x1); x++ {
				img.Set(x, y, col)
			}
		}
	}
	bar(0.28, 0.28, 0.62, white)
	bar(0.455, 0.28, 0.74, yellow)
	bar(0.63, 0.28, 0.56, white)
	return img
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	sizes := []int{16, 24, 32, 48, 64, 256}
	var pngs [][]byte
	for _, sz := range sizes {
		var b bytes.Buffer
		if err := png.Encode(&b, draw(sz)); err != nil {
			log.Fatal(err)
		}
		pngs = append(pngs, b.Bytes())
	}
	if err := os.WriteFile(filepath.Join(dir, "icon.png"), pngs[len(pngs)-1], 0o644); err != nil {
		log.Fatal(err)
	}

	// ICO: header, directory, then PNG payloads.
	var ico bytes.Buffer
	binary.Write(&ico, binary.LittleEndian, [3]uint16{0, 1, uint16(len(sizes))})
	offset := 6 + 16*len(sizes)
	for i, sz := range sizes {
		dim := uint8(sz)
		if sz >= 256 {
			dim = 0
		}
		binary.Write(&ico, binary.LittleEndian, struct {
			W, H, Colors, Reserved uint8
			Planes, BitCount       uint16
			Size, Offset           uint32
		}{dim, dim, 0, 0, 1, 32, uint32(len(pngs[i])), uint32(offset)})
		offset += len(pngs[i])
	}
	for _, p := range pngs {
		ico.Write(p)
	}
	if err := os.WriteFile(filepath.Join(dir, "icon.ico"), ico.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
}
