from fastapi import FastAPI, HTTPException, UploadFile, File
from fastapi.responses import Response, HTMLResponse
from pydantic import BaseModel
import httpx
import logging
import os
import redis
from rq import Queue
from .detector import detect_watermark
from .remover import remove_watermark
from .utils import decode_image, encode_image
# Import worker task for queuing
from .worker import process_watermark_task

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Redis setup
REDIS_HOST = os.getenv('REDIS_HOST', 'redis')
REDIS_PORT = os.getenv('REDIS_PORT', '6379')
redis_conn = redis.Redis(host=REDIS_HOST, port=REDIS_PORT)
task_queue = Queue('watermark_tasks', connection=redis_conn)

app = FastAPI(title="ENCAR Watermark Remover Gateway")

class ImageRequest(BaseModel):
    image_url: str

@app.get("/")
def index():
    """Return a simple HTML form for testing."""
    return HTMLResponse(content="""
    <html>
        <head>
            <title>ENCAR Watermark Remover</title>
            <style>
                body { font-family: sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; background-color: #f0f0f0; }
                .container { background-color: white; padding: 2rem; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
                h1 { margin-top: 0; }
                form { display: flex; flex-direction: column; gap: 1rem; }
                input[type="file"] { border: 1px solid #ccc; padding: 0.5rem; border-radius: 4px; }
                button { background-color: #007bff; color: white; border: none; padding: 0.75rem; border-radius: 4px; cursor: pointer; }
                button:hover { background-color: #0056b3; }
            </style>
        </head>
        <body>
            <div class="container">
                <h1>ENCAR Watermark Remover</h1>
                <p>Upload a car image to remove the ENCAR watermark.</p>
                <form action="/remove/upload" method="post" enctype="multipart/form-data">
                    <input type="file" name="file" accept="image/*" required>
                    <button type="submit">Remove Watermark</button>
                </form>
            </div>
        </body>
    </html>
    """)

@app.get("/health")
def health_check():
    return {"status": "ok"}

@app.post("/remove")
async def remove_url(req: ImageRequest):
    logger.info(f"Downloading image from {req.image_url}")
    try:
        async with httpx.AsyncClient() as client:
            resp = await client.get(req.image_url)
            resp.raise_for_status()
            image_bytes = resp.content
    except Exception as e:
        logger.error(f"Failed to fetch image: {e}")
        raise HTTPException(status_code=422, detail="Failed to fetch image from URL")

    img = decode_image(image_bytes)
    if img is None:
        raise HTTPException(status_code=400, detail="Invalid image content")
    
    bbox = detect_watermark(img)
    result = remove_watermark(img, bbox)
    res_bytes = encode_image(result)
    
    return Response(content=res_bytes, media_type="image/jpeg")

@app.post("/remove/upload")
async def remove_upload(file: UploadFile = File(...)):
    image_bytes = await file.read()
    img = decode_image(image_bytes)
    if img is None:
        raise HTTPException(status_code=400, detail="Invalid image content")
        
    bbox = detect_watermark(img)
    result = remove_watermark(img, bbox)
    res_bytes = encode_image(result)
    
    return Response(content=res_bytes, media_type="image/jpeg")

@app.post("/queue")
async def queue_task(req: ImageRequest):
    """
    Queue the watermark removal task for asynchronous processing.
    The Go API can call this for high-volume scaling.
    """
    try:
        job = task_queue.enqueue(process_watermark_task, req.image_url)
        return {
            "status": "queued",
            "job_id": job.get_id(),
            "url": req.image_url
        }
    except Exception as e:
        logger.error(f"Failed to queue task: {e}")
        raise HTTPException(status_code=500, detail="Failed to queue task")

@app.get("/queue/{job_id}")
async def get_job_status(job_id: str):
    """Check the status and result of a queued task."""
    job = task_queue.fetch_job(job_id)
    if job is None:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return {
        "job_id": job.get_id(),
        "status": job.get_status(),
        "result": job.result
    }
