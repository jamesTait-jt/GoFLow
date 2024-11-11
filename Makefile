build-workerpool:
	go mod tidy
	docker build -t workerpool -f dockerfiles/Dockerfile.workerpool .

build-pluginbuilder:
	go mod tidy
	docker build -t plugin-builder -f dockerfiles/Dockerfile.pluginbuilder .

build-goflow:
	go mod tidy
	protoc --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
    	cmd/goflow/goflow/goflow.proto

	docker build -t goflow-server -f dockerfiles/Dockerfile.goflow .

build: build-goflow build-workerpool build-pluginbuilder

clean:
	go clean -modcache

test-unit: 
	go test -tags=unit -race -coverprofile=coverage.out -covermode=atomic -shuffle=on ./...

INTEGRATION_TEST_PLUGIN_DIR = test/integration/testdata/handlers

test-integration: clean
	go mod tidy
	rm -f $(INTEGRATION_TEST_PLUGIN_DIR)/*.so
	find $(INTEGRATION_TEST_PLUGIN_DIR) -name "*.go" | while read -r gofile; do \
		go build -buildmode=plugin -o "$${gofile%.go}.so" "$$gofile"; \
	done
	go test -tags=integration -timeout=1m -shuffle=on ./test/integration/...

