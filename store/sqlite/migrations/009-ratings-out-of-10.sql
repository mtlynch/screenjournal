-- Change ratings from being out of 10 to out of 5.
UPDATE reviews
SET rating = rating * 2;
