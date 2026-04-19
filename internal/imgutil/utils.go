package imgutil

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif" // Register GIF decoder
	_ "image/png" // Register PNG decoder

	_ "golang.org/x/image/webp" // Register WebP decoder

	_ "github.com/gen2brain/avif" // Register AVIF decoder
	"gocv.io/x/gocv"
)

// DecodeImage decodes an image and ensures it is in a visually correct BGR format.
func DecodeImage(data []byte) (gocv.Mat, error) {
	// 1. Try OpenCV first
	raw, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err == nil && !raw.Empty() {
		// OpenCV's IMDecode reads ENCAR/NomadO source JPEGs as-is, which means
		// their R and B channels are physically swapped in the Mat.
		// We correct this immediately so the rest of the pipeline works with
		// visually correct pixels.
		img := gocv.NewMat()
		gocv.CvtColor(raw, &img, gocv.ColorBGRToRGB)
		raw.Close()
		return img, nil
	}

	// 2. Fallback: Go stdlib image.Decode
	reader := bytes.NewReader(data)
	src, _, err := image.Decode(reader)
	if err != nil {
		return gocv.Mat{}, fmt.Errorf("decoding failed: %v", err)
	}

	matRGBA, err := gocv.ImageToMatRGBA(src)
	if err != nil {
		return gocv.Mat{}, fmt.Errorf("failed to convert decoded image to Mat: %v", err)
	}
	defer matRGBA.Close()

	// image.Decode produces a visually correct image. We convert it to BGR
	// order so the Mat is in standard OpenCV BGR format.
	mat := gocv.NewMat()
	gocv.CvtColor(matRGBA, &mat, gocv.ColorRGBAToBGR)
	return mat, nil
}

// ProcessingTag is appended to files that have already been processed to avoid redundant work.
const ProcessingTag = "\x00PROCESSED_BY_WATERMARK_REMOVER\x00"

// HasProcessingTag returns true if the data ends with our custom processing tag.
func HasProcessingTag(data []byte) bool {
	return bytes.HasSuffix(data, []byte(ProcessingTag))
}

// EncodeToWebP encodes a gocv.Mat (BGR) to WebP bytes.
// It assumes the input Mat is a standard, visually correct BGR Mat.
func EncodeToWebP(img gocv.Mat, quality float32) ([]byte, error) {
	buf, err := gocv.IMEncodeWithParams(".webp", img, []int{gocv.IMWriteWebpQuality, int(quality)})
	if err != nil {
		return nil, err
	}
	defer buf.Close()

	data := buf.GetBytes()
	res := append(data, []byte(ProcessingTag)...)
	return res, nil
}

// EncodeToJPEG encodes a gocv.Mat (BGR) to JPEG bytes.
// It assumes the input Mat is a standard, visually correct BGR Mat.
func EncodeToJPEG(img gocv.Mat, quality int) ([]byte, error) {
	buf, err := gocv.IMEncodeWithParams(".jpg", img, []int{gocv.IMWriteJpegQuality, quality})
	if err != nil {
		return nil, err
	}
	defer buf.Close()

	data := buf.GetBytes()
	res := append(data, []byte(ProcessingTag)...)
	return res, nil
}
