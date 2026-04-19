package imgutil

import (
	"image"

	"gocv.io/x/gocv"
)

// DetectWatermark detects all ENCAR branding bounding boxes.
func DetectWatermark(img gocv.Mat) []image.Rectangle {
	rows := img.Rows()
	cols := img.Cols()

	// Define zones to check (Strictly Bottom-Right Corner only)
	zones := []image.Rectangle{
		// 1. Digital Watermark (Bottom-Right, clamped to bottom 8% to match logo height)
		image.Rect(int(float64(cols)*0.75), int(float64(rows)*0.92), cols, rows),
	}

	var results []image.Rectangle

	for idx, zone := range zones {
		// Ensure zone is within bounds
		zoneSlice := zone.Intersect(image.Rect(0, 0, cols, rows))
		if zoneSlice.Empty() {
			continue
		}

		crop := img.Region(zoneSlice)

		// Detection logic
		gray := gocv.NewMat()
		gocv.CvtColor(crop, &gray, gocv.ColorBGRToGray)

		threshVal := float32(220)
		threshType := gocv.ThresholdBinary
		if idx == 0 {
			threshVal = 0 // Otsu will calculate it
			threshType = gocv.ThresholdBinary | gocv.ThresholdOtsu
		}

		thresh := gocv.NewMat()
		gocv.Threshold(gray, &thresh, threshVal, 255, threshType)

		labels := gocv.NewMat()
		stats := gocv.NewMat()
		centroids := gocv.NewMat()
		numLabels := gocv.ConnectedComponentsWithStats(thresh, &labels, &stats, &centroids)

		for i := 1; i < numLabels; i++ {
			area := stats.GetIntAt(i, 4)
			left := int(stats.GetIntAt(i, 0))
			top := int(stats.GetIntAt(i, 1))
			width := int(stats.GetIntAt(i, 2))
			height := int(stats.GetIntAt(i, 3))

			isCandidate := false
			if idx == 0 {
				// Corner zone: ultra-sensitive, minimal filtering
				isCandidate = area > 10 && area < 15000
			} else {
				// Background zones: conservative, strict filtering (should not be reached now)
				ratio := float64(width) / float64(height)
				isCandidate = area > 100 && area < 20000 && (ratio > 0.5 && ratio < 4.0)
			}

			if isCandidate {
				res := image.Rect(zoneSlice.Min.X+left, zoneSlice.Min.Y+top, zoneSlice.Min.X+left+width, zoneSlice.Min.Y+top+height)
				// Padding
				res.Min.X = max(0, res.Min.X-3)
				res.Min.Y = max(0, res.Min.Y-3)
				res.Max.X = min(cols, res.Max.X+3)
				res.Max.Y = min(rows, res.Max.Y+3)
				results = append(results, res)
			}
		}

		// Cleanup
		crop.Close()
		gray.Close()
		thresh.Close()
		labels.Close()
		stats.Close()
		centroids.Close()
	}

	return results
}

