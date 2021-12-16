proxy:
	@echo "run proxy"
	@caddy run -config Caddyfile.dev

run-server:
	@modd -f server.modd.conf

.PHONY: oto