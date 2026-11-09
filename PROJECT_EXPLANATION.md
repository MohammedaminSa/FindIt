# FindIt Project - Complete Code Explanation Guide

**A Lost & Found Application Built with Go (Backend), Flutter (Mobile), and PostgreSQL (Database)**

This guide walks through every file in your project from entry point to end, explaining how the application starts, processes requests, and manages data.

---

## 📋 Table of Contents

1. [Project Overview](#project-overview)
2. [Application Entry Point - main.go](#application-entry-point---maingo)
3. [Configuration Management - config.go](#configuration-management---configgo)
4. [Database Layer - db.go](#database-layer---dbgo)
5. [Middleware - cors.go](#middleware---corsgo)
6. [Handlers - health.go](#handlers---healthgo)
7. [Database Schema - Migrations](#database-schema---migrations)
8. [Infrastructure - Docker Compose](#infrastructure---docker-compose)
9. [Complete Request Flow](#complete-request-flow)
10. [Key Go & PostgreSQL Concepts](#key-go--postgresql-concepts)

---

## 🎯 Project Overview

**FindIt** is a lost and found item tracking system where users can:
- Report lost items
- Report found items
- Claim items
- Message other users
- Get notifications

### Technology Stack:
- **Backend**: Go 1.25 with Gin web framework
- **Database**: PostgreSQL 16
- **Database Driver**: pgx/v5 (high-performance PostgreSQL driver)
- **Mobile**: Flutter (not yet in repo)
- **Infrastructure**: Docker Compose

---

## 🚀 Application Entry Point - main.go

**Location**: `backend/cmd/api/main.go`

This is where your Go application **starts executing**. Think of it as the "front door" of your backend.

### Complete Flow Explanation:

```go
package main
```
- Declares this is the `main` package
- Only the `main` package can create an executable program in Go
- The `main()` function inside will run when you execute the program

### Imports Section:

```go
import (
    "context"
    "errors"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/MohammedaminSa/FindIt/backend/internal/config"
    "github.com/MohammedaminSa/FindIt/backend/internal/database"
    "github.com/MohammedaminSa/FindIt/backend/internal/handler"
    "github.com/MohammedaminSa/FindIt/backend/internal/middleware"
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
)
```

**What each import does:**
- `context`: Manages request timeouts and cancellation (crucial for database operations)
- `errors`: Checks if errors match specific types
- `fmt`: Formats strings (like creating the server address)
- `log`: Prints messages to console for debugging
- `net/http`: HTTP protocol constants and types
- `os`: Operating system functions (reading environment variables, signals)
- `os/signal`: Listens for system signals (Ctrl+C to stop server)
- `syscall`: Low-level system calls (SIGINT, SIGTERM)
- `time`: Time durations and timeouts

**Your project packages:**
- `config`: Loads environment variables (.env file)
- `database`: Connects to PostgreSQL
- `handler`: Handles HTTP requests (like /health endpoint)
- `middleware`: Processes requests before they reach handlers (CORS)

**Third-party packages:**
- `gin-gonic/gin`: Web framework (like Express.js for Node)
- `pgx/v5/pgxpool`: PostgreSQL connection pool manager

---

### The setupRouter Function:

```go
func setupRouter(db handler.Pinger) *gin.Engine {
    router := gin.New()
```

**Purpose**: Creates and configures the HTTP router (decides which function handles which URL)

**Parameters:**
- `db handler.Pinger`: An interface that can ping the database
- Why interface? Allows testing without a real database

**Returns:**
- `*gin.Engine`: A pointer to the configured router

```go
    // Global Middlewares
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(middleware.CORS())
```

**Middleware Explanation:**
1. `gin.Logger()`: Logs every request (GET /api/v1/health, POST /api/v1/items, etc.)
2. `gin.Recovery()`: Catches panics (crashes) and returns a 500 error instead of crashing the server
3. `middleware.CORS()`: Your custom CORS handler (explained later)

**Middleware execution order**: Logger → Recovery → CORS → Handler

```go
    // Handlers
    healthHandler := handler.NewHealthHandler(db)
```
- Creates a health check handler instance
- Passes the database connection so it can verify DB is working

```go
    // API v1 Route Group
    v1 := router.Group("/api/v1")
    {
        v1.GET("/health", healthHandler.Check)
    }
```

**Route Group**: All routes inside this group start with `/api/v1`
- Full URL becomes: `http://localhost:8080/api/v1/health`
- `GET`: HTTP method (reading data, no modifications)
- `healthHandler.Check`: The function that responds to this request

**Why versioning (v1)?** When you make breaking changes later, you can add v2 routes while keeping v1 working for existing mobile apps.

---

### The main Function - Application Lifecycle:

```go
func main() {
```
**This is where execution begins** when you run `go run main.go`

#### **Step 1: Load Configuration**

```go
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load application configuration: %v", err)
    }
```

**What happens:**
1. Calls `config.Load()` (explained in next section)
2. Reads `.env` file and environment variables
3. If error, `log.Fatalf()` prints error and exits the program

**Error handling pattern in Go:**
```go
result, err := someFunction()
if err != nil {
    // Handle error
}
// Use result
```
Go doesn't have try/catch - errors are returned as second values

```go
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }
```
- In production mode: Disables debug logging for better performance
- In development: Shows detailed logs

---

#### **Step 2: Initialize PostgreSQL Connection Pool**

```go
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
```

**Context with timeout:**
- `context.Background()`: Creates a base context
- `WithTimeout(..., 10*time.Second)`: Adds a 10-second deadline
- If database doesn't respond in 10 seconds, the operation is cancelled
- `defer cancel()`: Ensures cleanup happens when function exits

**Why defer?** The `cancel()` function releases resources. `defer` means "run this before function returns"

```go
    var dbPool *pgxpool.Pool
    dbPool, err = database.NewPool(ctx, cfg)
    if err != nil {
        log.Printf("Warning: Database connection failed: %v (running in degraded mode)", err)
    } else {
        defer dbPool.Close()
        log.Println("PostgreSQL connection pool established successfully")
    }
```

**Connection Pool Concept:**
- Instead of creating a new database connection for every request (slow)
- Create 5-25 connections and reuse them (fast)
- Pool manages: opening, closing, health checks

**Degraded mode:**
- If DB connection fails, server still starts
- Health endpoint will show "database: DOWN"
- Better than crashing - allows debugging

---

#### **Step 3: Setup Router and Routes**

```go
    var pinger handler.Pinger
    if dbPool != nil {
        pinger = dbPool
    }
    router := setupRouter(pinger)
```

- If database connected successfully, pass it to router
- If not, pass `nil` (health check will handle this gracefully)

---

#### **Step 4: Configure HTTP Server**

```go
    serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
    srv := &http.Server{
        Addr:         serverAddr,           // ":8080"
        Handler:      router,               // Gin router
        ReadTimeout:  15 * time.Second,     // Max time to read request
        WriteTimeout: 15 * time.Second,     // Max time to write response
        IdleTimeout:  60 * time.Second,     // Max time between requests
    }
```

**Timeouts are crucial for production:**
- **ReadTimeout**: Prevents slow clients from holding connections
- **WriteTimeout**: Prevents long-running handlers from blocking
- **IdleTimeout**: Closes inactive connections to free resources

**Without timeouts:** Attackers could open thousands of slow connections and crash your server (Slowloris attack)

---

#### **Step 5: Start HTTP Server in Background**

```go
    go func() {
        log.Printf("FindIt API server listening on http://localhost:%s (%s mode)", 
                   cfg.ServerPort, cfg.Environment)
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            log.Fatalf("Server failed to listen: %v", err)
        }
    }()
```

**Goroutine (`go func()`):**
- Runs the function in the background (like a separate thread)
- Main function continues executing while server runs
- Why? So we can listen for shutdown signals below

**`srv.ListenAndServe()`:**
- Blocks and listens for HTTP requests
- Returns error if server fails to start
- Returns `http.ErrServerClosed` during graceful shutdown (not an error)

---

#### **Step 6: Graceful Shutdown**

```go
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
```

**Channel communication:**
- `make(chan os.Signal, 1)`: Creates a channel (like a pipe for messages)
- `signal.Notify(quit, ...)`: OS sends signals to this channel
- `<-quit`: Waits here until signal received (Ctrl+C or kill command)

**When does this trigger?**
- `SIGINT`: User presses Ctrl+C
- `SIGTERM`: Cloud platform wants to restart your app

```go
    log.Println("Shutdown signal received, shutting down gracefully...")

    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer shutdownCancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("FindIt API server gracefully terminated")
```

**Graceful shutdown process:**
1. Stop accepting new requests
2. Wait for in-flight requests to complete (max 5 seconds)
3. Close database connections
4. Exit cleanly

**Why important?**
- Prevents data corruption
- Doesn't drop user requests mid-processing
- Cloud platforms (AWS, GCP) expect this

---

## ⚙️ Configuration Management - config.go

**Location**: `backend/internal/config/config.go`

This file manages all settings (database credentials, server port, etc.) that change between environments (dev vs production).

### The Config Struct:

```go
type Config struct {
    // Database Connection Settings
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    DBSSLMode  string

    // Database Connection Pool Tuning
    DBMaxConns        int32
    DBMinConns        int32
    DBMaxConnLifetime int // in minutes
    DBMaxConnIdleTime int // in minutes

    // Application Server Settings
    ServerPort  string
    Environment string
    JWTSecret   string
}
```

**Struct in Go:** Like a class in other languages, but no inheritance
- Each field has a type (`string`, `int32`)
- Groups related configuration values

**Why separate database and server settings?**
- Easier to find what you need
- Can add more categories (Redis, S3, etc.) later

---

### The Load Function:

```go
func Load() (*Config, error) {
    _ = godotenv.Load(".env", "../.env")
```

**godotenv.Load():**
- Reads `.env` file from current directory or parent
- Loads variables into environment (accessible via `os.Getenv()`)
- `_` ignores the error (file might not exist in production using real env vars)

**Why try two locations?**
- If running from `backend/cmd/api/`, looks in `backend/cmd/api/.env`
- Then tries `backend/.env`

```go
    cfg := &Config{
        DBHost:            getEnv("DB_HOST", "localhost"),
        DBPort:            getEnv("DB_PORT", "5432"),
        DBUser:            getEnv("DB_USER", "findit_user"),
        // ... etc
    }
```

**`&Config{...}`**: Creates a Config struct and returns a pointer to it
- Pointer (`*Config`): Reference to memory location
- Why pointers? Efficient - doesn't copy entire struct

**getEnv() with defaults:**
- Tries to read from environment
- If not found or empty, uses default value
- Makes local development easy (no .env file needed)

---

### Helper Functions:

```go
func getEnv(key, fallback string) string {
    if val, ok := os.LookupEnv(key); ok && val != "" {
        return val
    }
    return fallback
}
```

**`os.LookupEnv()` vs `os.Getenv()`:**
- `LookupEnv()`: Returns value AND whether key exists
- Distinguishes between "not set" and "set to empty string"

**Multiple return values in Go:**
```go
value, exists := os.LookupEnv("DB_HOST")
if exists {
    // Use value
}
```

```go
func getEnvAsInt(key string, fallback int) int {
    valStr := getEnv(key, "")
    if val, err := strconv.Atoi(valStr); err == nil {
        return val
    }
    return fallback
}
```

**Type conversion:**
- Environment variables are always strings
- `strconv.Atoi()`: "ASCII to Integer" conversion
- If conversion fails, returns default

---

### The DatabaseURL Method:

```go
func (c *Config) DatabaseURL() string {
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s",
        c.DBUser,
        c.DBPassword,
        c.DBHost,
        c.DBPort,
        c.DBName,
        c.DBSSLMode,
    )
}
```

**Method syntax: `(c *Config)`**
- This function "belongs to" the Config struct
- Called like: `cfg.DatabaseURL()`
- `c` is the receiver (like `this` in other languages)

**PostgreSQL Connection String Format:**
```
postgres://username:password@host:port/database?sslmode=disable
```

Example output:
```
postgres://findit_user:findit_secret@localhost:5432/findit_db?sslmode=disable
```

**Why a method?**
- Encapsulation: Other code doesn't need to know URL format
- If format changes, only update here

---

## 🗄️ Database Layer - db.go

**Location**: `backend/internal/database/db.go`

This file creates and manages the PostgreSQL connection pool.

### NewPoolConfig Function:

```go
func NewPoolConfig(cfg *config.Config) (*pgxpool.Config, error) {
    poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
    if err != nil {
        return nil, fmt.Errorf("failed to parse database connection URL: %w", err)
    }
```

**`pgxpool.ParseConfig()`:**
- Takes the connection string
- Parses it into a structured configuration object
- Validates format

**Error wrapping (`%w`):**
```go
fmt.Errorf("failed to parse database connection URL: %w", err)
```
- Wraps the original error with additional context
- Allows error chain inspection
- Better debugging: "failed to parse database connection URL: invalid port '99999'"

```go
    poolConfig.MaxConns = cfg.DBMaxConns
    poolConfig.MinConns = cfg.DBMinConns
    poolConfig.MaxConnLifetime = time.Duration(cfg.DBMaxConnLifetime) * time.Minute
    poolConfig.MaxConnIdleTime = time.Duration(cfg.DBMaxConnIdleTime) * time.Minute
    poolConfig.HealthCheckPeriod = 1 * time.Minute
```

**Connection Pool Parameters Explained:**

| Parameter | Default | Purpose |
|-----------|---------|---------|
| MaxConns | 25 | Maximum connections in pool |
| MinConns | 5 | Minimum connections kept open |
| MaxConnLifetime | 60 min | Close and recreate after this time |
| MaxConnIdleTime | 30 min | Close if unused this long |
| HealthCheckPeriod | 1 min | How often to verify connections are alive |

**Why MaxConnLifetime?**
- PostgreSQL limits total connections (usually 100-200)
- Load balancers may reset connections periodically
- Prevents stale connections

**Why MinConns?**
- Keeps connections warm
- First request doesn't wait for connection creation
- Better performance under load

---

### NewPool Function:

```go
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
    poolConfig, err := NewPoolConfig(cfg)
    if err != nil {
        return nil, err
    }

    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize pgx connection pool: %w", err)
    }
```

**`pgxpool.NewWithConfig()`:**
- Creates the actual connection pool
- Opens MinConns connections immediately
- Returns error if database unreachable

```go
    pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := pool.Ping(pingCtx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("database ping failed: %w", err)
    }

    return pool, nil
}
```

**Ping verification:**
- Sends a simple query to PostgreSQL
- Verifies credentials are correct
- Verifies network connectivity
- Times out after 5 seconds

**Critical: `pool.Close()` on failure**
- If ping fails, connections are still open
- Must close them to avoid resource leak
- Go doesn't have try/finally - must manually handle cleanup

---

## 🛡️ Middleware - cors.go

**Location**: `backend/internal/middleware/cors.go`

CORS (Cross-Origin Resource Sharing) allows your Flutter app to make requests to your Go API.

### The Problem CORS Solves:

**Without CORS:**
```
Flutter app (http://localhost:3000) 
    → tries to call API (http://localhost:8080)
    → Browser blocks request (security)
```

**With CORS:**
```
Flutter app → API
    ← Response includes Access-Control-Allow-Origin: *
    → Browser allows request
```

### The CORS Middleware:

```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
```

**Function that returns a function:**
- `CORS()` returns a middleware function
- Pattern allows configuration (e.g., `CORS(allowedOrigins)`)

```go
        origin := c.Request.Header.Get("Origin")
        if origin == "" {
            origin = "*"
        }
```

**Origin header:**
- Browser automatically sends where request came from
- Examples: `http://localhost:3000`, `https://findlt.com`
- `*` means "allow from anywhere" (only for development)

```go
        c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")
```

**Header meanings:**

| Header | Purpose |
|--------|---------|
| Allow-Origin | Which domains can make requests |
| Allow-Credentials | Allow sending cookies/auth tokens |
| Allow-Headers | Which request headers are allowed |
| Allow-Methods | Which HTTP methods are allowed |
| Max-Age | Cache preflight response for 24 hours |

```go
        if c.Request.Method == http.MethodOptions {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
```

**Preflight requests:**
- Browser sends OPTIONS request first (before actual request)
- Asks "am I allowed to make this request?"
- We respond 204 (No Content) with headers
- Then browser sends real request

**`c.Next()`:**
- Continues to next middleware or handler
- Without this, request would stop here

---

## 🏥 Handlers - health.go

**Location**: `backend/internal/handler/health.go`

Health checks tell monitoring systems (Kubernetes, AWS) if your app is working.

### The Pinger Interface:

```go
type Pinger interface {
    Ping(ctx context.Context) error
}
```

**Interface in Go:**
- Defines behavior, not implementation
- Any type with `Ping(ctx) error` method satisfies this interface
- `*pgxpool.Pool` has this method, so it's automatically a Pinger

**Why use interface?**
- Testing: Create fake Pinger that always succeeds/fails
- Flexibility: Could use different database later

---

### HealthHandler Struct:

```go
type HealthHandler struct {
    db        Pinger
    startTime time.Time
}
```

**Storing start time:**
- Lets us calculate uptime
- Useful for knowing when server last restarted

```go
func NewHealthHandler(db Pinger) *HealthHandler {
    return &HealthHandler{
        db:        db,
        startTime: time.Now(),
    }
}
```

**Constructor pattern in Go:**
- Function that creates and initializes struct
- Conventionally named `New[StructName]`

---

### The isNilPinger Function:

```go
func isNilPinger(p Pinger) bool {
    if p == nil {
        return true
    }
    v := reflect.ValueOf(p)
    return v.Kind() == reflect.Ptr && v.IsNil()
}
```

**Interface nil check gotcha:**
```go
var db *pgxpool.Pool = nil
var pinger Pinger = db

pinger == nil  // false! (interface contains type info)
```

**Solution:**
- Use reflection to check if underlying pointer is nil
- `reflect.ValueOf()`: Inspects value at runtime
- `v.IsNil()`: Checks if pointer is nil

---

### HealthResponse Struct:

```go
type HealthResponse struct {
    Status    string            `json:"status"`
    Uptime    string            `json:"uptime"`
    Timestamp string            `json:"timestamp"`
    Services  map[string]string `json:"services"`
}
```

**JSON struct tags:**
- `` `json:"status"` ``: Field name in JSON response
- Go convention: CapitalCase field names
- JSON convention: lowercase field names

**Example response:**
```json
{
  "status": "UP",
  "uptime": "2h15m30s",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "database": "UP"
  }
}
```

---

### The Check Handler:

```go
func (h *HealthHandler) Check(c *gin.Context) {
    services := make(map[string]string)
    overallStatus := "UP"
    httpStatusCode := http.StatusOK
```

**`make()` function:**
- Creates and initializes map
- `map[string]string`: Keys are strings, values are strings
- Alternative: `services := map[string]string{}`

```go
    if !isNilPinger(h.db) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
        defer cancel()

        if err := h.db.Ping(ctx); err != nil {
            services["database"] = "DOWN"
            overallStatus = "DEGRADED"
            httpStatusCode = http.StatusServiceUnavailable
        } else {
            services["database"] = "UP"
        }
    } else {
        services["database"] = "NOT_CONFIGURED"
    }
```

**Health check logic:**
1. If database exists, ping it (2 second timeout)
2. If ping fails → status 503 (Service Unavailable)
3. If no database → NOT_CONFIGURED (still 200 OK)

**Why 2 second timeout?**
- Health checks run frequently (every 5-10 seconds)
- Don't want slow DB to make health checks slow
- Kubernetes might kill pod if health checks timeout

```go
    c.JSON(httpStatusCode, HealthResponse{
        Status:    overallStatus,
        Uptime:    time.Since(h.startTime).Truncate(time.Second).String(),
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Services:  services,
    })
}
```

**Response formatting:**
- `time.Since()`: Current time - start time = duration
- `.Truncate(time.Second)`: Removes milliseconds (cleaner)
- `.UTC()`: Always use UTC for APIs (consistency)
- `time.RFC3339`: ISO 8601 format (2024-01-15T10:30:00Z)

---

## 🗃️ Database Schema - Migrations

**Location**: `backend/migrations/`

Migrations are versioned database schema changes.

### 000001_init_schema.up.sql

This creates your database tables.

#### Enable UUID Extension:

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
```

**UUID (Universally Unique Identifier):**
- 128-bit random value
- Example: `550e8400-e29b-41d4-a716-446655440000`
- Almost impossible to collide (duplicate)
- Better than auto-incrementing IDs for distributed systems

**pgcrypto extension:**
- Adds `gen_random_uuid()` function
- PostgreSQL 13+ includes this by default

---

#### Users Table:

```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Column explanations:**

| Column | Type | Constraints | Purpose |
|--------|------|-------------|---------|
| id | UUID | PRIMARY KEY | Unique identifier |
| name | VARCHAR(100) | NOT NULL | User's display name |
| email | VARCHAR(255) | UNIQUE, NOT NULL | Login credential |
| password_hash | VARCHAR(255) | NOT NULL | Encrypted password |
| phone | VARCHAR(20) | - | Optional contact |
| avatar_url | TEXT | - | Profile picture URL |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT | When user registered |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT | Last profile update |

**Key SQL concepts:**
- `PRIMARY KEY`: Unique identifier, indexed automatically
- `UNIQUE`: No two users can have same email
- `NOT NULL`: Value required
- `DEFAULT`: Auto-fills if not provided
- `TIMESTAMPTZ`: Timestamp with timezone (always use this, not TIMESTAMP)

---

#### Categories Table:

```sql
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    icon VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**SERIAL type:**
- Auto-incrementing integer (1, 2, 3...)
- Good for categories (small, fixed set)
- Creates a sequence automatically

**Why not UUID here?**
- Categories are rarely created
- Integers take less space (4 bytes vs 16 bytes)
- Easier to work with in foreign keys

---

#### Items Table:

```sql
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('LOST', 'FOUND')),
    title VARCHAR(150) NOT NULL,
    description TEXT NOT NULL,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    location_name VARCHAR(255) NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    date_lost_or_found TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CLAIMED', 'RESOLVED')),
    image_urls TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Foreign keys:**
```sql
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
```
- `REFERENCES users(id)`: Must be valid user ID
- `ON DELETE CASCADE`: If user deleted, delete all their items
- Maintains referential integrity

**CHECK constraints:**
```sql
CHECK (type IN ('LOST', 'FOUND'))
```
- Database-level validation
- Prevents invalid data
- Better than trusting application code

**Array column:**
```sql
image_urls TEXT[] DEFAULT '{}'
```
- PostgreSQL supports arrays natively
- Stores multiple image URLs in one column
- `'{}'` is empty array literal
- Access in Go: `[]string{"url1", "url2"}`

**ON DELETE CASCADE vs RESTRICT:**
- `CASCADE`: Delete child records automatically
- `RESTRICT`: Prevent deletion if child records exist
- Category `RESTRICT`: Can't delete "Electronics" if items use it

---

#### Performance Indexes:

```sql
CREATE INDEX IF NOT EXISTS idx_items_type_status ON items (type, status);
CREATE INDEX IF NOT EXISTS idx_items_category ON items (category_id);
CREATE INDEX IF NOT EXISTS idx_items_created_at ON items (created_at DESC);
```

**Index purpose:**
- Speeds up queries on those columns
- Like book index - find pages quickly

**Composite index:**
```sql
idx_items_type_status ON items (type, status)
```
- Speeds up: `WHERE type = 'LOST' AND status = 'ACTIVE'`
- Also speeds up: `WHERE type = 'LOST'` (leftmost column)
- Does NOT speed up: `WHERE status = 'ACTIVE'` (not leftmost)

**Full-text search index:**
```sql
CREATE INDEX IF NOT EXISTS idx_items_search ON items 
USING gin (to_tsvector('english', title || ' ' || description));
```
- `GIN` index type for text search
- `to_tsvector()`: Converts text to searchable tokens
- Searches "lost iPhone 13" faster than `LIKE '%iPhone%'`
- Handles word variations (losing → lost)

---

#### Seed Data:

```sql
INSERT INTO categories (name, icon) VALUES
    ('Electronics', 'devices'),
    ('Wallets & Cards', 'account_balance_wallet'),
    ('Keys', 'vpn_key'),
    -- ...
ON CONFLICT (name) DO NOTHING;
```

**ON CONFLICT:**
- If category already exists (name is UNIQUE)
- Don't insert again, don't fail
- Safe for re-running migrations

---

### 000001_init_schema.down.sql

```sql
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS messages CASCADE;
DROP TABLE IF EXISTS claims CASCADE;
DROP TABLE IF EXISTS items CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS users CASCADE;
```

**Down migration:**
- Reverses the up migration
- Used for rolling back changes
- Must drop in reverse order (child tables first)

**CASCADE in DROP:**
- Also drops dependent objects (foreign keys)
- Without it, drop fails if dependencies exist

---

## 🐳 Infrastructure - Docker Compose

**Location**: `docker-compose.yml`

Docker Compose runs PostgreSQL in a container.

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: findit-postgres
    restart: unless-stopped
```

**Image explanation:**
- `postgres:16`: PostgreSQL version 16
- `alpine`: Minimal Linux distribution (smaller image size)
- `unless-stopped`: Auto-restart if crashes (not if manually stopped)

```yaml
    environment:
      POSTGRES_USER: ${DB_USER:-findit_user}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-findit_secret}
      POSTGRES_DB: ${DB_NAME:-findit_db}
```

**Environment variable substitution:**
- `${DB_USER:-findit_user}`: Use $DB_USER if set, else "findit_user"
- Reads from .env file in same directory

```yaml
    ports:
      - "${DB_PORT:-5432}:5432"
```

**Port mapping:**
- `host:container`
- Left side: Your computer
- Right side: Inside container
- `5432:5432` means localhost:5432 → container:5432

```yaml
    volumes:
      - pgdata:/var/lib/postgresql/data
```

**Volume:**
- Persistent storage
- Data survives container restarts
- Without this, deleting container deletes all data

```yaml
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-findit_user} -d ${DB_NAME:-findit_db}"]
      interval: 5s
      timeout: 5s
      retries: 5
```

**Health check:**
- Docker periodically runs `pg_isready` command
- Every 5 seconds
- Timeout after 5 seconds
- Try 5 times before marking unhealthy
- Your Go app can wait for healthy status before connecting

---

## 🔄 Complete Request Flow

Let's trace a request from start to finish:

### Example: GET /api/v1/health

```
1. Client Request
   ↓
   Flutter app sends: GET http://localhost:8080/api/v1/health

2. Network Layer
   ↓
   Request arrives at Go HTTP server (srv.ListenAndServe)

3. Gin Router
   ↓
   Matches route: v1.GET("/health", healthHandler.Check)

4. Middleware Chain
   ↓
   gin.Logger()  → Logs: "GET /api/v1/health"
   ↓
   gin.Recovery()  → Catches panics if handler crashes
   ↓
   middleware.CORS()  → Adds CORS headers
   ↓

5. Handler Execution
   ↓
   healthHandler.Check(c *gin.Context)
     ↓
     Creates context with 2-second timeout
     ↓
     Calls db.Ping(ctx)
       ↓
       PostgreSQL connection pool sends "SELECT 1"
       ↓
       PostgreSQL responds: Success
     ↓
     Builds HealthResponse{Status: "UP", ...}
     ↓
     c.JSON(200, response)

6. Response
   ↓
   Gin serializes struct to JSON
   ↓
   Adds Content-Type: application/json header
   ↓
   Sends HTTP response:

   HTTP/1.1 200 OK
   Content-Type: application/json
   Access-Control-Allow-Origin: *
   
   {
     "status": "UP",
     "uptime": "2h15m30s",
     "timestamp": "2024-01-15T10:30:00Z",
     "services": {
       "database": "UP"
     }
   }

7. Flutter app receives JSON and displays health status
```

---

## 🎓 Key Go & PostgreSQL Concepts

### Go Concepts:

#### 1. **Pointers**
```go
// Value copy (slow for large structs)
func processConfig(cfg Config) {
    // Changes don't affect original
}

// Pointer (fast, modifies original)
func processConfig(cfg *Config) {
    cfg.ServerPort = "9090"  // Original modified
}
```

#### 2. **Error Handling**
```go
// Go idiom: handle errors immediately
result, err := someFunction()
if err != nil {
    return fmt.Errorf("context: %w", err)
}
// Use result safely here
```

#### 3. **Goroutines**
```go
// Sequential
doTask1()
doTask2()

// Concurrent
go doTask1()  // Runs in background
go doTask2()  // Runs in background
```

#### 4. **Channels**
```go
quit := make(chan bool)

go func() {
    time.Sleep(5 * time.Second)
    quit <- true  // Send signal
}()

<-quit  // Wait for signal
```

#### 5. **Interfaces**
```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Any type with Read method is a Reader
// No explicit "implements" needed
```

#### 6. **Defer**
```go
func openFile() {
    f, _ := os.Open("file.txt")
    defer f.Close()  // Runs when function exits
    
    // Work with file
    // Close called automatically
}
```

### PostgreSQL Concepts:

#### 1. **ACID Properties**
- **Atomicity**: All or nothing (transaction succeeds or fails completely)
- **Consistency**: Data follows rules (constraints, foreign keys)
- **Isolation**: Concurrent transactions don't interfere
- **Durability**: Committed data survives crashes

#### 2. **Connection Pooling**
```
Without pool:          With pool:
Request → New conn     Request → Reuse conn
Request → New conn     Request → Reuse conn
Request → New conn     Request → Reuse conn
(Slow)                 (Fast)
```

#### 3. **Indexes**
```sql
-- Without index: Full table scan (slow)
SELECT * FROM items WHERE type = 'LOST';

-- With index: Direct lookup (fast)
CREATE INDEX idx_type ON items (type);
```

#### 4. **Foreign Keys**
```sql
-- Ensures data integrity
user_id REFERENCES users(id)

-- Can't insert item with non-existent user
-- Can't delete user with existing items (if RESTRICT)
```

---

## 📚 Learning Path

### To Run Your Project:

1. **Start PostgreSQL:**
   ```cmd
   docker-compose up -d
   ```

2. **Run migrations:**
   ```cmd
   cd backend
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   migrate -path migrations -database "postgres://findit_user:findit_secret@localhost:5432/findit_db?sslmode=disable" up
   ```

3. **Start Go server:**
   ```cmd
   cd backend/cmd/api
   go run main.go
   ```

4. **Test health endpoint:**
   ```cmd
   curl http://localhost:8080/api/v1/health
   ```

### Next Steps to Learn:

1. **Add more endpoints:**
   - POST /api/v1/items (create lost/found item)
   - GET /api/v1/items (list items with filters)
   - GET /api/v1/items/:id (get single item)

2. **Add authentication:**
   - JWT tokens
   - Password hashing with bcrypt
   - Protected routes

3. **Build Flutter app:**
   - Login/signup screens
   - List items with filters
   - Map view with pins
   - Chat between users

4. **Learn more patterns:**
   - Repository pattern (separate DB logic)
   - Dependency injection
   - Unit testing with mocks

---

## 🎯 Summary

Your FindIt project is a well-structured Go backend following industry best practices:

- ✅ **Clean Architecture**: Separated concerns (config, database, handlers, middleware)
- ✅ **Production-Ready**: Timeouts, graceful shutdown, health checks, connection pooling
- ✅ **Testable**: Interfaces (Pinger) allow mocking
- ✅ **Scalable**: Connection pooling, indexes, proper database design
- ✅ **Secure**: Password hashing, prepared statements (pgx prevents SQL injection)
- ✅ **Observable**: Health checks, logging, uptime tracking

The code flows from **main.go** → **config** → **database** → **router** → **middleware** → **handlers**, with each layer having a clear responsibility.

---

**Questions to explore:**
1. How would you add pagination to item listings?
2. How would you implement image upload to S3?
3. How would you add real-time notifications with WebSockets?
4. How would you implement geospatial search (find items near me)?

Happy coding! 🚀
