-- ============================================================
-- Afrisinc Asset Server — initial schema
-- Run with: psql $DATABASE_URL -f migrations/001_init.sql
-- ============================================================

BEGIN;

-- UUID extension (Postgres 13+ ships with gen_random_uuid built-in,
-- but we use application-generated IDs so this is optional)
CREATE EXTENSION IF NOT EXISTS pg_trgm; -- for fast ILIKE searches

-- ── Folders ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS folders (
    id          TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL,
    slug        TEXT        NOT NULL UNIQUE,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Assets ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS assets (
    id            TEXT        PRIMARY KEY,
    folder_id     TEXT        REFERENCES folders(id) ON DELETE SET NULL,
    name          TEXT        NOT NULL,
    original_name TEXT        NOT NULL,
    mime_type     TEXT        NOT NULL,
    size_bytes    BIGINT      NOT NULL DEFAULT 0,
    width         INT,
    height        INT,
    storage_key   TEXT        NOT NULL UNIQUE,
    tags          TEXT[]      NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ           -- soft delete
);

-- ── Indexes ───────────────────────────────────────────────────────────────────
-- Fast look-up by folder
CREATE INDEX IF NOT EXISTS idx_assets_folder_id
    ON assets (folder_id)
    WHERE deleted_at IS NULL;

-- Partial index for active assets only (most queries filter deleted_at IS NULL)
CREATE INDEX IF NOT EXISTS idx_assets_active_created
    ON assets (created_at DESC)
    WHERE deleted_at IS NULL;

-- Fast ILIKE search on name / original_name using trigrams
CREATE INDEX IF NOT EXISTS idx_assets_name_trgm
    ON assets USING GIN (name gin_trgm_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_assets_original_name_trgm
    ON assets USING GIN (original_name gin_trgm_ops)
    WHERE deleted_at IS NULL;

-- Tag array search
CREATE INDEX IF NOT EXISTS idx_assets_tags
    ON assets USING GIN (tags)
    WHERE deleted_at IS NULL;

-- MIME type prefix queries (e.g. LIKE 'image/%')
CREATE INDEX IF NOT EXISTS idx_assets_mime_type
    ON assets (mime_type)
    WHERE deleted_at IS NULL;

-- ── Auto-update updated_at ─────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE OR REPLACE TRIGGER assets_updated_at
    BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE TRIGGER folders_updated_at
    BEFORE UPDATE ON folders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
