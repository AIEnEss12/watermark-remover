import cv2
import sys
from pathlib import Path
from io import BytesIO
from PIL import Image
import numpy as np
import httpx

# Add project root to path for imports
sys.path.append(str(Path(__file__).parent.parent))

from app.detector import detect_watermark
from app.remover import remove_watermark

def main():
    url = "https://ci.encar.com/carpicture08/pic4008/40083772_024.jpg?impolicy=heightRate&rh=1080&cw=1920&ch=1080&cg=Center&wtmk=https://ci.encar.com/wt_mark/w_mark_04.png&t=20260223153827"
    logo_svg = "logo.svg"
    
    output_dir = Path("output")
    output_dir.mkdir(exist_ok=True)
    
    result_path = output_dir / "result_svg.jpg"
    
    print(f"Downloading image from {url}")
    try:
        with httpx.Client(follow_redirects=True) as client:
            resp = client.get(url, headers={'User-Agent': 'Mozilla/5.0'})
            resp.raise_for_status()
            image_bytes = resp.content
    except Exception as e:
        print(f"Failed to fetch {url}: {e}")
        return

    try:
        img_pil = Image.open(BytesIO(image_bytes)).convert("RGB")
        img = np.array(img_pil)[:, :, ::-1].copy()
    except Exception as e:
        print(f"Failed to decode image: {e}")
        return
    
    cv2.imwrite(str(output_dir / "original_avif.jpg"), img)
    
    bbox = detect_watermark(img)
    print(f"Detected watermark bbox: {bbox}")
    
    # Debug image (bbox viz)
    debug_img = img.copy()
    x1, y1, x2, y2 = bbox
    cv2.rectangle(debug_img, (x1, y1), (x2, y2), (0, 0, 255), 3)
    cv2.imwrite(str(output_dir / "debug_avif.jpg"), debug_img)
    
    # Run remover (which now places result in the corner)
    result = remove_watermark(img, bbox, logo_path=logo_svg)
    cv2.imwrite(str(result_path), result)
    
    print(f"Done. Result saved to {result_path}")

if __name__ == "__main__":
    main()
