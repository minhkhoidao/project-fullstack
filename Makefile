.PHONY: all build test lint clean \
	migrate-up migrate-down \
	docker-up docker-down docker-logs docker-build docker-build-web \
	dev web-install web-dev web-build \
	k8s-staging k8s-production argocd \
	run-gateway run-user run-product run-cart run-order \
	run-payment run-inventory run-notification run-review run-admin

GO := go
GOFLAGS := -v
BINARY_DIR := bin
DATABASE_URL ?= postgres://fashion:fashion_secret@localhost:5432/fashion_ecommerce?sslmode=disable

SERVICES := gateway user product cart order payment inventory notification review admin

# ──────────────────────────────────────────────────────────────── Build & Test

all: build

build:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/$$svc ./cmd/$$svc; \
	done

build-%:
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/$* ./cmd/$*

test:
	$(GO) test -race -cover ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BINARY_DIR)

# ──────────────────────────────────────────────────────────────── Migrations

migrate-up:
	@echo "Running all migrations..."
	@for schema in users products orders inventory reviews; do \
		echo "Migrating $$schema..."; \
		migrate -path migrations/$$schema -database "$(DATABASE_URL)" up; \
	done

migrate-down:
	@echo "Rolling back all migrations..."
	@for schema in reviews inventory orders products users; do \
		echo "Rolling back $$schema..."; \
		migrate -path migrations/$$schema -database "$(DATABASE_URL)" down 1; \
	done

# ──────────────────────────────────────────────────────────────── Docker

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-build:
	@for svc in $(SERVICES); do \
		echo "Building Docker image for $$svc..."; \
		docker build --build-arg SERVICE=$$svc \
			-f deploy/docker/Dockerfile.service \
			-t fashion-ecommerce/$$svc .; \
	done

docker-build-%:
	docker build --build-arg SERVICE=$* \
		-f deploy/docker/Dockerfile.service \
		-t fashion-ecommerce/$* .

docker-build-web:
	docker build -f deploy/docker/Dockerfile.web -t fashion-ecommerce/web .

# ──────────────────────────────────────────────────────────────── Frontend

web-install:
	cd web && yarn install

web-dev:
	cd web && yarn run dev

web-build:
	cd web && yarn run build

web-lint:
	cd web && yarn run lint

# ──────────────────────────────────────────────────────────────── Kubernetes

k8s-staging:
	kubectl apply -k deploy/k8s/overlays/staging

k8s-production:
	kubectl apply -k deploy/k8s/overlays/production

argocd:
	kubectl apply -f deploy/argocd/application.yaml
	kubectl apply -f deploy/argocd/application-staging.yaml

# ──────────────────────────────────────────────────────────── Run Services

define run-service
run-$(1):
	$(GO) run ./cmd/$(1)
endef

$(foreach svc,$(SERVICES),$(eval $(call run-service,$(svc))))

# ─────────────────────────────────────────────────────────────── Dev Mode

dev: docker-up
	@echo "Infrastructure started. Run individual services with 'make run-<service>'"
	@echo "Run the frontend with 'make web-dev'"
