ALTER TABLE blogs ADD COLUMN slug VARCHAR(255);
ALTER TABLE blogs ADD COLUMN is_published BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE blogs ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

-- Since slug is required and unique, we will set temporary slugs for existing rows and then add UNIQUE constraint
UPDATE blogs SET slug = 'blog-' || id WHERE slug IS NULL;
ALTER TABLE blogs ALTER COLUMN slug SET NOT NULL;
ALTER TABLE blogs ADD CONSTRAINT blogs_slug_key UNIQUE (slug);

CREATE INDEX idx_blogs_slug ON blogs(slug);
CREATE INDEX idx_blogs_published ON blogs(is_published);
