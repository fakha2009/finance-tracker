-- Откат миграции для валют и счетов

-- Удаление индексов
DROP INDEX IF EXISTS idx_transactions_account_id;
DROP INDEX IF EXISTS idx_exchange_rates_last_updated;
DROP INDEX IF EXISTS idx_exchange_rates_base_target;
DROP INDEX IF EXISTS idx_accounts_currency_id;
DROP INDEX IF EXISTS idx_accounts_user_id;

-- Удаление столбца account_id из transactions
ALTER TABLE transactions DROP COLUMN IF EXISTS account_id;

-- Удаление столбца default_currency_id из users
ALTER TABLE users DROP COLUMN IF EXISTS default_currency_id;

-- Удаление таблиц
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS currencies;
