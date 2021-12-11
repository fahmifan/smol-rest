oto:
	@echo "generate go server"
	@mkdir -p backend/restapi/gen
	@oto -template ./backend/restapi/definitions/templates/server.go.plush \
    	-out ./backend/restapi/gen/oto.gen.go \
    	-ignore Ignorer \
    	-pkg gen \
    	./backend/restapi/definitions
	@gofmt -w ./backend/restapi/gen/oto.gen.go ./backend/restapi/gen/oto.gen.go
	@echo "generate ts client"
	@mkdir -p web/src/service
	@oto -template ./backend/restapi/definitions/templates/client.ts.plush \
    	-out ./frontend/src/service/oto.gen.ts \
    	-ignore Ignorer \
    	-pkg gen \
    	./backend/restapi/definitions

proxy:
	@echo "run proxy"
	@caddy run -config Caddyfile.dev

run-server:
	@modd -f server.modd.conf

.PHONY: oto