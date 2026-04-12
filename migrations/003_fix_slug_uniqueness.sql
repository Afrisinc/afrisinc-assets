-- ============================================================
-- Fix slug uniqueness constraint for hierarchical folders
-- Allow same slug under different parents (e.g., multiple "templates" folders)
-- Run with: psql $DATABASE_URL -f migrations/003_fix_slug_uniqueness.sql
-- ============================================================

BEGIN;

-- Drop the global UNIQUE constraint on slug
ALTER TABLE folders
  DROP CONSTRAINT IF EXISTS folders_slug_key;

-- Add composite unique constraint: (parent_id, slug)
-- This allows "templates" folder under different parents, but prevents duplicates within same parent
ALTER TABLE folders
  ADD CONSTRAINT folders_parent_id_slug_key UNIQUE (parent_id, slug);

COMMIT;
