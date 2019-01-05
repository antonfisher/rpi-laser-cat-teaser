package editor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
)

// Editor to draw on jpeg images
type Editor struct {
	Image *image.RGBA
}

// Width of editing image
func (e *Editor) Width() int {
	return e.Image.Bounds().Size().X
}

// Height of editing image
func (e *Editor) Height() int {
	return e.Image.Bounds().Size().Y
}

// JpegBytes returns edited image bytes
func (e *Editor) JpegBytes(quality int) []byte {
	w := new(bytes.Buffer)

	jpeg.Encode(w, e.Image, &jpeg.Options{
		Quality: quality,
	})

	return w.Bytes()
}

// DrawCrosshead on image
func (e *Editor) DrawCrosshead(x, y, crossheadSize, crossheadStrokeWidth int) {
	imageSize := e.Image.Bounds().Size()
	cBlue := &color.RGBA{0, 0, 255, 255}

	for i := 0; i < crossheadSize; i++ {
		if i > crossheadSize/3 && i < crossheadSize*2/3 {
			continue
		}
		for w := -crossheadStrokeWidth / 2; w < crossheadStrokeWidth/2; w++ {
			xd := x - crossheadSize/2 + i
			if xd >= 0 && xd < imageSize.X && y >= 0 && y < imageSize.Y {
				e.Image.Set(xd, y+w, cBlue)
			}
			yd := y - crossheadSize/2 + i
			if x >= 0 && x < imageSize.X && yd >= 0 && yd < imageSize.Y {
				e.Image.Set(x+w, yd, cBlue)
			}
		}
	}
}

// DrawRect on image
func (e *Editor) DrawRect(x0, y0, x1, y1 int) {
	imageSize := e.Image.Bounds().Size()
	cRed := &color.RGBA{255, 0, 0, 255}

	for x := x0; x <= x1 && x <= imageSize.X; x++ {
		e.Image.Set(x, y0, cRed)
		e.Image.Set(x, y1, cRed)
	}
	for y := y0; y <= y1 && y <= imageSize.Y; y++ {
		e.Image.Set(x0, y, cRed)
		e.Image.Set(x1, y, cRed)
	}
}

// DiffGreen diff with another image based green channel
func (e *Editor) DiffGreen(img image.Image, threshold uint32) [][]int {
	cDiff := &color.RGBA{255, 255, 0, 255}
	w := e.Width()
	h := e.Height()

	diffArray := make([][]int, w)
	for i := range diffArray {
		diffArray[i] = make([]int, h)
	}

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			//TODO  what channel to use, detector still self-triggers by laser dot
			_, g1, _, _ := e.Image.At(x, y).RGBA()
			_, g2, _, _ := img.At(x, y).RGBA()
			if colorDiff(g1, g2) > threshold {
				e.Image.Set(x, y, cDiff)
				diffArray[x][y] = 1
				//} else {
				//	e.Image.Set(x, y, &color.RGBA{0, uint8(g2), 0, 255})
			}
		}
	}

	return diffArray
}

// Clone creates a of currently edited image
func (e *Editor) Clone() image.Image {
	size := e.Image.Bounds().Size()

	// create empty image with the same size (to be able to draw)
	newImage := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))

	// copy source image to the new one
	draw.Draw(newImage, e.Image.Bounds(), e.Image, image.ZP, draw.Src)

	return newImage
}

func colorDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

// NewEditorFromJpegBytes creates new jpeg image from []byte
func NewEditorFromJpegBytes(imageBytes []byte) (*Editor, error) {
	imageReader := bytes.NewReader(imageBytes)

	jpegImage, err := jpeg.Decode(imageReader)
	if err != nil {
		return nil, fmt.Errorf("[Image Editor] cannot decode jpeg, error: %v", err)
	}

	size := jpegImage.Bounds().Size()

	// create empty image with the same size (to be able to draw)
	newImage := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))

	// copy source image to the new one
	draw.Draw(newImage, jpegImage.Bounds(), jpegImage, image.ZP, draw.Src)

	return &Editor{Image: newImage}, nil
}
