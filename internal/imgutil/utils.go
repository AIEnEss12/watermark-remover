package imgutil

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF decoder
	"image/jpeg"
	_ "image/png"  // Register PNG decoder
	_ "golang.org/x/image/webp" // Register WebP decoder

	"github.com/chai2010/webp"
	_ "github.com/gen2brain/avif" // Register AVIF decoder
	"gocv.io/x/gocv"
)

// DecodeImage decodes an image from bytes into a gocv.Mat.
func DecodeImage(data []byte) (gocv.Mat, error) {
	// 1. Try OpenCV first (fastest)
	img, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err == nil && !img.Empty() {
		return img, nil
	}

	// 2. Fallback to image.Decode (handles formats like WebP/AVIF if registered)
	reader := bytes.NewReader(data)
	src, _, err := image.Decode(reader)
	if err != nil {
		return gocv.Mat{}, fmt.Errorf("decoding failed: %v (data length: %d)", err, len(data))
	}

	// Convert image.Image to gocv.Mat (RGBA first, then BGR for consistency)
	matRGBA, err := gocv.ImageToMatRGBA(src)
	if err != nil {
		return gocv.Mat{}, fmt.Errorf("failed to convert decoded image to Mat: %v", err)
	}
	defer matRGBA.Close()

	mat := gocv.NewMat()
	gocv.CvtColor(matRGBA, &mat, gocv.ColorRGBAToBGR)

	return mat, nil
}

// EncodeToWebP encodes a gocv.Mat to WebP bytes.
func EncodeToWebP(img gocv.Mat, quality float32) ([]byte, error) {
	// Convert BGR to RGB for Go encoder
	imgRGB := gocv.NewMat()
	defer imgRGB.Close()
	gocv.CvtColor(img, &imgRGB, gocv.ColorBGRToRGB)

	goImg, err := imgRGB.ToImage()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = webp.Encode(&buf, goImg, &webp.Options{Lossless: false, Quality: quality})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeToJPEG encodes a gocv.Mat to JPEG bytes.
func EncodeToJPEG(img gocv.Mat, quality int) ([]byte, error) {
	// Convert BGR to RGB for Go encoder
	imgRGB := gocv.NewMat()
	defer imgRGB.Close()
	gocv.CvtColor(img, &imgRGB, gocv.ColorBGRToRGB)

	goImg, err := imgRGB.ToImage()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, goImg, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
