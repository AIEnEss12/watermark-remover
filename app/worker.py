import os
import redis
from rq import Worker, Queue, Connection
import httpx
import logging
from .detector import detect_watermark
from .remover import remove_watermark
from .utils import decode_image, encode_image

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Redis configuration
REDIS_HOST = os.getenv('REDIS_HOST', 'redis')
REDIS_PORT = os.getenv('REDIS_PORT', '6379')

def process_watermark_task(image_url: str):
    """
    Task function to be executed by workers.
    """
    logger.info(f"Worker processing: {image_url}")
    try:
        # Use sync client for worker task
        with httpx.Client(follow_redirects=True, timeout=30) as client:
            resp = client.get(image_url)
            resp.raise_for_status()
            image_bytes = resp.content
    except Exception as e:
        logger.error(f"Failed to fetch image: {e}")
        return {"status": "error", "error": str(e), "url": image_url}

    img = decode_image(image_bytes)
    if img is None:
        return {"status": "error", "error": "Decode failed", "url": image_url}
    
    bbox = detect_watermark(img)
    result = remove_watermark(img, bbox)
    res_bytes = encode_image(result)
    
    # NOTE: In production, upload res_bytes to S3/Minio and return URL
    # For demonstration, we simply return success and metadata
    logger.info(f"Finished processing: {image_url}")
    return {
        "status": "success", 
        "url": image_url, 
        "result_size": len(res_bytes)
    }

if __name__ == '__main__':
    # Connect to Redis and start working
    try:
        conn = redis.Redis(host=REDIS_HOST, port=REDIS_PORT)
        with Connection(conn):
            worker = Worker(Queue('watermark_tasks'))
            worker.work()
    except Exception as e:
        logger.error(f"Worker failed to start: {e}")
