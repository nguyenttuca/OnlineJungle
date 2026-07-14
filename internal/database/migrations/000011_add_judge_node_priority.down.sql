ALTER TABLE judge_nodes DROP COLUMN IF EXISTS is_local;
ALTER TABLE judge_nodes DROP CONSTRAINT IF EXISTS judge_nodes_base_url_key;
