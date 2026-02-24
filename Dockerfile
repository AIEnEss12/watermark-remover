FROM python:3.10-slim

# Устанавливаем системные зависимости для OpenCV
RUN apt-get update && apt-get install -y \
    libgl1-mesa-glx \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Ставим torch (CPU версия, чтобы образ был легче) и iopaint
RUN pip install --no-cache-dir torch torchvision --index-url https://download.pytorch.org/whl/cpu
RUN pip install --no-cache-dir iopaint

# Railway прокидывает порт через переменную $PORT
ENV PORT=8080

# Запуск: модель lama подгрузится автоматически при первом старте
CMD iopaint start --model=lama --device=cpu --host=0.0.0.0 --port=$PORT