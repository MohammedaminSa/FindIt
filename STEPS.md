# FindIt: Master Step-by-Step Implementation Plan

This document tracks every phase, step, sub-step, and task required to build **FindIt** from scratch to production.

---

## Progress Overview

- [x] **Phase 1: Environment & Database Foundation**
  - [x] Step 1: Git Repository, Dockerized PostgreSQL & Environment
  - [x] Step 2: Relational Database Schema, SQL Migrations & Category Seeding
- [ ] **Phase 2: Go Backend Architecture & Security**
  - [x] Step 3: Clean Architecture, Connection Pooling & Gin Engine
  - [ ] Step 4: User Authentication & Security Layer (Bcrypt, JWT & Middleware)
    - [x] Step 4.1: User Models & DTOs (`internal/models/user.go`)
    - [x] Step 4.2: Password Hashing (`pkg/hasher/hasher.go`)
    - [x] Step 4.3: JWT Token Management (`pkg/token/jwt.go`)
    - [x] Step 4.4: User Repository (`internal/repository/user_repository.go`)
    - [x] Step 4.5: Auth Service (`internal/service/auth_service.go`)
    - [ ] Step 4.6: Auth Handler & Routes (`internal/handler/auth.go`)
    - [ ] Step 4.7: Auth Middleware (`internal/middleware/auth.go`)
    - [ ] Step 4.8: Verification & API Testing
- [ ] **Phase 3: Core Domain Services (Lost & Found Features)**
  - [ ] Step 5: Category & Item Reporting APIs (Image Uploads & Location)
  - [ ] Step 6: Search & Filtering Engine (PostgreSQL Full-Text Search)
  - [ ] Step 7: Ownership Claims Lifecycle & Verification Workflow
  - [ ] Step 8: In-App Messaging & Notifications
- [ ] **Phase 4: Flutter Mobile Application**
  - [ ] Step 9: Mobile Scaffolding & State Architecture (Riverpod & GoRouter)
  - [ ] Step 10: Authentication UI & Auth State Flow
  - [ ] Step 11: Lost & Found Discovery Feed & Search UI
  - [ ] Step 12: Item Reporting Flow (Post Lost/Found)
  - [ ] Step 13: Claim Submission & Claim Review Dashboard
  - [ ] Step 14: Messaging & Chat Screen
- [ ] **Phase 5: Production Deployment & DevOps**
  - [ ] Step 15: Production Containerization & Cloud Deployment

---

## Detailed Step-by-Step Breakdown

### Phase 1: Environment & Database Foundation

#### [x] Step 1: Git Repository & Dockerized PostgreSQL
- **Goal:** Set up local environment, version control, and containerized PostgreSQL 16.
- **Tasks:**
  - [x] Initialize Git repository with main branch.
  - [x] Create `.gitignore` for Go binaries, secrets, and Flutter artifacts.
  - [x] Create `.env.example` and `.env` following Twelve-Factor config.
  - [x] Create `docker-compose.yml` with `postgres:16-alpine`, health check, and persistent volume `pgdata`.
- **Verification:**
  ```bash
  docker compose up -d
  docker compose exec postgres pg_isready -U findit_user -d findit_db
  ```

#### [x] Step 2: Relational Database Schema & SQL Migrations
- **Goal:** Design the complete relational data model and version-controlled migrations.
- **Tasks:**
  - [x] Create `backend/migrations/000001_init_schema.up.sql`.
  - [x] Define tables: `users`, `categories`, `items`, `claims`, `messages`, `notifications`.
  - [x] Configure UUIDv4 primary keys via `gen_random_uuid()`.
  - [x] Add foreign key cascades (`CASCADE` for claims/messages, `RESTRICT` for categories).
  - [x] Add check constraints (`LOST`/`FOUND` types, `ACTIVE`/`CLAIMED`/`RESOLVED` statuses).
  - [x] Create PostgreSQL GIN Full-Text Search index on items (`idx_items_search`).
  - [x] Add category seed data (`Electronics`, `Wallets & Purses`, `Keys`, `Documents & IDs`, `Bags & Luggage`, `Jewelry & Watches`, `Pets`, `Clothing`, `Other`).
  - [x] Create `backend/migrations/000001_init_schema.down.sql` for rollback safety.
