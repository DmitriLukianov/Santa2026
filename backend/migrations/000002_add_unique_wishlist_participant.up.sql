CREATE UNIQUE INDEX IF NOT EXISTS idx_wishlists_participant_id
    ON wishlists (participant_id)
    WHERE participant_id IS NOT NULL;
