build-workerpool: tidy
	docker build -t goflow-workerpool -f dockerfiles/Dockerfile.workerpool .
	docker tag goflow-workerpool:latest jamestait12/goflow-workerpool:latest
	docker push jamestait12/goflow-workerpool:latest

build-pluginbuilder: tidy
	docker build -t goflow-plugin-builder -f dockerfiles/Dockerfile.pluginbuilder .
	docker tag goflow-plugin-builder:latest jamestait12/goflow-plugin-builder:latest
	docker push jamestait12/goflow-plugin-builder:latest

build-server: tidy
	protoc --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
    	grpc/proto/goflow.proto

	docker build -t goflow-server -f dockerfiles/Dockerfile.server .
	docker tag goflow-server:latest jamestait12/goflow-server:latest
	docker push jamestait12/goflow-server:latest

build: build-server build-workerpool build-pluginbuilder

clean:
	go clean -modcache

tidy:
	go mod tidy

test-unit: 
	go test -tags=unit -race -coverprofile=coverage.out -covermode=atomic -shuffle=on ./...

INTEGRATION_TEST_PLUGIN_DIR = test/integration/testdata/handlers

build-test-plugins:
	rm -f $(INTEGRATION_TEST_PLUGIN_DIR)/*.so
	find $(INTEGRATION_TEST_PLUGIN_DIR) -name "*.go" | while read -r gofile; do \
		go build -buildmode=plugin -o "$${gofile%.go}.so" "$$gofile"; \
	done

test-integration: clean tidy build-test-plugins
	go test -tags=integration -timeout=1m -shuffle=on ./test/integration/...

