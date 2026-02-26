FROM python:3.10-slim

# Системные зависимости для OpenCV
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    && rm -rf /var/lib/apt/lists/*

# 1. Сначала ставим NumPy и SciPy из стандартного репозитория (PyPI)
RUN pip install --no-cache-dir "numpy<2.0.0" "scipy<1.13.0"

# 2. Затем ставим PyTorch (CPU-версию) из официального индекса PyTorch
RUN pip install --no-cache-dir torch torchvision --index-url https://download.pytorch.org/whl/cpu

# 3. Устанавливаем iopaint
RUN pip install --no-cache-dir iopaint

ENV PORT=8080

CMD ["sh", "-c", "iopaint start --model=lama --device=cpu --host=0.0.0.0 --port=${PORT}"]