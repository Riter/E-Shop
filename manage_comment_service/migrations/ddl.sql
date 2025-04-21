CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,         -- ID товара из внешней БД
    user_id BIGINT NOT NULL,            -- ID пользователя, который оставил комментарий
    content TEXT NOT NULL,              -- Текст комментария
    rating INT CHECK (rating >= 1 AND rating <= 5), -- Оценка от 1 до 5
    created_at TIMESTAMP DEFAULT NOW(), -- Когда оставлен комментарий
    updated_at TIMESTAMP DEFAULT NOW()  -- Когда обновлён комментарий
);
