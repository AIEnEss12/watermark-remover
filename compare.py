import cv2
import numpy as np

orig = cv2.imread('output/copy.webp')
result = cv2.imread('output/test_result.webp')

if orig is None or result is None:
    print("Could not read images")
    exit(1)

diff = cv2.absdiff(orig, result)
print(f"Total differing pixels: {np.count_nonzero(diff.sum(axis=2) > 0)}")

# Find bounding box of differences
gray = cv2.cvtColor(diff, cv2.COLOR_BGR2GRAY)
_, thresh = cv2.threshold(gray, 1, 255, cv2.THRESH_BINARY)
contours, _ = cv2.findContours(thresh, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)

for c in contours:
    x, y, w, h = cv2.boundingRect(c)
    if w > 10 and h > 10:
        print(f"Significant difference at: x={x}, y={y}, w={w}, h={h}")

# check the bottom right specifically
h, w = orig.shape[:2]
print(f"Image shape: {w}x{h}")
