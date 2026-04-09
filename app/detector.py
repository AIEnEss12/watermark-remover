import cv2
import numpy as np

def detect_watermark(img: np.ndarray) -> tuple[int, int, int, int]:
    """
    Detect ENCAR watermark bounding box.
    Returns (x1, y1, x2, y2).
    """
    h, w = img.shape[:2]
    
    # Method 1 - fixed zone (last 15% height, 25% width right corner)
    x1_fixed = int(w * 0.75)
    y1_fixed = int(h * 0.85)
    x2_fixed = w
    y2_fixed = h
    
    # Cut crop for the zone
    crop = img[y1_fixed:y2_fixed, x1_fixed:x2_fixed]
    
    # Method 2 - contrast/brightness text detection (optional refine)
    gray = cv2.cvtColor(crop, cv2.COLOR_BGR2GRAY)
    
    # threshold to find bright pixels: white/grey text
    _, thresh = cv2.threshold(gray, 200, 255, cv2.THRESH_BINARY)
    
    # Find connected components
    num_labels, labels, stats, centroids = cv2.connectedComponentsWithStats(thresh, connectivity=8)
    
    # Determine the bounding rect around all detected 'text-like' components
    # Start with empty values
    min_x, min_y, max_x, max_y = w, h, 0, 0
    found_any = False
    
    for i in range(1, num_labels): # skip background 0
        area = stats[i, cv2.CC_STAT_AREA]
        # filter out noise and very big blobs that can't be text
        if 50 < area < 10000:
            found_any = True
            left = stats[i, cv2.CC_STAT_LEFT]
            top = stats[i, cv2.CC_STAT_TOP]
            width = stats[i, cv2.CC_STAT_WIDTH]
            height = stats[i, cv2.CC_STAT_HEIGHT]
            
            min_x = min(min_x, left)
            min_y = min(min_y, top)
            max_x = max(max_x, left + width)
            max_y = max(max_y, top + height)
            
    if found_any:
        # Translate local crop coordinates back to global
        final_x1 = x1_fixed + min_x
        final_y1 = y1_fixed + min_y
        final_x2 = x1_fixed + max_x
        final_y2 = y1_fixed + max_y
        
        # Add 10px padding
        final_x1 = max(0, final_x1 - 10)
        final_y1 = max(0, final_y1 - 10)
        final_x2 = min(w, final_x2 + 10)
        final_y2 = min(h, final_y2 + 10)
        return final_x1, final_y1, final_x2, final_y2
    
    # If refining fails, return the fixed box as fallback fallback
    return x1_fixed, y1_fixed, x2_fixed, y2_fixed
