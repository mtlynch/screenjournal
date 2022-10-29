package store

import "github.com/mtlynch/screenjournal/v2"

type Store interface {
	ReadReviews() ([]screenjournal.Review, error)
}
