module github.com/mtlynch/screenjournal/v2

go 1.23

require (
	codeberg.org/mtlynch/simpleauth/v3 v3.0.0
	github.com/go-test/deep v1.0.8
	github.com/gomarkdown/markdown v0.0.0-20240723152757-afa4a469d4f9
	github.com/gorilla/mux v1.8.0
	github.com/kylelemons/godebug v1.1.0
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/mtlynch/gorilla-handlers v1.5.2
	github.com/ncruces/go-sqlite3 v0.22.0
	github.com/ryanbradynd05/go-tmdb v0.0.0-20220721194547-2ab6191c6273
)

require gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect

require (
	codeberg.org/mtlynch/go-evolutionary-migrate v0.0.1
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.1 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/kylelemons/go-gypsy v1.0.0 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/tetratelabs/wazero v1.8.2 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace codeberg.org/mtlynch/simpleauth/v3 => ./third_party/simpleauth
