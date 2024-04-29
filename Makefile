

.PHONY: system-server fmt vet

all: system-server

tidy: 
	go mod tidy

devbox-server: ;$(info $(M)...Begin to build system-server.) @
	go build -o output/devbox-server ./cmd/devbox/main.go

linux: ;$(info $(M)...Begin to build system-server - linux version.) @
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc CGO_LDFLAGS="-static" go build -a -o output/devbox-server ./cmd/devbox/main.go


run: ; $(info $(M)...Run system-server.)
	go run --tags "sqlite_trace" ./cmd/devbox/main.go -v 4 --db /tmp/test.db

.PHONY: docker-build-frontend
docker-build-frontend: ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push-frontend
docker-push-frontend: ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: docker-build-server
docker-build-server: ## Build docker image with the manager.
	docker build -t ${IMG} -f Dockerfile.server .

.PHONY: docker-push-server
docker-push-server: ## Push docker image with the manager.
	docker push ${IMG}