package art

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif" // register GIF decoder
	"image/jpeg"
	"image/png"

	_ "golang.org/x/image/webp" // register WEBP decoder

	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/transform"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

const (
	maxArtSize        = 512
	minBackgroundSize = 32
	imageQuality      = 70
)

func Normalize(raw []byte) ([]byte, error) {
	img, err := decode(raw)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, square(img, maxArtSize), &jpeg.Options{Quality: imageQuality}); err != nil {
		return nil, fmt.Errorf("art: encode artwork: %w", err)
	}
	return buf.Bytes(), nil
}

func Background(raw []byte, frac float64) (*gdk.Texture, error) {
	size := int(maxArtSize - float64(maxArtSize-minBackgroundSize)*frac)

	img, err := decode(raw)
	if err != nil {
		return nil, err
	}
	img = square(img, size)

	if radius := frac * float64(size); radius > 0 {
		img = blur.Gaussian(img, radius)
	}
	return encodeTexture(img)
}

func Thumbnail(raw []byte, size int) (*gdk.Texture, error) {
	img, err := decode(raw)
	if err != nil {
		return nil, err
	}
	return encodeTexture(square(img, size))
}

func square(img image.Image, size int) image.Image {
	img = transform.Crop(img, centeredSquare(img.Bounds()))
	if img.Bounds().Dx() > size {
		img = transform.Resize(img, size, size, transform.Linear)
	}
	return img
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

func decode(raw []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("art: decode image: %w", err)
	}
	return img, nil
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
