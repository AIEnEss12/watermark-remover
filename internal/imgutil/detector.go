package imgutil

import (
	"image"
	"gocv.io/x/gocv"
)

// DetectWatermark detects all ENCAR branding bounding boxes.
func DetectWatermark(img gocv.Mat) []image.Rectangle {
	rows := img.Rows()
	cols := img.Cols()

	// Define zones to check (Restored Wall and Plate zones)
	zones := []image.Rectangle{
		// 1. Digital Watermark (Bottom-Right)
		image.Rect(int(float64(cols)*0.75), int(float64(rows)*0.85), cols, rows),
		// 2. Wall Logo (Center-Right Background)
		image.Rect(int(float64(cols)*0.60), int(float64(rows)*0.10), int(float64(cols)*0.95), int(float64(rows)*0.50)),
		// 3. License Plate (Bottom-Center)
		image.Rect(int(float64(cols)*0.25), int(float64(rows)*0.60), int(float64(cols)*0.75), int(float64(rows)*0.95)),
	}

	var results []image.Rectangle

	for _, zone := range zones {
		// Ensure zone is within bounds
		zone = zone.Intersect(image.Rect(0, 0, cols, rows))
		if zone.Empty() {
			continue
		}

		crop := img.Region(zone)
		
		// Detection logic
		gray := gocv.NewMat()
		gocv.CvtColor(crop, &gray, gocv.ColorBGRToGray)
		
		thresh := gocv.NewMat()
		gocv.Threshold(gray, &thresh, 140, 255, gocv.ThresholdBinary) // Lowered to 140 for transparency
		
		labels := gocv.NewMat()
		stats := gocv.NewMat()
		centroids := gocv.NewMat()
		numLabels := gocv.ConnectedComponentsWithStats(thresh, &labels, &stats, &centroids)

		for i := 1; i < numLabels; i++ {
			area := stats.GetIntAt(i, 4)
			if area > 30 && area < 20000 {
				left := int(stats.GetIntAt(i, 0))
				top := int(stats.GetIntAt(i, 1))
				width := int(stats.GetIntAt(i, 2))
				height := int(stats.GetIntAt(i, 3))

				res := image.Rect(zone.Min.X+left, zone.Min.Y+top, zone.Min.X+left+width, zone.Min.Y+top+height)
				// Add 5px padding (smaller than before for precision)
				res.Min.X = max(0, res.Min.X-5)
				res.Min.Y = max(0, res.Min.Y-5)
				res.Max.X = min(cols, res.Max.X+5)
				res.Max.Y = min(rows, res.Max.Y+5)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
