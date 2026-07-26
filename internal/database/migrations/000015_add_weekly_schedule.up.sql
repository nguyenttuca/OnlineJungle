ALTER TABLE classes ADD COLUMN weekly_schedule JSONB NOT NULL DEFAULT '[]'::jsonb;
