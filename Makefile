proxy:
	@echo "run proxy"
	@caddy run -config Caddyfile.dev

build:
	@GOOS=linux go build -o ./bin/smol ./internal/cmd/smol/main.go

run-server: doc
	@modd -f server.modd.conf

doc:
	@swag init -g internal/cmd/smol/main.go -o swagger

.PHONY: oto doc

