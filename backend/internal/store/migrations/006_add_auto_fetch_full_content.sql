ALTER TABLE feeds ADD COLUMN auto_fetch_full_content INTEGER DEFAULT 0;
ALTER TABLE groups ADD COLUMN auto_fetch_full_content INTEGER DEFAULT 0;
ALTER TABLE items ADD COLUMN full_content TEXT;
