# manage_item_crud/Dockerfile
FROM python:3.11-slim

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем требования и ставим зависимости
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копируем остальное приложение
COPY . .

# Экспонируем порт FastAPI
EXPOSE 8000

# Запуск сервиса
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
