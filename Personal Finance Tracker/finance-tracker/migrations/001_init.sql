-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы категорий
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(10) CHECK (type IN ('income', 'expense')) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы транзакций
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    amount DECIMAL(15,2) NOT NULL,
    description TEXT,
    date DATE NOT NULL,
    type VARCHAR(10) CHECK (type IN ('income', 'expense')) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы сессий
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для улучшения производительности
CREATE INDEX IF NOT EXISTS idx_transactions_user_id_date ON transactions(user_id, date);
CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);
CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Вставка стандартных категорий для демо
INSERT INTO categories (user_id, name, description, type) VALUES
(NULL, 'Salary', 'Monthly salary', 'income'),
(NULL, 'Freelance', 'Freelance work', 'income'),
(NULL, 'Investments', 'Investment income', 'income'),
(NULL, 'Food', 'Groceries and dining', 'expense'),
(NULL, 'Transport', 'Transportation costs', 'expense'),
(NULL, 'Utilities', 'Bills and utilities', 'expense'),
(NULL, 'Entertainment', 'Entertainment expenses', 'expense'),
(NULL, 'Healthcare', 'Medical expenses', 'expense')
ON CONFLICT DO NOTHING;

