package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/username/watermark-remover/internal/imgutil"
	"gocv.io/x/gocv"
)

func main() {
	// Download baboon
	resp, err := http.Get("https://raw.githubusercontent.com/opencv/opencv/master/samples/data/baboon.jpg")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	// Decode
	img, _ := gocv.IMDecode(data, gocv.IMReadColor)
	defer img.Close()

	// Get center pixel in BGR (OpenCV convention: channels per row)
	stride := img.Cols() * 3
	row := 256
	col := 256
	bVal := img.GetUCharAt(row, col*3+0)
	gVal := img.GetUCharAt(row, col*3+1)
	rVal := img.GetUCharAt(row, col*3+2)
	fmt.Printf("After IMDecode at (%d,%d) BGR: B=%d G=%d R=%d\n", row, col, bVal, gVal, rVal)
	_ = stride

	// Encode direct (no CvtColor)
	bufDirect, _ := gocv.IMEncodeWithParams(".webp", img, []int{gocv.IMWriteWebpQuality, 80})
	defer bufDirect.Close()
	os.WriteFile("/app/output/direct.webp", bufDirect.GetBytes(), 0644)
	fmt.Println("Wrote direct.webp")

	// Encode with BGR→RGB
	imgRGB := gocv.NewMat()
	defer imgRGB.Close()
	gocv.CvtColor(img, &imgRGB, gocv.ColorBGRToRGB)
	bufSwapped, _ := gocv.IMEncodeWithParams(".webp", imgRGB, []int{gocv.IMWriteWebpQuality, 80})
	defer bufSwapped.Close()
	os.WriteFile("/app/output/swapped.webp", bufSwapped.GetBytes(), 0644)
	fmt.Println("Wrote swapped.webp")

	// Use imgutil pipeline
	imgDecoded, _ := imgutil.DecodeImage(data)
	defer imgDecoded.Close()

	// Test with webp encode
	webpBytes, _ := imgutil.EncodeToWebP(imgDecoded, 80)
	os.WriteFile("/app/output/pipeline.webp", webpBytes, 0644)
	fmt.Println("Wrote pipeline.webp")
}
