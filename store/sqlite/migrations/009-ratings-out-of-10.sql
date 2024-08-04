-- Change ratings from being out of 5 to out of 10.
-- This undoes migration 005.
UPDATE reviews
SET rating = rating * 2;
