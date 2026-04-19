package imgutil

import (
	"image"
	"sync"

	"github.com/disintegration/imaging"
)

var (
	logoCache     sync.Map // key: int (width) -> image.Image (pre-scaled)
	logoOnce      sync.Once
	fullLogoImage image.Image
)

// GetScaledLogoImage returns a cached, scaled logo as a Go image.Image.
// targetW is the desired width; aspectH=0 means preserve aspect ratio from source.
// The caller should treat the returned image as read-only.
func GetScaledLogoImage(logoPath string, targetW int) (image.Image, int, int, error) {
	// Initialize the source logo once
	var loadErr error
	logoOnce.Do(func() {
		fullLogoImage, loadErr = imaging.Open(logoPath)
	})
	if loadErr != nil {
		return nil, 0, 0, loadErr
	}
	if fullLogoImage == nil {
		// logo file didn't exist or wasn't loaded — no-op
		return nil, 0, 0, nil
	}

	// Check cache
	if val, ok := logoCache.Load(targetW); ok {
		scaled := val.(image.Image)
		return scaled, scaled.Bounds().Dx(), scaled.Bounds().Dy(), nil
	}

	// Scale preserving aspect ratio
	logoBounds := fullLogoImage.Bounds()
	scale := float64(targetW) / float64(logoBounds.Dx())
	targetH := int(float64(logoBounds.Dy()) * scale)

	resized := imaging.Resize(fullLogoImage, targetW, targetH, imaging.Lanczos)
	logoCache.Store(targetW, image.Image(resized))
	return resized, targetW, targetH, nil
}
