.PHONY: all
all: format test build

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	go build -o amgate.exe .


.PHONY: devenv-create
devenv-create:
	kind create cluster --name amgate
	kubectl cluster-info --context kind-amgate

.PHONY: devenv-destroy
devenv-destroy:
	kind delete cluster --name amgate

development-image-load:
	docker image build -t amgate:develop .
	kind load docker-image amgate:develop --name amgate

