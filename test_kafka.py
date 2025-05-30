import requests
import json
from uuid import uuid4

# URL сервиса
url = "http://localhost:8000/items"

# Данные для создания товара
data = {
    "operation_type": 3,  # CREATE
    "item": {
        "id": str(uuid4()),  # Генерируем случайный UUID
        "name": "Test Product",
        "description": "This is a test product",
        "price": 100.0,
        "category": "Electronics",
        "stock": 10
    }
}

# Отправка POST запроса
response = requests.post(url, json=data)

# Вывод результата
print(f"Status Code: {response.status_code}")
print(f"Response: {response.text}")
print(f"Created item ID: {data['item']['id']}")  # Сохраняем ID для последующих операций 