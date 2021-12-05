oto:
	@echo "generate go server"
	@mkdir -p internal/restapi/generated
	@oto -template ./internal/restapi/definitions/templates/server.go.plush \
    	-out ./internal/restapi/generated/oto.gen.go \
    	-ignore Ignorer \
    	-pkg generated \
    	./internal/restapi/definitions
	@gofmt -w ./internal/restapi/generated/oto.gen.go ./internal/restapi/generated/oto.gen.go
	@echo "generate ts client"
	@mkdir -p web/src/service
	@oto -template ./internal/restapi/definitions/templates/client.ts.plush \
    	-out ./web/src/service/oto.gen.ts \
    	-ignore Ignorer \
    	-pkg generated \
    	./internal/restapi/definitions

.PHONY: oto