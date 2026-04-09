# ENCAR Watermark Remover

A fully containerized Python FastAPI service to remove ENCAR watermarks from car images without GPU boundaries (CPU only).
Utilizes OpenCV Inpainting inside a bounding box area.

## Running

1. Spin up the container:
```bash
docker-compose up --build
```
2. Call API:
```bash
curl -X POST http://localhost:8000/remove \
  -H "Content-Type: application/json" \
  -d '{"image_url": "https://img.nomadocars.com/unsafe/rs:fit:1200:0/plain/s3://sorted-tote-m1bwf4wyssq-2/carpicture03/pic4033/40336791_001.jpg"}' \
  --output result.jpg
```
