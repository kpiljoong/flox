APP=flox
BIN=./flox
SRC=./...

.PHONY: all build test lint run clean

all: test lint build

build:
	go build -o $(BIN) main.go

docker-build:
	docker build -t flox:dev .

kind-deploy: docker-build
	kind load docker-image flox:dev
	kubectl apply -f manifests/flox-configmap.yaml
	kubectl apply -f manifests/flox-dev-stack.yaml

test:
	go test -v $(SRC)

lint:
	golangci-lint run

run:
	go run main.go --config pipeline.yaml

clean:
	rm -f $(BIN)
