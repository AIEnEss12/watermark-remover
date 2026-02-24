FROM python:3.10-slim

# 1. Устанавливаем актуальные библиотеки для работы с графикой (OpenCV)
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    && rm -rf /var/lib/apt/lists/*

# 2. Устанавливаем torch (CPU версия)
RUN pip install --no-cache-dir torch torchvision --index-url https://download.pytorch.org/whl/cpu

# 3. Устанавливаем iopaint
RUN pip install --no-cache-dir iopaint

# Настройка порта для Railway
ENV PORT=8080

# 4. Используем JSON-формат для CMD (как советовал ворнинг), чтобы корректно ловить сигналы ОС
CMD ["sh", "-c", "iopaint start --model=lama --device=cpu --host=0.0.0.0 --port=${PORT}"]