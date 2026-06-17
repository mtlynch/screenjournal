ALTER TABLE reviews
ADD COLUMN is_draft INTEGER NOT NULL DEFAULT 0;

-- A user can have at most one draft per movie or per (TV show, season). We
-- COALESCE the nullable media columns because SQLite treats each NULL as
-- distinct in a UNIQUE index, which would otherwise let duplicate movie drafts
-- (tv_show_id/tv_show_season NULL) slip through.
CREATE UNIQUE INDEX reviews_unique_draft_per_media
ON reviews (
    review_owner,
    COALESCE(movie_id, -1),
    COALESCE(tv_show_id, -1),
    COALESCE(tv_show_season, -1)
)
WHERE is_draft = 1;
