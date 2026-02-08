-- Миграция для добавления функционала валют и счетов

-- Создание таблицы валют
CREATE TABLE currencies (
    id SERIAL PRIMARY KEY,
    code VARCHAR(3) UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    symbol VARCHAR(5) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы счетов
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    currency_id INTEGER REFERENCES currencies(id),
    balance DECIMAL(15,2) DEFAULT 0,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы курсов валют
CREATE TABLE exchange_rates (
    id SERIAL PRIMARY KEY,
    base_currency_id INTEGER REFERENCES currencies(id),
    target_currency_id INTEGER REFERENCES currencies(id),
    rate DECIMAL(15,6) NOT NULL,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(base_currency_id, target_currency_id)
);

-- Добавление столбца валюты по умолчанию для пользователей
ALTER TABLE users ADD COLUMN default_currency_id INTEGER REFERENCES currencies(id);

-- Добавление столбца счета для транзакций
ALTER TABLE transactions ADD COLUMN account_id INTEGER REFERENCES accounts(id);

-- Вставка базовых валют
INSERT INTO currencies (code, name, symbol) VALUES
('RUB', 'Russian Ruble', '₽'),
('USD', 'US Dollar', '$'),
('TJS', 'Tajikistani Somoni', 'SM');

-- Индексы для улучшения производительности
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_accounts_currency_id ON accounts(currency_id);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_base_target ON exchange_rates(base_currency_id, target_currency_id);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_last_updated ON exchange_rates(last_updated);
CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
