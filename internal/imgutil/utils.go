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

// DecodeImage decodes an image from bytes into a standard gocv.Mat (BGR).
func DecodeImage(data []byte) (gocv.Mat, error) {
	// 1. Try OpenCV first (decodes directly to BGR)
	img, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err == nil && !img.Empty() {
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

// EncodeToWebP encodes a gocv.Mat to WebP bytes.
// It applies a BGR→RGB swap to correct for the R/B channel order issue
// found in ENCAR/NomadO source images.
func EncodeToWebP(img gocv.Mat, quality float32) ([]byte, error) {
	imgRGB := gocv.NewMat()
	defer imgRGB.Close()
	gocv.CvtColor(img, &imgRGB, gocv.ColorBGRToRGB)

	buf, err := gocv.IMEncodeWithParams(".webp", imgRGB, []int{gocv.IMWriteWebpQuality, int(quality)})
	if err != nil {
		return nil, err
	}
	defer buf.Close()

	data := buf.GetBytes()
	res := append(data, []byte(ProcessingTag)...)
	return res, nil
}

// EncodeToJPEG encodes a gocv.Mat to JPEG bytes.
// It applies a BGR→RGB swap to correct for the R/B channel order issue
// found in ENCAR/NomadO source images.
func EncodeToJPEG(img gocv.Mat, quality int) ([]byte, error) {
	imgRGB := gocv.NewMat()
	defer imgRGB.Close()
	gocv.CvtColor(img, &imgRGB, gocv.ColorBGRToRGB)

	buf, err := gocv.IMEncodeWithParams(".jpg", imgRGB, []int{gocv.IMWriteJpegQuality, quality})
	if err != nil {
		return nil, err
	}
	defer buf.Close()

	data := buf.GetBytes()
	res := append(data, []byte(ProcessingTag)...)
	return res, nil
}
