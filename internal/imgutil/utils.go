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

// DecodeImage decodes an image from bytes into a gocv.Mat in BGR format.
//
// Note: Source JPEG files from ENCAR S3 may have R and B channels swapped
// at the source. We do NOT attempt to correct this here — the images are
// passed through as-is, and OpenCV's encode functions handle BGR→output correctly.
func DecodeImage(data []byte) (gocv.Mat, error) {
	// 1. Try OpenCV first (fastest, handles JPEG/PNG/BMP/etc.)
	img, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err == nil && !img.Empty() {
		return img, nil
	}

	// 2. Fallback: Go stdlib image.Decode (handles WebP/AVIF via registered decoders).
	//    image.Decode returns RGB-ordered pixels. We convert RGBA→BGR so the Mat
	//    is in standard OpenCV BGR order for consistent processing and encoding.
	reader := bytes.NewReader(data)
	src, _, err := image.Decode(reader)
	if err != nil {
		return gocv.Mat{}, fmt.Errorf("decoding failed: %v (data length: %d)", err, len(data))
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

// EncodeToWebP encodes a BGR gocv.Mat to WebP bytes using the native OpenCV encoder.
// OpenCV's IMEncodeWithParams correctly interprets the BGR input for encoding.
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

// EncodeToJPEG encodes a BGR gocv.Mat to JPEG bytes using the native OpenCV encoder.
// OpenCV's IMEncodeWithParams correctly interprets the BGR input for encoding.
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
