-- ============================================================
-- Add nested folder support (parent-child relationships)
-- Run with: psql $DATABASE_URL -f migrations/002_add_nested_folders.sql
-- ============================================================

BEGIN;

-- Add parent_id and path columns to folders table
ALTER TABLE folders
  ADD COLUMN IF NOT EXISTS parent_id TEXT REFERENCES folders(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS path TEXT NOT NULL DEFAULT '';

-- Create index for fast parent lookups
CREATE INDEX IF NOT EXISTS idx_folders_parent_id
  ON folders (parent_id)
  WHERE parent_id IS NOT NULL;

-- Create index for path queries
CREATE INDEX IF NOT EXISTS idx_folders_path
  ON folders (path);

-- Update trigger to also update folders table
CREATE OR REPLACE TRIGGER folders_updated_at
  BEFORE UPDATE ON folders
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