- **Verification:**
  ```bash
  docker compose exec -i postgres psql -U findit_user -d findit_db < backend/migrations/000001_init_schema.up.sql
  docker compose exec -i postgres psql -U findit_user -d findit_db -c "\dt"
  ```

---

### Phase 2: Go Backend Architecture & Security

#### [x] Step 3: Clean Architecture, Connection Pooling & Gin Engine
- **Goal:** Build the Go server foundation, connection pooling, and health endpoints.
- **Sub-Steps:**
  - [x] **Step 3.1: Go Module & Dependencies**
    - [x] Initialize Go module `github.com/MohammedaminSa/FindIt/backend`.
    - [x] Install `gin-gonic/gin`, `jackc/pgx/v5`, and `joho/godotenv`.
    - [x] Generate `go.mod` and `go.sum`.
  - [x] **Step 3.2: Configuration Management (`internal/config/config.go`)**
    - [x] Create `Config` struct mapping all `.env` variables.
    - [x] Implement safe environment loaders and default fallbacks.
    - [x] Construct `DatabaseURL()` connection DSN.
  - [x] **Step 3.3: PostgreSQL Connection Pooling (`internal/database/db.go`)**
    - [x] Configure `pgxpool.Pool` with `MaxConns`, `MinConns`, idle/life timeouts.
    - [x] Implement ping verification on startup.
  - [x] **Step 3.4: Health Handler & CORS Middleware (`internal/handler/health.go`, `internal/middleware/cors.go`)**
    - [x] Create `GET /api/v1/health` endpoint reporting uptime and database connectivity.
    - [x] Add CORS middleware for mobile and web cross-origin requests.
  - [x] **Step 3.5: Server Entrypoint & Graceful Shutdown (`cmd/api/main.go`)**
    - [x] Wire config, database pool, router, and middleware.
    - [x] Implement OS signal listener (`SIGINT`, `SIGTERM`) for zero-downtime shutdown.
  - [x] **Step 3.6: Live Verification & Testing**
    - [x] Compile binary and test `/api/v1/health`.
- **Verification:**
  ```bash
  cd backend
  go build -o bin/api.exe ./cmd/api
  ./bin/api.exe
  curl http://localhost:8080/api/v1/health
  ```

