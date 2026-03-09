# Fashion Ecommerce Platform

A full-stack fashion ecommerce platform built with Go microservices, PostgreSQL, Redis, Kafka, and a Next.js 16 frontend.

## Architecture

```
Browser / Mobile ──HTTPS──▶ API Gateway :8080
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
   User Service :8081    Product Service :8082    Cart Service :8083
   Order Service :8084   Payment Service :8085   Inventory Service :8086
   Review Service :8087  Notification :8088      Admin Service :8089
        │                       │                       │
        └───────────┬───────────┴───────────┬───────────┘
                    │                       │
              PostgreSQL                  Redis
                    │
                  Kafka
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, Chi router, pgx, sqlc |
| Database | PostgreSQL 17 (schema-per-service) |
| Cache | Redis 7 (sessions, cart) |
| Events | Apache Kafka (order/payment/inventory events) |
| Auth | JWT (access + refresh tokens), bcrypt |
| Frontend | Next.js 16, TypeScript, Tailwind CSS |
| Containers | Multi-stage Docker builds, distroless runtime |
| Orchestration | Kubernetes (Kustomize), ArgoCD |

## Project Structure

```
fashion-ecommerce/
├── cmd/                           # Service entrypoints
│   ├── gateway/                   # API Gateway (:8080)
│   ├── user/                      # User Service (:8081)
│   ├── product/                   # Product Service (:8082)
│   ├── cart/                      # Cart Service (:8083)
│   ├── order/                     # Order Service (:8084)
│   ├── payment/                   # Payment Service (:8085)
│   ├── inventory/                 # Inventory Service (:8086)
│   ├── review/                    # Review Service (:8087)
│   ├── notification/              # Notification Service (:8088)
│   └── admin/                     # Admin Service (:8089)
├── internal/                      # Private Go packages
│   ├── platform/                  # Shared: config, db, redis, kafka, logger, auth, server
│   ├── user/                      # handler / service / repository / model
│   ├── product/
│   ├── cart/
│   ├── order/
│   ├── payment/
│   ├── inventory/
│   ├── notification/
│   ├── review/
│   └── admin/
├── pkg/                           # Public shared code
│   ├── dto/                       # Pagination params
│   ├── event/                     # Kafka event types
│   └── httputil/                  # JSON response helpers, middleware
├── migrations/                    # SQL migrations per domain
├── web/                           # Next.js 16 frontend
│   ├── src/app/                   # Pages: /, /products, /cart, /checkout, /account, /admin
│   ├── src/components/            # UI components
│   ├── src/stores/                # Zustand state management
│   └── src/lib/                   # API client
├── deploy/
│   ├── docker/                    # Dockerfiles (multi-stage)
│   ├── k8s/                       # Kustomize: base + overlays
│   └── argocd/                    # ArgoCD Application manifests
├── docker-compose.yml             # Local dev stack
├── Makefile                       # Build, test, run targets
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Node.js 20+ (for frontend)

### Start Infrastructure

```bash
make docker-up
```

This starts PostgreSQL, Redis, Kafka, and Zookeeper.

### Run Migrations

```bash
export DATABASE_URL="postgres://fashion:fashion_secret@localhost:5432/fashion_ecommerce?sslmode=disable"
make migrate-up
```

### Run Services

Run each service in a separate terminal:

```bash
make run-gateway
make run-user
make run-product
make run-cart
make run-order
make run-payment
make run-inventory
make run-notification
make run-review
make run-admin
```

Or build all binaries:

```bash
make build
```

### Run Frontend

```bash
cd web
yarn install
yarn  dev
```

The frontend runs at `http://localhost:3000` and proxies API requests to the gateway at `http://localhost:8080`.

### Run Tests

```bash
make test
```

## API Endpoints

### Auth
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | No | Register new user |
| POST | `/api/v1/auth/login` | No | Login, get JWT tokens |

### Users
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/users/me` | Yes | Get profile |
| PUT | `/api/v1/users/me` | Yes | Update profile |
| POST | `/api/v1/users/me/addresses` | Yes | Add address |
| GET | `/api/v1/users/me/addresses` | Yes | List addresses |
| DELETE | `/api/v1/users/me/addresses/{id}` | Yes | Delete address |

### Products
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/products` | No | List products (filters, pagination) |
| GET | `/api/v1/products/{idOrSlug}` | No | Get product details |
| GET | `/api/v1/categories` | No | List categories |

### Cart
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/cart` | Yes | Get cart |
| POST | `/api/v1/cart/items` | Yes | Add item |
| PUT | `/api/v1/cart/items/{variantID}` | Yes | Update quantity |
| DELETE | `/api/v1/cart/items/{variantID}` | Yes | Remove item |
| DELETE | `/api/v1/cart` | Yes | Clear cart |

### Orders
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/orders` | Yes | Create order from cart |
| GET | `/api/v1/orders` | Yes | List my orders |
| GET | `/api/v1/orders/{id}` | Yes | Get order details |
| POST | `/api/v1/orders/{id}/cancel` | Yes | Cancel order |

### Reviews
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/reviews/product/{id}` | No | List product reviews |
| GET | `/api/v1/reviews/product/{id}/summary` | No | Rating summary |
| POST | `/api/v1/reviews` | Yes | Create review |
| PUT | `/api/v1/reviews/{id}` | Yes | Update review |
| DELETE | `/api/v1/reviews/{id}` | Yes | Delete review |

### Admin
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/admin/dashboard` | Admin | Dashboard stats |
| GET | `/api/v1/admin/analytics/revenue` | Admin | Revenue by day |
| GET | `/api/v1/admin/analytics/top-products` | Admin | Top selling products |
| GET | `/api/v1/admin/orders` | Admin | All orders |
| PUT | `/api/v1/admin/orders/{id}/status` | Admin | Update order status |
| GET | `/api/v1/admin/users` | Admin | All users |
| PUT | `/api/v1/admin/users/{id}/role` | Admin | Update user role |

## Kafka Events

| Topic | Producer | Consumers |
|-------|----------|-----------|
| `order.created` | Order Service | Inventory, Notification |
| `order.paid` | Payment Service | Notification |
| `order.cancelled` | Order Service | Inventory, Notification |
| `inventory.low-stock` | Inventory Service | Notification |
| `payment.completed` | Payment Service | Order, Notification |
| `payment.failed` | Payment Service | Notification |

## Database Schemas

- **users**: `users`, `addresses`
- **products**: `categories`, `products`, `product_images`, `product_variants`
- **orders**: `orders`, `order_items`, `payments`
- **inventory**: `inventory`
- **reviews**: `reviews`, `review_images`

## Deployment

### Docker Build

```bash
# Build a specific service
docker build --build-arg SERVICE=gateway -f deploy/docker/Dockerfile.service -t fashion-ecommerce/gateway .

# Build the frontend
docker build -f deploy/docker/Dockerfile.web -t fashion-ecommerce/web .
```

### Kubernetes

```bash
# Staging
kubectl apply -k deploy/k8s/overlays/staging

# Production
kubectl apply -k deploy/k8s/overlays/production
```

### ArgoCD

```bash
kubectl apply -f deploy/argocd/application.yaml
kubectl apply -f deploy/argocd/application-staging.yaml
```
