package imgutil

import (
	"image"
	"image/color"

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

	logoBGRA := gocv.IMRead(logoPath, gocv.IMReadUnchanged)
	if logoBGRA.Empty() {
		return result, nil
	}
	defer logoBGRA.Close()

	targetW := int(float64(cols) * 0.18)
	logoCols := logoBGRA.Cols()
	logoRows := logoBGRA.Rows()
	scale := float64(targetW) / float64(logoCols)
	targetH := int(float64(logoRows) * scale)

	if targetW > 0 && targetH > 0 {
		resizedLogo := gocv.NewMat()
		defer resizedLogo.Close()
		gocv.Resize(logoBGRA, &resizedLogo, image.Pt(targetW, targetH), 0, 0, gocv.InterpolationLanczos4)

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

			// Overlay logo with Alpha channel support
			roi := result.Region(image.Rect(startX, startY, endX, endY))
			defer roi.Close()

			if resizedLogo.Channels() == 4 {
				// Split channels to get Alpha mask
				channels := gocv.Split(resizedLogo)
				defer func() {
					for i := range channels {
						channels[i].Close()
					}
				}()

				bgrLogo := gocv.NewMat()
				defer bgrLogo.Close()
				gocv.Merge(channels[0:3], &bgrLogo)

				// Use channel 3 (Alpha) as mask
				bgrLogo.CopyToWithMask(&roi, channels[3])
			} else {
				resizedLogo.CopyTo(&roi)
			}
		}
	}

	return result, nil
}
