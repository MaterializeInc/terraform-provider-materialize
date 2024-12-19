BINARY=terraform-provider-materialize
PLATFORM=darwin_arm64
PLUGIN_PATH=~/.terraform.d/plugins/materialize.com/devex/materialize/0.1/${PLATFORM}

default: testacc

.PHONY: fmt
fmt:
	gofmt -l -s -w .
	terraform fmt -recursive

.PHONY: build
build:
	go build -o ${BINARY}

.PHONY: release
release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

.PHONY: install
install:
	mkdir -p ${PLUGIN_PATH}
	go build -o ${PLUGIN_PATH}/${BINARY}

.PHONY: test
test:
	go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m -run TestAcc -parallel 1

.PHONY: docs
docs:
	go generate ./...
