proxy:
	@echo "run proxy"
	@caddy run -config Caddyfile.dev

run-server: doc
	@modd -f server.modd.conf

doc:
	@swag init -g internal/cmd/smol/main.go -o swagger

.PHONY: oto doc

