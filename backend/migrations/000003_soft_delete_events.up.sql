-- Soft delete для событий: вместо физического удаления ставим метку deleted_at.
-- Это сохраняет историю жеребьёвки, участников и сообщений чата.

ALTER TABLE events ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Индекс для фильтрации неудалённых событий
CREATE INDEX IF NOT EXISTS idx_events_deleted_at ON events (deleted_at) WHERE deleted_at IS NULL;
