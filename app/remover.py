import cv2
import numpy as np
import os
from io import BytesIO
from PIL import Image

def remove_watermark(img: np.ndarray, bbox: tuple[int, int, int, int], logo_path: str = "logo.svg") -> np.ndarray:
    """
    Remove watermark using cv2.inpaint and return the cleaned image.
    Then, if logo_path exists, overlay the new logo onto the same bbox.
    Supports SVG and PNG.
    """
    x1, y1, x2, y2 = bbox
    h, w = img.shape[:2]
    
    # Create mask for inpainting
    mask = np.zeros((h, w), dtype=np.uint8)
    cv2.rectangle(mask, (x1, y1), (x2, y2), 255, -1)
    
    # Dilate mask to capture text edges: kernel 3x3, iter = 2
    kernel = np.ones((3, 3), np.uint8)
    mask = cv2.dilate(mask, kernel, iterations=2)
    
    # Apply cv2.inpaint
    result = cv2.inpaint(img, mask, inpaintRadius=7, flags=cv2.INPAINT_TELEA)
    
    # Try finding logo
    current_dir = os.path.dirname(os.path.abspath(__file__))
    project_root = os.path.dirname(current_dir)
    
    logo_file = None
    # If using default or simplified name, check both .svg and .png in root
    base_logo = os.path.splitext(logo_path)[0]
    for ext in [".svg", ".png"]:
        p = os.path.join(project_root, f"{base_logo}{ext}")
        if os.path.exists(p):
            logo_file = p
            break
            
    # Fallback to direct path check if not found in root or if absolute path provided
    if not logo_file:
        if os.path.exists(logo_path):
            logo_file = logo_path
        else:
            p = os.path.normpath(os.path.join(project_root, logo_path))
            if os.path.exists(p):
                logo_file = p
                
    if not logo_file:
        return result

    try:
        if logo_file.lower().endswith('.svg'):
            import resvg_python
            with open(logo_file, 'r', encoding='utf-8') as f:
                svg_data = f.read()
            png_list = resvg_python.svg_to_png(svg_data)
            png_bytes = bytes(png_list)
            img_pil = Image.open(BytesIO(png_bytes)).convert('RGBA')
            # Convert RGBA to BGRA for OpenCV
            logo = np.array(img_pil)[:, :, [2, 1, 0, 3]].copy()
        else:
            logo = cv2.imread(logo_file, cv2.IMREAD_UNCHANGED)
            
        if logo is not None:
            target_w = x2 - x1
            target_h = y2 - y1
            
            logo_h, logo_w = logo.shape[:2]
            
            # Calculate scale to fit within bbox while maintaining aspect ratio
            scale = min(target_w / logo_w, target_h / logo_h)
            new_w = int(logo_w * scale)
            new_h = int(logo_h * scale)
            
            if new_w > 0 and new_h > 0:
                resized_logo = cv2.resize(logo, (new_w, new_h), interpolation=cv2.INTER_AREA)
                
                # Center the logo within the bbox
                off_x = x1 + (target_w - new_w) // 2
                off_y = y1 + (target_h - new_h) // 2
                
                # Overlay with alpha channel
                if resized_logo.shape[2] == 4:
                    alpha = resized_logo[:, :, 3] / 255.0
                    logo_rgb = resized_logo[:, :, :3]
                    
                    # Ensure indices are within bounds
                    off_y_end = min(off_y + new_h, h)
                    off_x_end = min(off_x + new_w, w)
                    actual_h = off_y_end - off_y
                    actual_w = off_x_end - off_x
                    
                    if actual_h > 0 and actual_w > 0:
                        roi = result[off_y:off_y_end, off_x:off_x_end]
                        alpha_crop = alpha[:actual_h, :actual_w, np.newaxis]
                        logo_rgb_crop = logo_rgb[:actual_h, :actual_w]
                        
                        result[off_y:off_y_end, off_x:off_x_end] = (
                            (1.0 - alpha_crop) * roi + alpha_crop * logo_rgb_crop
                        ).astype(np.uint8)
                else:
                    # No alpha, just replace
                    if len(resized_logo.shape) == 2:
                        resized_logo = cv2.cvtColor(resized_logo, cv2.COLOR_GRAY2BGR)
                    
                    off_y_end = min(off_y + new_h, h)
                    off_x_end = min(off_x + new_w, w)
                    actual_h = off_y_end - off_y
                    actual_w = off_x_end - off_x
                    
                    if actual_h > 0 and actual_w > 0:
                        result[off_y:off_y_end, off_x:off_x_end] = resized_logo[:actual_h, :actual_w, :3]
    except Exception as e:
        # Log error but return the inpainted result at least
        print(f"Error processing logo {logo_file}: {e}")
            
    return result
