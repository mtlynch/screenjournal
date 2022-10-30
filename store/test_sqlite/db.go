package test_sqlite

import (
	"fmt"

	"github.com/mtlynch/screenjournal/v2/random"
	"github.com/mtlynch/screenjournal/v2/store"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func New() store.Store {
	const optimizeForLitestream = false
	return sqlite.New(ephemeralDbURI(), optimizeForLitestream)
}

func ephemeralDbURI() string {
	name := random.String(
		10,
		[]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"))
	return fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
}
