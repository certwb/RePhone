-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT,
    oauth_provider TEXT,
    oauth_id TEXT,
    city_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (city_id) REFERENCES cities(id)
);

-- Справочник городов
CREATE TABLE IF NOT EXISTS cities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

-- Справочник брендов
CREATE TABLE IF NOT EXISTS brands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

-- Таблица объявлений
CREATE TABLE IF NOT EXISTS phones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    brand_id INTEGER NOT NULL,
    model TEXT NOT NULL,
    storage INTEGER, -- Объем памяти в ГБ
    battery_health INTEGER, -- Состояние аккумулятора в %
    price INTEGER NOT NULL, -- Цена в KZT
    condition TEXT NOT NULL,
    description TEXT,
    contact_phone TEXT DEFAULT '',
    views_count INTEGER DEFAULT 0,
    status TEXT DEFAULT 'active', -- 'active', 'sold', 'deleted'
    city_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    bumped_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (brand_id) REFERENCES brands(id),
    FOREIGN KEY (city_id) REFERENCES cities(id)
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_phones_brand_id ON phones(brand_id);
CREATE INDEX IF NOT EXISTS idx_phones_price ON phones(price);
CREATE INDEX IF NOT EXISTS idx_phones_status ON phones(status);
CREATE INDEX IF NOT EXISTS idx_phones_city_id ON phones(city_id);
CREATE INDEX IF NOT EXISTS idx_phones_search ON phones(status, city_id, brand_id, price);

-- Таблица изображений
CREATE TABLE IF NOT EXISTS images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone_id INTEGER NOT NULL,
    image_url TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT 0,
    FOREIGN KEY (phone_id) REFERENCES phones(id) ON DELETE CASCADE
);

-- Первичное заполнение справочника брендов
INSERT OR IGNORE INTO brands (name) VALUES 
('Apple'), ('Samsung'), ('Xiaomi'), ('Google'), ('OnePlus'), ('Huawei'), ('Sony'), ('Другой');

-- Первичное заполнение справочника городов
INSERT OR IGNORE INTO cities (name) VALUES 
('Алматы'), ('Астана'), ('Шымкент'), ('Актобе'), ('Караганда'), ('Атырау'), ('Актау'), ('Павлодар'), ('Костанай'), ('Уральск');

-- Таблица сессий
CREATE TABLE IF NOT EXISTS sessions (
    session_token TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Таблица избранного (Wishlist)
CREATE TABLE IF NOT EXISTS favorites (
    user_id INTEGER NOT NULL,
    phone_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, phone_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (phone_id) REFERENCES phones(id) ON DELETE CASCADE
);

-- Таблица сообщений (Messages)
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id INTEGER NOT NULL,
    receiver_id INTEGER NOT NULL,
    phone_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(sender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(receiver_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(phone_id) REFERENCES phones(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS promo_pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone_id INTEGER NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    html_content TEXT NOT NULL,
    views_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(phone_id) REFERENCES phones(id) ON DELETE CASCADE
);

-- Индексы для быстрого поиска диалогов пользователя
CREATE INDEX IF NOT EXISTS idx_messages_users ON messages(sender_id, receiver_id);

-- Таблица отзывов (Reviews)
CREATE TABLE IF NOT EXISTS reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    reviewer_id INTEGER NOT NULL,
    seller_id INTEGER NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (seller_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(reviewer_id, seller_id) -- Один пользователь может оставить только один отзыв одному продавцу
);

-- Индекс для быстрого подсчета рейтинга продавца
CREATE INDEX IF NOT EXISTS idx_reviews_seller ON reviews(seller_id);
