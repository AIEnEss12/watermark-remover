# Watermark Remover Go Service - Claude Coding Guidelines

You are an AI programming assistant working on the watermark-remover service. 
When making modifications to the service, you MUST adhere to the following rules to prevent regression of hard-learned architecture patterns and GoCV compiler errors.

## 1. Important GoCV Usage Conventions
- **Inpaint**: **Never** call `gocv.InpaintTelea`. It will cause a compilation error (`undefined: gocv.InpaintTelea`). Always use `gocv.Inpaint(src, mask, dest, radius, gocv.Telea)`.
- **Color Spaces (CRITICAL)**: `gocv` operates in `BGR` memory format. Standard Go libraries (`image/color`, `image/draw`) operate in `RGB/RGBA`. **You must always explicitly use `gocv.CvtColor` with `gocv.ColorBGRToRGB` before calling `.ToImage()`**. Do the reverse (`gocv.ColorRGBAToBGR`) after converting back to `gocv.Mat`. Failure to perform these steps leads to severe color-swap bugs (blue watermarks instead of red).

## 2. Repository Structure
- `internal/api/`: Gin HTTP layers. Handles requests, validation, and JSON error responses.
- `internal/imgutil/`: Processing logic. Contains `detector.go` (bounding box detection with Otsu thresholding), `remover.go` (Inpaint + Logo overlay), and `utils.go`.
- `tests/` & `run_test.sh`: Contains scripts to validate behavior.
- `logo.png`: Replaces the found watermark.

## 3. Architecture Logic Constraints
- **Detector Zone**: The detection algorithm is limited to `width * 0.75` and `height * 0.85` (Bottom-Right box) deliberately. Do not broaden this zone without instruction, to minimize false positives elsewhere in the image.
- **Blur & Resizing**: Background blur behind the logo must have a conservative `ksize=15`. 
- **Migration note**: The codebase was migrated strictly to Go from Python. Avoid re-introducing Python processing steps.

Always review these constraints before writing any code that mutates `imgutil` or updates the image processing pipeline.
