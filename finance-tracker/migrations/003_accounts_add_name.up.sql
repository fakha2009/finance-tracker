-- Add name column to accounts
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS name VARCHAR(100);

-- Optional: backfill existing rows with a default name
UPDATE accounts SET name = COALESCE(name, 'Main Account');
