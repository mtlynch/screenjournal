//go:build dev

package handlers

import (
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func (s *Server) addDevRoutes() {
	log.Printf("adding dev routes")
	s.router.HandleFunc("/api/debug/wipe-db", s.wipeDB()).Methods(http.MethodGet)
}

func (s Server) wipeDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sqlStore, ok := s.store.(*sqlite.DB)
		if !ok {
			log.Fatalf("store is not SQLite, can't wipe database")
		}
		sqlStore.Clear()
	}
}
