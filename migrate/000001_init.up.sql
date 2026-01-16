CREATE TABLE incidents (
   id SERIAL PRIMARY KEY,
   description TEXT NOT NULL,
   x DOUBLE PRECISION NOT NULL,
   y DOUBLE PRECISION NOT NULL,
   status VARCHAR(20) NOT NULL DEFAULT 'active'
);

-- Таблица для логирования проверок местоположения
CREATE TABLE IF NOT EXISTS location_checks (
                                               id SERIAL PRIMARY KEY,
                                               user_id BIGINT NOT NULL,          -- ID пользователя из мобильного приложения
                                               x DOUBLE PRECISION NOT NULL,      -- Координата X
                                               y DOUBLE PRECISION NOT NULL,      -- Координата Y
                                               created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() -- Время проверки для статистики
    );

CREATE INDEX IF NOT EXISTS idx_location_checks_stats
    ON location_checks (created_at, user_id);