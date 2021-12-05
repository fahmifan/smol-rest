oto:
	@echo "generate go server"
	@mkdir -p backend/restapi/generated
	@oto -template ./backend/restapi/definitions/templates/server.go.plush \
    	-out ./backend/restapi/generated/oto.gen.go \
    	-ignore Ignorer \
    	-pkg generated \
    	./backend/restapi/definitions
	@gofmt -w ./backend/restapi/generated/oto.gen.go ./backend/restapi/generated/oto.gen.go
	@echo "generate ts client"
	@mkdir -p web/src/service
	@oto -template ./backend/restapi/definitions/templates/client.ts.plush \
    	-out ./web/src/service/oto.gen.ts \
    	-ignore Ignorer \
    	-pkg generated \
    	./backend/restapi/definitions

proxy:
	@echo "run proxy"
	@caddy run -config Caddyfile.dev

run-server:
	@modd -f server.modd.conf

.PHONY: oto