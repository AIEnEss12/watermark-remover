package imgutil

import (
	"image"
	"sync"

	"github.com/disintegration/imaging"
)

var (
	logoImageCache sync.Map // key: int (width) -> image.Image (pre-scaled Go image)
	logoImageOnce  sync.Once
	fullLogoRef    image.Image
)

// GetScaledLogoImage returns a cached, high-quality scaled logo as a Go image.Image.
func GetScaledLogoImage(logoPath string, targetW int) (image.Image, int, int, error) {
	// Initialize once
	var loadErr error
	logoImageOnce.Do(func() {
		fullLogoRef, loadErr = imaging.Open(logoPath)
	})
	if loadErr != nil {
		return nil, 0, 0, loadErr
	}
	if fullLogoRef == nil {
		return nil, 0, 0, nil
	}

	// Check cache
	if val, ok := logoImageCache.Load(targetW); ok {
		img := val.(image.Image)
		return img, img.Bounds().Dx(), img.Bounds().Dy(), nil
	}

	// Scale using high-quality Lanczos
	logoBounds := fullLogoRef.Bounds()
	scale := float64(targetW) / float64(logoBounds.Dx())
	targetH := int(float64(logoBounds.Dy()) * scale)
	resized := imaging.Resize(fullLogoRef, targetW, targetH, imaging.Lanczos)

	logoImageCache.Store(targetW, resized)
	return resized, targetW, targetH, nil
}
