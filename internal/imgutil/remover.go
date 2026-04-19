package imgutil

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png" // Register PNG decoder

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

	targetW := int(float64(cols) * 0.18)

	// Use cached high-quality logo image
	resizedLogo, logoW, logoH, err := GetScaledLogoImage(logoPath, targetW)
	if err != nil {
		fmt.Printf("Error getting logo: %v\n", err)
		return result, nil
	}
	if resizedLogo == nil || logoW == 0 || logoH == 0 {
		return result, nil
	}

	padding := 10
	offX := cols - logoW - padding
	offY := rows - logoH - padding

	startX := offX
	if startX < 0 {
		startX = 0
	}
	startY := offY
	if startY < 0 {
		startY = 0
	}
	endX := startX + logoW
	if endX > cols {
		endX = cols
	}
	endY := startY + logoH
	if endY > rows {
		endY = rows
	}

	if endX > startX && endY > startY {
		// Apply subtle blur behind the logo if watermark was detected
		if len(bboxes) > 0 {
			blurX1 := startX - 3
			if blurX1 < 0 {
				blurX1 = 0
			}
			blurX2 := endX + 3
			if blurX2 > cols {
				blurX2 = cols
			}
			blurY1 := startY - 3
			if blurY1 < 0 {
				blurY1 = 0
			}
			blurY2 := endY + 3
			if blurY2 > rows {
				blurY2 = rows
			}
			roi := result.Region(image.Rect(blurX1, blurY1, blurX2, blurY2))
			defer roi.Close()
			gocv.GaussianBlur(roi, &roi, image.Pt(5, 5), 0, 0, gocv.BorderDefault)
		}

		// High-quality logo overlay using Go's draw package.
		// We convert the Mat to RGB first because Go's ToImage() reads channel 0 as R.
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

		// Convert back to BGR Mat for standard pipeline storage.
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

	return result, nil
}
