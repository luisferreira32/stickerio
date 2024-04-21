# {{{

.PHONY: _gen_oas_api
_gen_oas_api: generated/

generated/: api.yaml
	rm -rf generated
	docker run --user $$(id -u):$$(id -g) --rm -v $$PWD:/local -w /local openapitools/openapi-generator-cli generate -i /local/api.yaml -g go -o /local/generated --additional-properties=packageName=generated
	mkdir -p api/ && cp generated/*.go api/

bin/stickerio-api: internal/* cmd/stickerio-api/main.go _gen_oas_api
	CGO_ENABLED=0 go build -o bin/stickerio-api ./cmd/stickerio-api/main.go

bin/stickerio: internal/* cmd/stickerio/main.go _gen_oas_api
	CGO_ENABLED=0 go build -o bin/stickerio ./cmd/stickerio/main.go

.PHONY: _clean_bin
_clean_bin:
	rm -rf bin/*

.PHONY: compile
compile: bin/stickerio-api

# }}}

# {{{

tmp/mockdb.sqlite3: databases/gamedb/schema.sql
	cat databases/gamedb/schema.sql | sqlite3 tmp/mockdb.sqlite3

.PHONY: _clean_mock_db
_clean_mock_db:
	rm -f tmp/mockdb.sqlite3


.PHONY: run_dummy
dummy_server: tmp/mockdb.sqlite3 bin/stickerio-api
	DB_HOST=tmp/mockdb.sqlite3 bin/stickerio-api

# }}}


# {{{

.PHONY: test
test: tmp/mockdb.sqlite3 bin/stickerio-api
	DB_HOST=tmp/mockdb.sqlite3 bin/stickerio-api &
	go test -v  ./test/

.PHONY: clean
clean: _clean_bin _clean_mock_db

# }}}