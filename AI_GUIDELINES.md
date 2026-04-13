# AI Coding Guidelines for Watermark Remover

This document outlines the rules for any AI coding assistant (`geminie`, `cursor`, `copilot`, etc.) operating on this repository.

## Overview
This is a Go-based service designed to remove specific watermarks (ENCAR) from car images and overlay a standard replacement logo.

## Core Rules

1. **NO Python**: This project was ported from Python to Go. Do NOT reintroduce Python scripts for image processing in the core logic. Standardize around Go (`CGO_ENABLED=1`, `gocv`).
2. **GoCV Compilation Quirk**: NEVER use `gocv.InpaintTelea`. It will not compile. Use `gocv.Inpaint` and pass `gocv.Telea` as the flag argument.
3. **Color Channels (RGB vs BGR)**:
   - OpenCV (`gocv`) defaults to **BGR**.
   - Go's `image` module defaults to **RGB/RGBA**.
   - **Crucial**: Whenever you convert from `gocv.Mat` to standard `image.Image` (`mat.ToImage()`), you MUST explicitly convert it from BGR to RGB first. Example: `gocv.CvtColor(result, &resultRGB, gocv.ColorBGRToRGB)`. Reverse this process (`gocv.ColorRGBAToBGR`) before returning the modified matrix back to OpenCV operations. Failing this will corrupt image colors.
4. **Detection Zone Constraint**: In `detector.go`, detection is scoped strictly to the bottom-right corner to prevent false positives (like headlights). Do not expand this scope without express permission.
5. **Blur Constants**: Use `ksize=15` for the halo blur. Do not scale this aggressively, as it will leak into the car structure.
6. **Rescale Logic**: Ensure logo resizing is dynamic based on standard bounds (18% of the image width) and uses `imaging.Lanczos`.

**Any modifications to `internal/imgutil/` must carefully respect the memory bounds and `defer .Close()` rules of CGO pointers in `gocv`.**
