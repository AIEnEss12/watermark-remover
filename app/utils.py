import cv2
import numpy as np

def encode_image(img: np.ndarray) -> bytes:
    """Encode OpenCV image to JPEG bytes."""
    _, encoded = cv2.imencode('.jpg', img)
    return encoded.tobytes()

def decode_image(img_bytes: bytes) -> np.ndarray:
    """Decode image bytes to OpenCV image (BGR) using Pillow."""
    try:
        from io import BytesIO
        from PIL import Image
        img_pil = Image.open(BytesIO(img_bytes)).convert('RGB')
        open_cv_image = np.array(img_pil)
        # Convert RGB to BGR
        return open_cv_image[:, :, ::-1].copy()
    except Exception:
        return None
