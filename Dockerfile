FROM python:3.10-slim

# Системные зависимости для OpenCV
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    && rm -rf /var/lib/apt/lists/*

# 1. Устанавливаем совместимые версии NumPy и SciPy до установки iopaint
# NumPy < 2.0.0 критически важен для старых нейронок
RUN pip install --no-cache-dir \
    "numpy<2.0.0" \
    "scipy<1.13.0" \
    torch torchvision --index-url https://download.pytorch.org/whl/cpu

# 2. Устанавливаем iopaint
# Он подтянет остальные зависимости, но не будет трогать уже установленные numpy/scipy
RUN pip install --no-cache-dir iopaint

ENV PORT=8080

# Запуск с использованием переменной окружения Railway
CMD ["sh", "-c", "iopaint start --model=lama --device=cpu --host=0.0.0.0 --port=${PORT}"]