#### [ ] Step 4: User Authentication & Security Layer
- **Goal:** Implement secure user registration, login, and JWT-guarded routes.
- **Sub-Steps:**
  - [x] **Step 4.1: User Models & DTOs (`internal/models/user.go`)**
    - Define `User` model, `RegisterRequest`, `LoginRequest`, and `UserResponse`.
  - [x] **Step 4.2: Password Hashing (`pkg/hasher/hasher.go`)**
    - Implement `HashPassword` and `CheckPasswordHash` using `golang.org/x/crypto/bcrypt`.
  - [x] **Step 4.3: JWT Token Management (`pkg/token/jwt.go`)**
    - Implement `GenerateToken` and `ValidateToken` with HMAC-SHA256 signing.
  - [x] **Step 4.4: User Repository (`internal/repository/user_repository.go`)**
    - Implement database queries: `CreateUser`, `GetUserByEmail`, `GetUserByID`.
  - [x] **Step 4.5: Auth Service (`internal/service/auth_service.go`)**
    - Implement registration validation, duplicate email check, login verification, token generation.
  - [ ] **Step 4.6: Auth Handler & Routes (`internal/handler/auth.go`)**
    - Register endpoints: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`.
  - [ ] **Step 4.7: Auth Middleware (`internal/middleware/auth.go`)**
    - Intercept requests, validate `Authorization: Bearer <token>`, inject user context.
    - Add protected route: `GET /api/v1/auth/me`.
  - [ ] **Step 4.8: Verification & API Testing**
- **Verification:**
  ```bash
  # Register
  curl -X POST http://localhost:8080/api/v1/auth/register -H "Content-Type: application/json" -d "{\"name\":\"Alice\",\"email\":\"alice@example.com\",\"password\":\"password123\"}"
  # Login
  curl -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d "{\"email\":\"alice@example.com\",\"password\":\"password123\"}"
  # Get Profile (with token)
  curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/auth/me
  ```

---

### Phase 3: Core Domain Services (Lost & Found Features)

#### [ ] Step 5: Category & Item Reporting APIs
- **Goal:** Allow users to browse categories and post lost or found items with images.
- **Sub-Steps:**
  - [ ] **Step 5.1: Categories API (`GET /api/v1/categories`)**
    - Handler, service, and repository to list active categories.
  - [ ] **Step 5.2: Item Domain Models (`internal/models/item.go`)**
    - Define `Item`, `CreateItemRequest`, `ItemResponse`, status/type enums.
  - [ ] **Step 5.3: Image Upload Service (`internal/service/upload_service.go`)**
    - Validate file size and MIME types (JPEG, PNG, WebP).
    - Store files with unique UUID filenames in `uploads/` and serve statically.
  - [ ] **Step 5.4: Item Repository & Service (`internal/repository/item_repository.go`, `internal/service/item_service.go`)**
    - SQL queries: `CreateItem`, `GetItemByID`, `UpdateItemStatus`.
  - [ ] **Step 5.5: Item Endpoints (`POST /api/v1/items`, `GET /api/v1/items/:id`)**
- **Verification:**
  ```bash
  curl http://localhost:8080/api/v1/categories
  curl -X POST http://localhost:8080/api/v1/items -H "Authorization: Bearer <TOKEN>" -F "title=Lost iPhone 14" -F "type=LOST" -F "category_id=<UUID>"
  ```

#### [ ] Step 6: High-Performance Search & Filtering Engine
- **Goal:** Enable keyword search across titles/descriptions with faceted filtering.
- **Sub-Steps:**
  - [ ] **Step 6.1: GIN Full-Text Search Query Builder**
    - Execute `to_tsquery('english', $keyword)` queries using the GIN index.
  - [ ] **Step 6.2: Multi-Criteria Dynamic Filter**
    - Compound filters: `type` (LOST/FOUND), `category_id`, `city`, `status`, date ranges.
  - [ ] **Step 6.3: Pagination & Relevance Ranking**
    - Cursor/offset pagination (`limit`, `offset`) and ordering by `ts_rank`.
  - [ ] **Step 6.4: Search Endpoint (`GET /api/v1/items`)**
- **Verification:**
  ```bash
  curl "http://localhost:8080/api/v1/items?q=iPhone&type=LOST&page=1&limit=10"
  ```

#### [ ] Step 7: Ownership Claims Lifecycle & Verification Workflow
- **Goal:** Allow users to claim items, submit proof, and allow finders to approve/reject.
- **Sub-Steps:**
  - [ ] **Step 7.1: Claim Models & Enums (`internal/models/claim.go`)**
    - States: `PENDING`, `APPROVED`, `REJECTED`.
  - [ ] **Step 7.2: Claim Business Validations**
    - Finder cannot claim their own item; prevent duplicate active claims per user.
  - [ ] **Step 7.3: Claim Review Transaction**
    - Approving a claim transitions the item to `CLAIMED` and automatically sets all other pending claims to `REJECTED`.
  - [ ] **Step 7.4: Claim Endpoints**
    - `POST /api/v1/items/:id/claims` (submit claim with proof)
    - `GET /api/v1/items/:id/claims` (finder reviews claims)
    - `PATCH /api/v1/claims/:id/status` (approve/reject claim)
- **Verification:**
  ```bash
  curl -X POST http://localhost:8080/api/v1/items/<ITEM_ID>/claims -H "Authorization: Bearer <TOKEN>" -d "{\"proof_description\":\"The lock screen shows photo of my dog\"}"
  ```

#### [ ] Step 8: In-App Messaging & Notifications
- **Goal:** Enable secure 1-to-1 chat between finder and claimant, plus system alerts.
- **Sub-Steps:**
  - [ ] **Step 8.1: Message Model & Repository (`internal/models/message.go`)**
    - Store sender, receiver, item reference, content, timestamps.
  - [ ] **Step 8.2: Messaging Endpoints**
    - `POST /api/v1/items/:id/messages` (send message)
    - `GET /api/v1/items/:id/messages` (retrieve message history)
  - [ ] **Step 8.3: Notifications System**
    - Trigger notifications on: claim submitted, claim approved/rejected, message received.
    - `GET /api/v1/notifications` endpoint.
- **Verification:**
  ```bash
  curl -X POST http://localhost:8080/api/v1/items/<ITEM_ID>/messages -H "Authorization: Bearer <TOKEN>" -d "{\"content\":\"Can we meet at the campus library?\"}"
  ```

---

### Phase 4: Flutter Mobile Application

#### [ ] Step 9: Mobile Scaffolding & State Architecture
- **Goal:** Scaffold the Flutter project, architecture layers, HTTP client, and routing.
- **Sub-Steps:**
  - [ ] **Step 9.1: Flutter Project Initialization (`findit_mobile`)**
  - [ ] **Step 9.2: Feature-First Directory Layout**
  - [ ] **Step 9.3: Dio HTTP Client Setup with Interceptors**
  - [ ] **Step 9.4: State Management with Riverpod**
  - [ ] **Step 9.5: Declarative Navigation with GoRouter and Auth Guards**
  - [ ] **Step 9.6: Theme & Design System (Color palette, typography, reusable widgets)**

#### [ ] Step 10: Authentication UI & Auth State Flow
- **Goal:** Build login, registration, and persistent user session flow.
- **Sub-Steps:**
  - [ ] **Step 10.1: Login & Register Screens with Form Validation**
  - [ ] **Step 10.2: Secure Token Storage (`flutter_secure_storage`)**
  - [ ] **Step 10.3: Profile Screen & Logout Handling**

#### [ ] Step 11: Lost & Found Discovery Feed & Search UI
- **Goal:** Build the main feed screen with search, categories, and item details.
- **Sub-Steps:**
  - [ ] **Step 11.1: Segmented Feed (Lost Items / Found Items)**
  - [ ] **Step 11.2: Search Bar with Debouncing & Category Filter Chips**
  - [ ] **Step 11.3: Item Detail Screen with Image Carousel, Location Badge & Claim Action**

#### [ ] Step 12: Item Reporting Flow (Post Lost/Found)
- **Goal:** Enable users to post a lost or found item from mobile.
- **Sub-Steps:**
  - [ ] **Step 12.1: Post Item Multi-Step Form**
  - [ ] **Step 12.2: Camera & Image Gallery Picker Integration**
  - [ ] **Step 12.3: Location Coordinate & Date Picker**
  - [ ] **Step 12.4: Form Submission with Upload Progress & Error Feedback**

#### [ ] Step 13: Claim Submission & Claim Review Dashboard
- **Goal:** Enable claimants to submit proof and finders to review claims on mobile.
- **Sub-Steps:**
  - [ ] **Step 13.1: "Claim Item" Modal with Proof Input & Image Attachment**
  - [ ] **Step 13.2: Finder's Claim Review Screen**
  - [ ] **Step 13.3: One-Tap "Approve" & "Reject" Actions with Status Pills**

#### [ ] Step 14: Messaging & Chat Screen
- **Goal:** Build real-time 1-to-1 chat and notifications screen.
- **Sub-Steps:**
  - [ ] **Step 14.1: Chat Screen between Finder and Claimant**
  - [ ] **Step 14.2: Polling / Message Refresh Mechanism**
  - [ ] **Step 14.3: Notifications List Screen with Unread Badges**

---

### Phase 5: Production Deployment & DevOps

#### [ ] Step 15: Production Containerization & Cloud Deployment
- **Goal:** Prepare the full stack for production deployment.
- **Sub-Steps:**
  - [ ] **Step 15.1: Multi-Stage Dockerfile for Go Backend (~20MB alpine image)**
  - [ ] **Step 15.2: Production Docker Compose (`docker-compose.prod.yml`) with Nginx Reverse Proxy & SSL**
  - [ ] **Step 15.3: Production Environment Hardening, Backup Scripts & Health Monitoring**
