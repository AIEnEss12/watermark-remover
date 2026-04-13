package imgutil

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png" // Register PNG decoder
	"os"

	"github.com/disintegration/imaging"
	"gocv.io/x/gocv"
)

// RemoveWatermark removes the watermark from the image and overlays a new logo.
func RemoveWatermark(img gocv.Mat, bboxes []image.Rectangle, logoPath string) (gocv.Mat, error) {
	cols := img.Cols()
	rows := img.Rows()

	// 1. Create result image base
	result := gocv.NewMat()
	if len(bboxes) > 0 {
		mask := gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV8U)
		defer mask.Close()
		mask.SetTo(gocv.Scalar{Val1: 0, Val2: 0, Val3: 0, Val4: 0})

		for _, bbox := range bboxes {
			gocv.Rectangle(&mask, bbox, color.RGBA{R: 255, G: 255, B: 255, A: 255}, -1)
		}

		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(7, 7))
		defer kernel.Close()
		gocv.Dilate(mask, &mask, kernel)

		gocv.Inpaint(img, mask, &result, 30, gocv.Telea)
	} else {
		img.CopyTo(&result)
	}

	// 2. Handle Logo Overlay
	if logoPath == "" {
		logoPath = "logo.png"
	}
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		return result, nil
	}

	logoImg, err := imaging.Open(logoPath)
	if err != nil {
		fmt.Printf("Error opening logo %s: %v\n", logoPath, err)
		return result, nil
	}

	targetW := int(float64(cols) * 0.18)
	logoBounds := logoImg.Bounds()
	scale := float64(targetW) / float64(logoBounds.Dx())
	targetH := int(float64(logoBounds.Dy()) * scale)

	if targetW > 0 && targetH > 0 {
		resizedLogo := imaging.Resize(logoImg, targetW, targetH, imaging.Lanczos)
		padding := 5
		offX := cols - targetW - padding
		offY := rows - targetH - padding

		startX := max(0, offX)
		startY := max(0, offY)
		endX := min(cols, startX+targetW)
		endY := min(rows, startY+targetH)

		if endX > startX && endY > startY {
			// Apply blur halo ONLY if there was an original watermark
			if len(bboxes) > 0 {
				roi := result.Region(image.Rect(max(0, startX-2), max(0, startY), min(cols, endX+2), min(rows, endY+2)))
				defer roi.Close()
				gocv.GaussianBlur(roi, &roi, image.Pt(7, 7), 0, 0, gocv.BorderDefault)
			}

			// Overlay logo using Go native libraries for high quality blending
			// Convert BGR to RGB first to avoid color swap in ToImage()
			resultRGB := gocv.NewMat()
			defer resultRGB.Close()
			gocv.CvtColor(result, &resultRGB, gocv.ColorBGRToRGB)

			resultImg, err := resultRGB.ToImage()
			if err != nil {
				return result, err
			}

			dst := image.NewRGBA(resultImg.Bounds())
			draw.Draw(dst, dst.Bounds(), resultImg, image.Point{}, draw.Src)
			draw.Draw(dst, image.Rect(startX, startY, endX, endY), resizedLogo, image.Point{}, draw.Over)

			// Convert back to Mat
			finalRGBA, err := gocv.ImageToMatRGBA(dst)
			if err != nil {
				return result, err
			}
			defer finalRGBA.Close()

			finalBGR := gocv.NewMat()
			gocv.CvtColor(finalRGBA, &finalBGR, gocv.ColorRGBAToBGR)
			result.Close()
			return finalBGR, nil
		}
	}

	return result, nil
}
