FROM python:3.10-slim

# Системные зависимости
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Фиксируем версии для стабильности на CPU
RUN pip install --no-cache-dir "numpy<2.0.0" "scipy<1.13.0"
RUN pip install --no-cache-dir torch torchvision --index-url https://download.pytorch.org/whl/cpu

# Ставим IOPaint и EasyOCR
RUN pip install --no-cache-dir iopaint easyocr fastapi uvicorn python-multipart

# Копируем наш код
COPY app.py .

ENV PORT=8080

# При первом запуске EasyOCR и IOPaint скачают модели (~300-500MB)
CMD ["python", "app.py"]