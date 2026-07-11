ALTER TABLE contests
  ADD COLUMN ranking_type VARCHAR(10) NOT NULL DEFAULT 'IOI' CHECK (ranking_type IN ('IOI', 'ICPC'));
