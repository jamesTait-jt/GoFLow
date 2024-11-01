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

test: 
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
