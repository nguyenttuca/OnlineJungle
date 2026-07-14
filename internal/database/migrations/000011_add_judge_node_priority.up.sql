-- Remove duplicate base_urls by keeping only the oldest record
DELETE FROM judge_nodes WHERE id NOT IN (
    SELECT MIN(id) FROM judge_nodes GROUP BY base_url
);

-- Add unique constraint
ALTER TABLE judge_nodes ADD CONSTRAINT judge_nodes_base_url_key UNIQUE (base_url);

-- Add is_local column
ALTER TABLE judge_nodes ADD COLUMN is_local BOOLEAN NOT NULL DEFAULT false;

-- Auto-detect existing local nodes
UPDATE judge_nodes SET is_local = true 
WHERE base_url LIKE '%localhost%' OR base_url LIKE '%127.0.0.1%' OR base_url LIKE '%::1%';
