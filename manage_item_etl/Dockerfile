# manage_item_etl/Dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Expose port for Prometheus metrics
EXPOSE 10667

# ETL-сервис просто запускает main.py
CMD ["python", "main.py"]
