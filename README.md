# ItemsCommentsMNG
This repository was created specially for containing a codebase for Items and Comments services

# Онлайн-магазин (Online Shop)

## Описание
Этот проект представляет собой сервер для онлайн-магазина, реализованный на языке Go с использованием фреймворка `go-chi`. Он поддерживает поиск товаров через REST API, взаимодействует с базой данных PostgreSQL и организован в соответствии с архитектурными принципами чистой архитектуры.

## Структура проекта
```
/online-shop
│── /cmd              # Входные точки приложения
│    ├── main.go      # Основной запуск сервера
│── /internal         # Основная логика приложения
│    ├── /db          # Подключение к БД
│    │    ├── db.go   # Настройки БД
│    ├── /models      # Структуры данных
│    │    ├── product.go  # Структура товара
│    ├── /repository  # Работа с БД (CRUD)
│    │    ├── product_repo.go
│    ├── /services    # Бизнес-логика (поиск товаров)
│    │    ├── search_service.go
│    ├── /handlers    # Обработчики HTTP-запросов
│    │    ├── search_handler.go
│── /config           # Файлы конфигурации
│    ├── config.go    # Чтение конфигов
│── /environment      # Секретные данные (НЕ пушим в Git)
│    ├── .env         # Переменные окружения (БД, API-ключи)
│── go.mod
│── go.sum
│── .gitignore  
```

## Установка и запуск

### 1. Клонирование репозитория
```sh
git clone https://github.com/MAPiryazev/ItemsCommentsMNG
cd online-shop
```

### 2. Настройка переменных окружения
Создайте файл `.env` в папке `environment/` и укажите параметры подключения к базе данных:
```ini
POSTGRES_HOST=localhost
и так далее
```

### 3. Запуск приложения
```sh
go run cmd/main.go
```
Сервер запустится на `http://localhost:51842`.

## API

### 1. Поиск товаров
**GET** `/search?q=название_товара`
#### Пример запроса:
```
http://localhost:51842/search?q=наушники
```
#### Пример ответа:
```json
[
  {
    "id": 1,
    "name": "iPhone 13",
    "price": 999.99,
    "description": "Новейший iPhone с A15 Bionic"
  },
  {
    "id": 2,
    "name": "Samsung Galaxy S21",
    "price": 799.99,
    "description": "Флагман Samsung с Exynos 2100"
  }
]
```

## Основные зависимости
- `github.com/go-chi/chi/v5` – роутер для обработки HTTP-запросов
- `github.com/lib/pq` – драйвер PostgreSQL
- `github.com/joho/godotenv` – загрузка переменных окружения



