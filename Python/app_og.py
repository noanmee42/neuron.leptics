# python/app.py

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
import logging

from claim_extractor import ClaimExtractor

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Создание FastAPI приложения
app = FastAPI(
    title="Hallucination Detector API",
    description="API для извлечения и верификации утверждений",
    version="1.0.0"
)

# Инициализация экстрактора (один раз при старте)
try:
    extractor = ClaimExtractor()
    logger.info("✓ ClaimExtractor успешно инициализирован")
except Exception as e:
    logger.error(f"❌ Ошибка инициализации: {e}")
    extractor = None


# === Pydantic модели для валидации ===

class ExtractClaimsRequest(BaseModel):
    """Запрос на извлечение утверждений"""
    text: str
    
    class Config:
        json_schema_extra = {
            "example": {
                "text": "Москва - столица России. Население более 12 миллионов."
            }
        }


class ExtractClaimsResponse(BaseModel):
    """Ответ с извлеченными утверждениями"""
    claims: List[str]
    
    class Config:
        json_schema_extra = {
            "example": {
                "claims": [
                    "Москва - столица России",
                    "Население Москвы более 12 миллионов"
                ]
            }
        }


# === Endpoints ===

@app.get("/health")
def health_check():
    """Проверка работоспособности API"""
    return {
        "status": "healthy",
        "extractor_ready": extractor is not None
    }


@app.post("/extract-claims", response_model=ExtractClaimsResponse)
def extract_claims_endpoint(request: ExtractClaimsRequest):
    """
    Извлекает утверждения из текста
    
    - **text**: Входной текст для анализа
    
    Возвращает список извлеченных утверждений
    """
    if extractor is None:
        raise HTTPException(
            status_code=500,
            detail="ClaimExtractor не инициализирован. Проверьте GEMINI_API_KEY."
        )
    
    if not request.text or not request.text.strip():
        raise HTTPException(
            status_code=400,
            detail="Текст не может быть пустым"
        )
    
    try:
        logger.info(f"Извлечение утверждений из текста ({len(request.text)} символов)")
        
        claims = extractor.extract(request.text)
        
        logger.info(f"✓ Извлечено {len(claims)} утверждений")
        
        return ExtractClaimsResponse(claims=claims)
        
    except Exception as e:
        logger.error(f"Ошибка при извлечении: {e}", exc_info=True)
        raise HTTPException(
            status_code=500,
            detail=f"Ошибка обработки: {str(e)}"
        )


# Запуск сервера (если запускаем напрямую)
if __name__ == "__main__":
    import uvicorn
    
    print("=" * 60)
    print("🚀 Запуск Hallucination Detector API")
    print("=" * 60)
    print(f"📍 URL: http://localhost:8000")
    print(f"📖 Docs: http://localhost:8000/docs")
    print("=" * 60)
    
    uvicorn.run(
        "app:app",
        host="0.0.0.0",
        port=8000,
        reload=True,  # Автоперезагрузка при изменении кода
        log_level="info"
    )