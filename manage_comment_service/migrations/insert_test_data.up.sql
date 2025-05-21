-- Вставляем тестовые комментарии
INSERT INTO comments (user_id, product_id, content, rating, created_at, updated_at) VALUES
(1, 1, 'Отличный товар! Качество на высоте, доставка быстрая.', 5, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(2, 1, 'Неплохой товар, но есть небольшие недочеты в качестве.', 4, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
(3, 1, 'Средний товар, ожидал большего за эти деньги.', 3, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours'),
(1, 2, 'Супер товар! Рекомендую всем!', 5, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(2, 2, 'Хороший товар, но цена немного завышена.', 4, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(3, 3, 'Ужасное качество, не рекомендую.', 1, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
(4, 3, 'Нормальный товар за свои деньги.', 3, NOW() - INTERVAL '18 hours', NOW() - INTERVAL '18 hours'),
(5, 4, 'Лучшая покупка в этом году!', 5, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(1, 4, 'Хороший товар, но есть проблемы с доставкой.', 4, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(2, 5, 'Отличное соотношение цена/качество!', 5, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'); 