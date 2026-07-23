package art

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	"image/png"

	_ "golang.org/x/image/webp" // register WEBP decoder

	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/transform"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

const (
	blurRadius  = 80.0
	maxBlurSize = 100
)

func Background(raw []byte) (*gdk.Texture, error) {
	img, err := decodeCapped(raw, maxBlurSize)
	if err != nil {
		return nil, err
	}

	blurred := blur.Gaussian(img, blurRadius)

	return encodeTexture(blurred)
}

func Thumbnail(raw []byte, size int) (*gdk.Texture, error) {
	img, err := decodeCapped(raw, 1<<20)
	if err != nil {
		return nil, err
	}

	cropped := transform.Crop(img, centeredSquare(img.Bounds()))
	small := transform.Resize(cropped, size, size, transform.Linear)

	return encodeTexture(small)
}

func centeredSquare(b image.Rectangle) image.Rectangle {
	side := b.Dx()
	if b.Dy() < side {
		side = b.Dy()
	}
	x0 := b.Min.X + (b.Dx()-side)/2
	y0 := b.Min.Y + (b.Dy()-side)/2
	return image.Rect(x0, y0, x0+side, y0+side)
}

func decodeCapped(raw []byte, maxDim int) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("art: decode image: %w", err)
	}

	b := img.Bounds()
	if b.Dx() <= maxDim && b.Dy() <= maxDim {
		return img, nil
	}

	w, h := b.Dx(), b.Dy()
	if w > h {
		h = h * maxDim / w
		w = maxDim
	} else {
		w = w * maxDim / h
		h = maxDim
	}
	return transform.Resize(img, w, h, transform.Linear), nil
}

func encodeTexture(img image.Image) (*gdk.Texture, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("art: encode texture: %w", err)
	}

	texture, err := gdk.NewTextureFromBytes(glib.NewBytes(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("art: build texture: %w", err)
	}
	return texture, nil
}
