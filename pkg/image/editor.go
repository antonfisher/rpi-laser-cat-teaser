package image

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

	for i := 0; i < crossheadSize; i++ {
		if i > crossheadSize/3 && i < crossheadSize*2/3 {
			continue
		}
		for w := -crossheadStrokeWidth / 2; w < crossheadStrokeWidth/2; w++ {
			xd := x - crossheadSize/2 + i
			if xd >= 0 && xd < imageSize.X && y >= 0 && y < imageSize.Y {
				e.Image.Set(xd, y+w, &color.RGBA{255, 0, 0, 255})
			}
			yd := y - crossheadSize/2 + i
			if x >= 0 && x < imageSize.X && yd >= 0 && yd < imageSize.Y {
				e.Image.Set(x+w, yd, &color.RGBA{255, 0, 0, 155})
			}
		}
	}
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
