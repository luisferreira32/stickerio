# {{{

bin/stickerio-api: internal/* cmd/stickerio-api/main.go
	CGO_ENABLED=0 go build -o bin/stickerio-api ./cmd/stickerio-api/main.go

bin/stickerio: internal/* cmd/stickerio/main.go
	CGO_ENABLED=0 go build -o bin/stickerio ./cmd/stickerio/main.go

.PHONY: _clean_bin
_clean_bin:
	rm -rf bin/*

.PHONY: compile
compile: bin/stickerio-api bin/stickerio

# }}}

# {{{

tmp/mockdata.sqlite3: databases/gamedb/schema.sql
	cat databases/gamedb/schema.sql | sqlite3 tmp/mockdata.sqlite3

.PHONY: _clean_mock_data
_clean_mock_data:
	rm -f tmp/mockdata.sqlite3

.PHONY: mock_db
mock_db: test/mockdata/*.sql
	cat test/mockdata/*.sql | sqlite3 tmp/mockdata.sqlite3

.PHONY: run_dummy
dummy_server: tmp/mockdata.sqlite3 bin/stickerio-api
	DB_HOST=tmp/mockdata.sqlite3 bin/stickerio-api

# }}}


# {{{

.PHONY: clean
clean: _clean_mock_data _clean_bin

# }}}