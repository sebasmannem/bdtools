build:
	sh ./set_version.sh
	go mod tidy
	go build -o ./bin/bdtools ./cmd/bdtools

build_image:
	docker build . --tag sebasmannem/bdtools

debug:
	go build -gcflags "all=-N -l" -o ./bin/bdtools ./cmd/bdtools
	~/go/bin/dlv --headless --listen=:2345 --api-version=2 --accept-multiclient exec ./bin/bdtools -- -c config/bdtools.yaml -d

run:
	./bin/bdtools -c config/bdtools.yaml -d

fmt:
	gofmt -w .
	goimports -w .
	gci write .

compose:
	./docker-compose-tests.sh

test: gotest sec lint

sec:
	gosec ./...

lint:
	golangci-lint run

gotest:
	go test -v ./...
