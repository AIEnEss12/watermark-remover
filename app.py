import numpy as np
import cv2
import easyocr
from fastapi import FastAPI, UploadFile, File
from fastapi.responses import Response
from iopaint.model_manager import ModelManager
from iopaint.schema import InpaintRequest
import torch
import io
from PIL import Image

app = FastAPI()
# Инициализируем OCR для поиска текста (английский + корейский для Encar)
reader = easyocr.Reader(['en', 'ko'], gpu=False)
# Загружаем модель LaMa
model_manager = ModelManager(name="lama", device=torch.device("cpu"))

@app.post("/auto-erase")
async def auto_erase(file: UploadFile = File(...)):
    # 1. Читаем изображение
    contents = await file.read()
    nparr = np.frombuffer(contents, np.uint8)
    img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
    h, w, _ = img.shape

    # 2. Ищем текст через EasyOCR
    results = reader.readtext(img)
    
    # 3. Создаем черную маску
    mask = np.zeros((h, w), dtype=np.uint8)
    
    for (bbox, text, prob) in results:
        # Рисуем белые прямоугольники на маске там, где найден текст
        top_left = tuple(map(int, bbox[0]))
        bottom_right = tuple(map(int, bbox[2]))
        cv2.rectangle(mask, top_left, bottom_right, 255, -1)

    # 4. Запускаем Inpainting (LaMa)
    # Конвертируем для IOPaint
    img_rgb = cv2.cvtColor(img, cv2.COLOR_BGR2RGB)
    
    config = InpaintRequest(
        ldm_steps=20,
        no_half=True,
        return_mask=False
    )
    
    # Вызов модели
    result_img = model_manager.inference(img_rgb, mask, config)
    
    # 5. Возвращаем результат
    res, im_png = cv2.imencode(".jpg", cv2.cvtColor(result_img, cv2.COLOR_RGB2BGR))
    return Response(content=im_png.tobytes(), media_type="image/jpeg")

if __name__ == "__main__":
    import uvicorn
    import os
    port = int(os.environ.get("PORT", 8080))
    uvicorn.run(app, host="0.0.0.0", port=port)