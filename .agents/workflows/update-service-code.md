---
description: How to correctly update and edit the watermark-remover service code
---

# Watermark Remover Service - AI Coding Guidelines

You have been tasked with updating the codebase of the Go-based watermark-remover service. Before making any code changes, **you MUST follow the rules in this document**.

## 1. Project Architecture
- **API Layer** (`internal/api/handlers.go`): Handles Gin HTTP requests for `/url` and `/upload` endpoints. Extracts the image byte array and delegates to `imgutil`.
- **Detector** (`internal/imgutil/detector.go`): Finds ENCAR watermarks focusing explicitly on the bottom-right corner zone using Otsu thresholding and connected components.
- **Remover** (`internal/imgutil/remover.go`): Uses OpenCV to create masks from bounding boxes, applies `gocv.Inpaint`, and overlays a custom `logo.png` after blurring the target background region.

## 2. strict GoCV / OpenCV Rules
### A. Inpainting Algorithm
Due to compilation issues with standard `gocv` wrappers, **DO NOT use `gocv.InpaintTelea`**.
*Correct approach*: Use the constant `gocv.Telea` as the flag in `gocv.Inpaint`:
```go
gocv.Inpaint(img, mask, &result, 30, gocv.Telea)
```

### B. Color Space Management (CRITICAL)
OpenCV (`gocv`) operates in BGR format by default, whereas the Go `image` package operates in RGB / RGBA.
Whenever you convert from a `gocv.Mat` to standard Go `image.Image` (e.g. `gocv.Mat.ToImage()`), or vice-versa, you **must explicitly convert the color space** to avoid swapping red and blue channels.
*Example of correct usage:*
```go
// 1. Convert BGR to RGB
resultRGB := gocv.NewMat()
gocv.CvtColor(result, &resultRGB, gocv.ColorBGRToRGB)

// 2. Safely create Go image
resultImg, _ := resultRGB.ToImage()

// 3. Draw operations (using 'image' package) ...

// 4. Convert back RGB to BGR before returning/saving
finalRGBA, _ := gocv.ImageToMatRGBA(dst)
finalBGR := gocv.NewMat()
gocv.CvtColor(finalRGBA, &finalBGR, gocv.ColorRGBAToBGR)
```

## 3. Image Sizing & Output Rules
- The replacement logo should be dynamically scaled to 18% of the image's width using `imaging.Lanczos`.
- When blurring the base layer behind the logo, maintain a local radius/padding constraint so standard image details are not degraded (`ksize=15`, padding=10).
- The detection zone is fixed to `x > 0.75 * width`, `y > 0.85 * height`. Do not expand this without explicitly being requested to, as it will trigger false positives on bright headlights or plates.
- The service exports images predominantly in JPEG and WebP. Rely on standard native formatting flags for encoding.

## 4. Work in Go (Not Python)
This project was migrated from Python. Do not introduce Python-based libraries, dependencies or pip commands. The repository is strictly a containerized Go service (`CGO_ENABLED=1`).
