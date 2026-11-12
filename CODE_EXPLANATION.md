# FindIt - Complete Code Explanation for Go & PostgreSQL Beginners

**Author:** Learning Guide for FindIt Project  
**Audience:** Developers new to Go and PostgreSQL  
**Goal:** Understand every line of code in the FindIt backend

---

## Table of Contents

1. [Project Structure Overview](#project-structure-overview)
2. [Go Basics You Need to Know](#go-basics-you-need-to-know)
3. [PostgreSQL Basics](#postgresql-basics)
4. [Configuration Layer](#configuration-layer)
5. [Database Layer](#database-layer)
6. [Models Layer](#models-layer)
7. [Security Packages](#security-packages)
8. [Repository Layer](#repository-layer)
9. [Service Layer](#service-layer)
10. [Handler Layer](#handler-layer)
11. [Middleware Layer](#middleware-layer)
12. [Main Application](#main-application)
13. [Testing Explained](#testing-explained)

---

## Project Structure Overview

```
backend/
├── cmd/
│   └── api/
│       ├── main.go          # Application entry point (starts the server)
│       └── main_test.go     # Tests for routing setup
├── internal/                # Private application code (can't be imported by other projects)
│   ├── config/             # Configuration management (loading .env files)
│   ├── database/           # Database connection setup
│   ├── handler/            # HTTP handlers (controllers in MVC)
│   ├── middleware/         # HTTP middleware (CORS, Auth)
│   ├── models/             # Data structures (User, Item, etc.)
│   ├── repository/         # Database queries (SQL operations)
│   └── service/            # Business logic (rules and workflows)
├── pkg/                    # Public reusable packages
│   ├── hasher/            # Password hashing utilities
│   └── token/             # JWT token management
├── migrations/             # Database schema changes (SQL files)
├── go.mod                  # Go module definition (like package.json)
└── go.sum                  # Dependency lock file (like package-lock.json)
```

### Why This Structure?

- **`cmd/`**: Contains main applications. You could have multiple (api, cli, worker)
- **`internal/`**: Go enforces that other projects can't import this. Keeps your code private.
- **`pkg/`**: Reusable code that other projects could import
- **`migrations/`**: Version-controlled database schema changes

---

## Go Basics You Need to Know

### 1. Package System

```go
package main  // This file belongs to the "main" package
```

- Every Go file starts with `package <name>`
- Files in the same folder with the same package name can access each other's code
- `package main` is special - it creates an executable program
- Other package names create libraries

### 2. Imports

```go
import (
    "fmt"                    // Standard library package (formatting)
    "context"                // Standard library (request context)
    "github.com/user/repo"   // External package from GitHub
)
```

- `import` brings in other packages
- Standard library packages have short names (no domain)
- External packages have full paths (usually GitHub URLs)

### 3. Types

```go
// Basic types
var age int = 25
var name string = "Alice"
var isActive bool = true
var price float64 = 99.99

// Composite types
var numbers []int = []int{1, 2, 3}     // Slice (dynamic array)
var scores [5]int = [5]int{1,2,3,4,5}  // Array (fixed size)
var person map[string]string            // Map (dictionary/hash table)

// Custom types
type User struct {
    ID    string
    Name  string
    Email string
}
```

### 4. Pointers

```go
var x int = 5
var ptr *int = &x    // ptr holds the memory address of x

*ptr = 10            // Changes the value at that address
// Now x is 10
```

- `&` gets the address of a variable
- `*` dereferences (gets the value at an address)
- Pointers let functions modify the original variable

### 5. Functions

```go
// Basic function
func add(a int, b int) int {
    return a + b
}

// Multiple return values (common in Go for error handling)
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

// Using it
result, err := divide(10, 2)
if err != nil {
    // Handle error
}
```

### 6. Structs and Methods

```go
// Define a struct
type Rectangle struct {
    width  int
    height int
}

// Method on Rectangle (receiver is r)
func (r *Rectangle) Area() int {
    return r.width * r.height
}

// Using it
rect := Rectangle{width: 10, height: 5}
area := rect.Area()  // Calls the method
```

- `(r *Rectangle)` is the "receiver" - like `this` in other languages
- `*Rectangle` means the method can modify the struct

### 7. Interfaces

```go
// Define interface (contract)
type Shape interface {
    Area() int
}

// Any type with an Area() method implements Shape
type Circle struct {
    radius int
}

func (c *Circle) Area() int {
    return 3.14 * c.radius * c.radius
}

// Function that accepts any Shape
func PrintArea(s Shape) {
    fmt.Println(s.Area())
}
```

- Interfaces are implemented implicitly (no "implements" keyword)
- Any type with matching methods satisfies the interface

### 8. Error Handling

```go
// Go doesn't have try/catch. Instead, functions return errors:

result, err := someFunction()
if err != nil {
    // Handle the error
    log.Printf("Error: %v", err)
    return err
}
// Continue with result
```

- `error` is a built-in interface type
- `nil` means no error (null/undefined in other languages)
- Always check `err != nil` after function calls

### 9. Context

```go
ctx := context.Background()  // Create a context
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)  // Add timeout
defer cancel()  // Cleanup when function ends

// Pass context to functions
result, err := database.Query(ctx, "SELECT * FROM users")
```

- `context.Context` carries deadlines, cancellation signals, and request-scoped values
- Used for request tracing and cancellation in web apps
- Always pass as the first parameter to functions

### 10. Defer

```go
func processFile() {
    file, _ := os.Open("data.txt")
    defer file.Close()  // This runs when function exits
    
    // Do work with file
    // file.Close() will be called automatically
}
```

- `defer` schedules a function call to run when the surrounding function returns
- Used for cleanup (closing files, database connections, etc.)

---

## PostgreSQL Basics

### What is PostgreSQL?

PostgreSQL is a **relational database** - it stores data in tables with rows and columns, like spreadsheets with relationships.

### Key Concepts

#### 1. Tables

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(255),
    created_at TIMESTAMPTZ
);
```

- Table = collection of rows (like a spreadsheet)
- Each column has a type (UUID, VARCHAR, TIMESTAMPTZ, etc.)

#### 2. Primary Keys

```sql
id UUID PRIMARY KEY
```

- Uniquely identifies each row
- No two rows can have the same primary key
- We use UUID (Universally Unique Identifier) instead of integers for security

#### 3. Foreign Keys

```sql
CREATE TABLE items (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id)
);
```

- Links rows between tables
- `user_id` must exist in the `users` table
- Creates relationships (each item belongs to a user)

#### 4. Constraints

```sql
email VARCHAR(255) UNIQUE NOT NULL
type VARCHAR(10) CHECK (type IN ('LOST', 'FOUND'))
```

- `NOT NULL`: Column must have a value
- `UNIQUE`: No duplicates allowed
- `CHECK`: Value must meet a condition

#### 5. Indexes

```sql
CREATE INDEX idx_items_search ON items USING gin (to_tsvector('english', title || ' ' || description));
```

- Speeds up queries (like a book index)
- Without index: database scans every row
- With index: database jumps to matching rows instantly

#### 6. SQL Query Basics

```sql
-- SELECT: Read data
SELECT name, email FROM users WHERE id = '123';

-- INSERT: Add data
INSERT INTO users (id, name, email) VALUES ('123', 'Alice', 'alice@example.com');

-- UPDATE: Modify data
UPDATE users SET name = 'Bob' WHERE id = '123';

-- DELETE: Remove data
DELETE FROM users WHERE id = '123';
```

#### 7. Parameterized Queries (Security)

```go
// ❌ DANGEROUS (SQL Injection vulnerability)
query := "SELECT * FROM users WHERE email = '" + email + "'"

// ✅ SAFE (Parameterized query)
query := "SELECT * FROM users WHERE email = $1"
db.Query(ctx, query, email)
```

- `$1`, `$2`, etc. are placeholders
- Database treats parameters as data, not code
- Prevents SQL injection attacks

---

## Configuration Layer

**File:** `internal/config/config.go`

### Purpose
Load settings from environment variables (`.env` file) so passwords and secrets aren't hardcoded.

### Line-by-Line Explanation

```go
package config
```
- This file belongs to the `config` package
- All files in `internal/config/` with `package config` can see each other's code

```go
import (
    "fmt"
    "os"
    "github.com/joho/godotenv"
)
```
- `fmt`: Formatting (string operations, printing)
- `os`: Operating system functions (reading environment variables)
- `godotenv`: Third-party library to load `.env` files

```go
type Config struct {
    DBHost      string
    DBPort      string
    DBUser      string
    DBPassword  string
    DBName      string
    DBSSLMode   string
    ServerPort  string
    Environment string
    JWTSecret   string
}
```

**What is a struct?**
- A `struct` is a custom data type that groups related fields
- Like a class in other languages, but without methods (those come separately)
- `Config` holds all application settings in one place

**Field naming:**
- Uppercase first letter (e.g., `DBHost`) = exported (public)
- Lowercase first letter (e.g., `dbHost`) = unexported (private)

```go
func Load() (*Config, error) {
```

**Breaking this down:**
- `func`: Keyword to define a function
- `Load`: Function name (exported because it starts with uppercase)
- `()`: No input parameters
- `(*Config, error)`: Returns two things:
  - `*Config`: A pointer to a Config struct (the `*` means pointer)
  - `error`: An error if something goes wrong (or `nil` if successful)

```go
    _ = godotenv.Load()
```

**What's happening:**
- `godotenv.Load()` reads the `.env` file and sets environment variables
- `_` means "I don't care about this value" (blank identifier)
- Even if the `.env` file doesn't exist, we continue (environment variables might be set another way)

**Why the underscore?**
- `godotenv.Load()` returns an error, but we're ignoring it
- Go requires you to explicitly ignore return values with `_`

```go
    cfg := &Config{
        DBHost:      getEnv("DB_HOST", "localhost"),
        DBPort:      getEnv("DB_PORT", "5432"),
        DBUser:      getEnv("DB_USER", "postgres"),
        DBPassword:  getEnv("DB_PASSWORD", ""),
        DBName:      getEnv("DB_NAME", "findit_db"),
        DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
        ServerPort:  getEnv("SERVER_PORT", "8080"),
        Environment: getEnv("ENVIRONMENT", "development"),
        JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
    }
```

**Breaking this down:**
- `cfg :=` creates a new variable (`:=` is shorthand for declaration + assignment)
- `&Config{...}` creates a pointer to a new Config struct
- `&` means "get the memory address of this struct"
- Each line sets a field using `getEnv` (explained below)

**Why a pointer?**
- Passing around a pointer is cheaper than copying the entire struct
- Pointers let functions modify the original struct

```go
    return cfg, nil
```
- Returns the config pointer
- `nil` means "no error" (success)

```go
func getEnv(key string, defaultValue string) string {
```
- Private function (lowercase `g`) - only used within this package
- Takes two strings: the environment variable name and a fallback value
- Returns a string

```go
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
```

**Breaking this down:**
- `os.Getenv(key)`: Reads an environment variable
- `if value := ...; value != "" {`: Short variable declaration in if statement
  - Declares `value` and assigns the result
  - Then checks if `value` is not empty
- If the environment variable exists and isn't empty, return it
- Otherwise, return the default value

**Why this pattern?**
- Safe defaults let the app run even if `.env` is missing
- Production environments can override defaults with real values

```go
func (c *Config) DatabaseURL() string {
```

**Breaking this down:**
- `(c *Config)`: This is a **method** on the `Config` type
  - `c` is the receiver (like `this` or `self` in other languages)
  - `*Config` means it works with a pointer to Config
- `DatabaseURL()`: Method name
- `string`: Returns a string

**What's a method vs a function?**
- Function: `Load()` - standalone
- Method: `c.DatabaseURL()` - attached to a type, called on an instance

```go
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s",
        c.DBUser,
        c.DBPassword,
        c.DBHost,
        c.DBPort,
        c.DBName,
        c.DBSSLMode,
    )
```

**Breaking this down:**
- `fmt.Sprintf`: Creates a formatted string (like template strings in JavaScript)
- `%s`: Placeholder for a string
- `c.DBUser`: Access the DBUser field of the Config struct
- Builds a PostgreSQL connection string:
  - Format: `postgres://username:password@host:port/database?sslmode=disable`
  - Example: `postgres://findit_user:secret@localhost:5432/findit_db?sslmode=disable`

**Why build the URL here?**
- Single source of truth for database connection
- Easy to test and modify
- Hides complexity from other parts of the app

### Summary: config.go

1. **Loads** environment variables from `.env` file
2. **Provides** safe defaults if variables are missing
3. **Builds** the PostgreSQL connection string
4. **Exports** a clean `Config` struct for other packages to use

**Usage Example:**
```go
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}
dbURL := cfg.DatabaseURL()
// dbURL = "postgres://findit_user:password@localhost:5432/findit_db?sslmode=disable"
```

---

## Database Layer

**File:** `internal/database/db.go`

### Purpose
Manage PostgreSQL connections using a **connection pool** for efficiency and reliability.

### What is a Connection Pool?

Imagine a restaurant:
- **Without pool**: Every customer opens a new kitchen just for their order, then demolishes it (slow, wasteful)
- **With pool**: Restaurant keeps 10 kitchens ready. Customers borrow one, use it, return it (fast, efficient)

Database connections work the same way:
- Opening a connection is slow (TCP handshake, SSL negotiation, authentication)
- A pool keeps 10-20 connections ready to use
- Requests borrow a connection, execute queries, return it

### Line-by-Line Explanation

```go
package database
```
- Files in `internal/database/` belong to this package

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/MohammedaminSa/FindIt/backend/internal/config"
    "github.com/jackc/pgx/v5/pgxpool"
)
```

**Imports explained:**
- `context`: Request lifecycle management
- `time`: Duration and timeout handling
- `config`: Our custom config package (notice the full import path)
- `pgxpool`: PostgreSQL connection pool library (fastest Go driver)

```go
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
```

**Breaking this down:**
- `NewPool`: Constructor function (common Go pattern: `New` prefix)
- `ctx context.Context`: For cancellation and timeouts
- `cfg *config.Config`: Pointer to our config (needs database settings)
- Returns: `*pgxpool.Pool` (pointer to connection pool) and `error`

```go
    poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
    if err != nil {
        return nil, fmt.Errorf("unable to parse database URL: %w", err)
    }
```

**Breaking this down:**
- `pgxpool.ParseConfig()`: Converts URL string to pool configuration object
  - Input: `"postgres://user:pass@host:port/db?sslmode=disable"`
  - Output: Structured config object
- `if err != nil`: Check if parsing failed
- `return nil, fmt.Errorf(...)`: Return error if parsing failed
  - `nil` for the pool (no pool created)
  - `%w`: Wraps the original error (preserves error chain)

**Why wrap errors?**
```go
// Original error: "invalid port"
// Wrapped error: "unable to parse database URL: invalid port"
// More context = easier debugging
```

```go
    poolConfig.MaxConns = 25
    poolConfig.MinConns = 5
    poolConfig.MaxConnLifetime = 1 * time.Hour
    poolConfig.MaxConnIdleTime = 30 * time.Minute
    poolConfig.HealthCheckPeriod = 1 * time.Minute
```

**Connection pool tuning:**

| Setting | Value | Explanation |
|---------|-------|-------------|
| `MaxConns` | 25 | Maximum simultaneous connections to database |
| `MinConns` | 5 | Always keep 5 connections warm and ready |
| `MaxConnLifetime` | 1 hour | Close and replace connections after 1 hour (prevents stale connections) |
| `MaxConnIdleTime` | 30 min | Close unused connections after 30 minutes (frees resources) |
| `HealthCheckPeriod` | 1 min | Test connections every minute to detect database crashes |

**Why these numbers?**
- 25 max connections: Balance between performance and database load
  - Too few: Requests wait for available connections
  - Too many: Database overloads, performance degrades
- 5 min connections: Instant response for first few requests
- Lifetimes prevent connection degradation over time

```go
    pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
```

**Breaking this down:**
- `context.WithTimeout(ctx, 5*time.Second)`: Creates a new context that automatically cancels after 5 seconds
- Returns two things:
  - `pingCtx`: New context with timeout
  - `cancel`: Function to cancel early if needed
- `defer cancel()`: Ensures cleanup happens when function exits
  - `defer` runs at function end, regardless of how it exits (return, panic, etc.)

**Why timeout?**
- If database is down, don't wait forever
- Fail fast (5 seconds) instead of hanging indefinitely

```go
    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("unable to create connection pool: %w", err)
    }
```

**Breaking this down:**
- `pgxpool.NewWithConfig()`: Creates the actual connection pool
- Uses our custom `poolConfig` settings
- If creation fails (database unreachable, wrong credentials), return error

```go
    if err := pool.Ping(pingCtx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("unable to ping database: %w", err)
    }
```

**Breaking this down:**
- `pool.Ping(pingCtx)`: Test the connection by sending a simple query
- Uses `pingCtx` (5-second timeout) instead of main `ctx`
- If ping fails:
  1. `pool.Close()`: Clean up the pool (release resources)
  2. Return error (don't return a broken pool)

**Why ping?**
- Config might be correct, but database could be:
  - Not running
  - Rejecting connections (firewall)
  - Wrong version
- Better to fail during startup than during first request

```go
    return pool, nil
```
- Success! Return the working pool and no error

### Summary: db.go

1. **Parses** database connection URL
2. **Configures** connection pool settings (size, timeouts, health checks)
3. **Creates** the pool with 5-25 pre-warmed connections
4. **Validates** connectivity with a ping test
5. **Returns** ready-to-use pool or descriptive error

**Usage Example:**
```go
cfg, _ := config.Load()
pool, err := database.NewPool(ctx, cfg)
if err != nil {
    log.Fatal("Database unavailable:", err)
}
defer pool.Close()

// Now use pool for queries
row := pool.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", userID)
```

---

## Models Layer

**File:** `internal/models/user.go`

### Purpose
Define data structures (DTOs - Data Transfer Objects) for users, representing data as it flows through the application.

### Why Separate Models?

```
HTTP Request → Handler (uses DTOs) → Service (uses DTOs) → Repository (uses Models) → Database
```

- **DTOs**: Data shapes for API requests/responses (external-facing)
- **Models**: Data shapes matching database tables (internal)

### Line-by-Line Explanation

```go
package models
```
- All model definitions live in the `models` package

```go
import (
    "time"
    "github.com/google/uuid"
)
```
- `time`: For timestamp fields
- `uuid`: For UUID type (universally unique identifiers)

### User Struct (Database Model)

```go
type User struct {
    ID           uuid.UUID  `json:"id"`
    Name         string     `json:"name"`
    Email        string     `json:"email"`
    PasswordHash string     `json:"-"`
    Phone        *string    `json:"phone,omitempty"`
    AvatarURL    *string    `json:"avatar_url,omitempty"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}
```

**Field by field:**

| Field | Type | Meaning |
|-------|------|---------|
| `ID` | `uuid.UUID` | Unique identifier (128-bit random number) |
| `Name` | `string` | User's display name |
| `Email` | `string` | User's email address |
| `PasswordHash` | `string` | Bcrypt hash of password (NOT the actual password) |
| `Phone` | `*string` | Optional phone number (pointer = nullable) |
| `AvatarURL` | `*string` | Optional profile picture URL |
| `CreatedAt` | `time.Time` | When account was created |
| `UpdatedAt` | `time.Time` | Last modification time |

**Struct tags explained:**

```go
`json:"id"`
```

- Backticks (`` ` ``) define struct tags (metadata)
- `json:"id"`: When converting to JSON, use "id" as the key name
  - Without tag: `{"ID": "123"}` (ugly)
  - With tag: `{"id": "123"}` (clean, JavaScript-friendly)

```go
`json:"-"`
```
- `"-"` means "never include in JSON output"
- **Critical for security**: Prevents password hash from being sent to clients

```go
`json:"phone,omitempty"`
```
- `"phone,omitempty"`: If `Phone` is `nil` or empty, omit it from JSON
- Result: `{"name": "Alice"}` instead of `{"name": "Alice", "phone": null}`

**Why pointers for Phone and AvatarURL?**

```go
// Without pointer (cannot represent "no value")
type User struct {
    Phone string  // Empty string "" vs not provided? Ambiguous!
}

// With pointer (clear distinction)
type User struct {
    Phone *string  
    // nil = field not set
    // &"" = explicitly set to empty string
    // &"555-1234" = has value
}
```

In databases:
- `NULL` = no value
- `''` = empty string
- Pointers let Go represent `NULL` properly

### RegisterRequest (Input DTO)

```go
type RegisterRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,max=72"`
}
```

**Purpose:** Define what data clients must send when registering

**Struct tags - `binding`:**

```go
`json:"name" binding:"required,min=2,max=100"`
```

- `binding`: Validation rules (used by Gin framework)
  - `required`: Field must be present in JSON
  - `min=2`: Minimum 2 characters
  - `max=100`: Maximum 100 characters
  
```go
`json:"email" binding:"required,email"`
```
- `required`: Must be present
- `email`: Must match email format (contains @ and .)

```go
`json:"password" binding:"required,min=6,max=72"`
```
- `min=6`: Passwords must be at least 6 characters
- `max=72`: Bcrypt limitation (can't hash passwords > 72 bytes)

**Why validation tags?**

Without validation:
```go
// Client sends: {"name": "", "email": "not-an-email"}
// Server accepts it, tries to save to database
// Database rejects it with cryptic error
```

With validation:
```go
// Gin checks rules automatically
// Returns: {"error": "name is required, email is invalid"}
// Never touches database
```

### LoginRequest (Input DTO)

```go
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}
```

**Simpler than register:**
- Only needs email and password
- No name, phone, etc.
- Same validation rules

### UserResponse (Output DTO)

```go
type UserResponse struct {
    ID        uuid.UUID  `json:"id"`
    Name      string     `json:"name"`
    Email     string     `json:"email"`
    Phone     *string    `json:"phone,omitempty"`
    AvatarURL *string    `json:"avatar_url,omitempty"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}
```

**What's different from User?**
- **No `PasswordHash`**: Never send hashed passwords to clients
- Same fields otherwise
- Used for API responses (GET /auth/me, etc.)

**Why separate UserResponse from User?**

```go
// ❌ BAD: Accidentally expose password hash
user := getUserFromDB()
json.Marshal(user)  // Might include PasswordHash if json:"-" is forgotten

// ✅ GOOD: Conversion forces us to think about what to expose
func toUserResponse(user *User) *UserResponse {
    return &UserResponse{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        Phone:     user.Phone,
        AvatarURL: user.AvatarURL,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        // PasswordHash deliberately not copied
    }
}
```

### AuthResponse (Output DTO)

```go
type AuthResponse struct {
    Token string        `json:"token"`
    User  *UserResponse `json:"user"`
}
```

**Purpose:** Response after successful login/register

**Structure:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alice",
    "email": "alice@example.com",
    "created_at": "2024-11-11T10:00:00Z",
    "updated_at": "2024-11-11T10:00:00Z"
  }
}
```

**Why return both?**
- `token`: Client needs this for authenticated requests
- `user`: Client needs this to display profile immediately (avoids extra API call)

### Summary: models/user.go

| Type | Purpose | Usage |
|------|---------|-------|
| `User` | Database row structure | Repository layer |
| `RegisterRequest` | Register API input | Handler validation |
| `LoginRequest` | Login API input | Handler validation |
| `UserResponse` | API output | Handler responses |
| `AuthResponse` | Login/register output | Handler responses |

**Data Flow:**
```
1. Client sends RegisterRequest JSON
2. Handler validates with binding tags
3. Service creates User struct
4. Repository saves User to database
5. Service creates AuthResponse with token
6. Handler sends AuthResponse JSON to client
```

---

## Security Packages

### Password Hashing - `pkg/hasher/hasher.go`

#### Purpose
Securely hash passwords using bcrypt so they can't be stolen from the database.

#### Why Hash Passwords?

**Problem:**
```
Database breach → Attacker gets passwords → Can login to user accounts
```

**Solution:**
```
Database breach → Attacker gets hashes → Can't reverse them → Accounts safe
```

**How bcrypt works:**
1. User enters password: `"MySecretPass123"`
2. Bcrypt adds random salt: `"$2a$10$randomsalt"`
3. Runs 2^10 (1024) rounds of hashing
4. Stores: `"$2a$10$randomsalt...hashedoutput"`
5. Each password gets a different hash even if passwords are identical

#### Line-by-Line Explanation

```go
package hasher
```
- Public package (in `pkg/`) - other projects could import this

```go
import (
    "errors"
    "golang.org/x/crypto/bcrypt"
)
```
- `errors`: Create error objects
- `bcrypt`: Official Go crypto library for bcrypt

```go
var (
    ErrPasswordTooLong = errors.New("password exceeds 72 bytes")
)
```

**What's happening:**
- `var ( ... )`: Declares a package-level variable
- `ErrPasswordTooLong`: Exported error (uppercase = public)
- `errors.New("...")`: Creates an error with a message

**Why 72 bytes?**
- Bcrypt has a hard limit of 72 bytes
- Longer passwords get silently truncated
- We explicitly reject them instead

```go
const bcryptCost = 10
```
- `const`: Constant (value never changes)
- `bcryptCost = 10`: Bcrypt work factor
  - Higher = more secure but slower
  - 10 = 2^10 = 1024 iterations (good balance in 2024)
  - Each increment doubles the time

**Cost over time:**
```
Cost 4:  very fast, weak (early 2000s)
Cost 10: ~100ms, good (2020s) ✅ We use this
Cost 12: ~400ms, strong (high security apps)
Cost 15: ~3 seconds, overkill
```

```go
func HashPassword(password string) (string, error) {
```
- Takes plain password string
- Returns hashed password string and error

```go
    if len(password) > 72 {
        return "", ErrPasswordTooLong
    }
```
- `len(password)`: Length in bytes (not characters!)
- If too long, return empty string and error
- **Security note:** Prevents silent truncation vulnerability

```go
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    if err != nil {
        return "", err
    }
```

**Breaking this down:**
- `[]byte(password)`: Convert string to byte slice (bcrypt works with bytes)
- `bcrypt.GenerateFromPassword()`: The actual hashing
  - Generates random salt
  - Runs 1024 iterations
  - Returns hash as byte slice
- `if err != nil`: Check if hashing failed
  - Could fail if password is invalid UTF-8

```go
    return string(hash), nil
```
- `string(hash)`: Convert byte slice back to string
- `nil`: No error (success)

**What does the output look like?**
```
Input:  "MyPassword123"
Output: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
         ││││││├─────────────────┼───────────────────────────────────┤
         │││││ └ salt (16 bytes) └ hash (24 bytes)
         │││││
         │││││ cost (10 = 2^10 iterations)
         │││└─ algorithm version
         ││└── bcrypt identifier
         │└─── $ separators
         └──── hash format
```

```go
func CheckPasswordHash(password, hash string) bool {
```
- Takes plain password and stored hash
- Returns `true` if password matches, `false` otherwise
- No error return (comparison either matches or doesn't)

```go
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
```

**Breaking this down:**
- `[]byte(hash)`: Convert stored hash to bytes
- `[]byte(password)`: Convert entered password to bytes
- `bcrypt.CompareHashAndPassword()`:
  1. Extracts salt from hash
  2. Hashes password with same salt and cost
  3. Compares result with stored hash
- `err == nil`: If no error, passwords match

**Why this is secure:**
- Timing-safe comparison (doesn't reveal info through timing)
- Can't reverse the hash
- Each verification takes ~100ms (prevents brute force)

#### Summary: hasher.go

```go
// Registration
password := "UserPassword123"
hash, err := hasher.HashPassword(password)
// hash = "$2a$10$randomsalt...hashedoutput"
// Save hash to database (NOT the password)

// Login
enteredPassword := "UserPassword123"
storedHash := getUserHashFromDatabase()
if hasher.CheckPasswordHash(enteredPassword, storedHash) {
    // Password correct, login success
} else {
    // Password wrong, reject login
}
```

---

### JWT Token Management - `pkg/token/jwt.go`

#### Purpose
Create and validate JSON Web Tokens (JWT) for stateless authentication.

#### What is a JWT?

**Traditional session auth:**
```
Login → Server creates session → Stores session ID in database → Cookie with session ID
Every request → Look up session in database → Verify user
```

**JWT auth (stateless):**
```
Login → Server creates signed token with user info → Returns token
Every request → Verify token signature → Trust the user info inside
No database lookup needed!
```

**JWT Structure:**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
├────────── Header ─────────────┤├─────────── Payload ────────────┤├─── Signature ───┤
```

1. **Header:** Algorithm and token type
2. **Payload:** Claims (user data, expiration)
3. **Signature:** Proves the token wasn't tampered with

#### Line-by-Line Explanation

```go
package token
```

```go
import (
    "errors"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)
```

- `jwt/v5`: Standard JWT library
- `uuid`: For user IDs
- `time`: For expiration handling

#### Error Definitions

```go
var (
    ErrInvalidToken = errors.New("invalid or malformed token")
    ErrExpiredToken = errors.New("token has expired")
)
```

- Pre-defined errors for common cases
- Exported (uppercase) so other packages can check them with `errors.Is()`

#### TokenManager Interface

```go
type TokenManager interface {
    GenerateToken(userID uuid.UUID, email string) (string, error)
    ValidateToken(tokenString string) (*Claims, error)
}
```

**What's an interface?**
- Defines a contract: "Any type with these methods is a TokenManager"
- Enables dependency injection and testing

**Why use an interface?**
```go
// Production code uses JWTManager
var manager TokenManager = NewJWTManager(...)

// Test code uses MockTokenManager
var manager TokenManager = &MockTokenManager{}

// Both satisfy the interface, code works with both
```

#### Claims Struct

```go
type Claims struct {
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
    jwt.RegisteredClaims
}
```

**What are claims?**
- Data stored inside the JWT payload
- Anyone can read them (they're base64, not encrypted)
- But signature prevents tampering

**Fields:**
- `UserID`: Who the token belongs to
- `Email`: User's email (for convenience)
- `jwt.RegisteredClaims`: Standard claims (expiration, issuer, etc.)

**Embedded struct:**
```go
jwt.RegisteredClaims
```
- No field name means "embed" (inheritance-like)
- Claims now has all fields from RegisteredClaims:
  - `ExpiresAt`: When token expires
  - `IssuedAt`: When token was created
  - `NotBefore`: Token not valid before this time
  - `Issuer`: Who created the token
  - `Subject`: Who the token is about

#### JWTManager Struct

```go
type JWTManager struct {
    secretKey []byte
    duration  time.Duration
}
```

**Implementation of TokenManager interface:**
- `secretKey`: Secret string used to sign tokens (like a password)
- `duration`: How long tokens are valid (24 hours)

**Why []byte instead of string?**
- JWT library needs bytes
- Converting once (at creation) is more efficient

```go
func NewJWTManager(secretKey string, duration time.Duration) (*JWTManager, error) {
```

**Constructor pattern:**
- `New...` prefix is Go convention for constructors
- Returns pointer to new JWTManager

```go
    if len(secretKey) < 32 {
        return nil, errors.New("secret key must be at least 32 characters")
    }
```

**Security check:**
- Short secrets are weak (easy to brute force)
- Require at least 32 characters (256 bits)
- Example: `"my-super-secret-jwt-signing-key-32chars"`

```go
    return &JWTManager{
        secretKey: []byte(secretKey),
        duration:  duration,
    }, nil
```

- Create and return the manager
- `&` creates a pointer to the struct

#### GenerateToken Method

```go
func (m *JWTManager) GenerateToken(userID uuid.UUID, email string) (string, error) {
```

- `(m *JWTManager)`: Receiver (method belongs to JWTManager)
- `*JWTManager`: Pointer receiver (can read `m.secretKey` and `m.duration`)

```go
    now := time.Now()
    expiresAt := now.Add(m.duration)
```

- `time.Now()`: Current time
- `now.Add(m.duration)`: Current time + 24 hours = expiration time

```go
    claims := &Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expiresAt),
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
            Issuer:    "findit-api",
            Subject:   userID.String(),
        },
    }
```

**Building the claims:**
- Custom claims: `UserID`, `Email`
- Standard claims:
  - `ExpiresAt`: Token expires in 24 hours
  - `IssuedAt`: Created right now
  - `NotBefore`: Valid starting right now
  - `Issuer`: Our API identifier
  - `Subject`: User ID as string

**What's jwt.NewNumericDate?**
- Converts Go `time.Time` to Unix timestamp (seconds since 1970)
- JWT standard uses Unix timestamps

```go
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
```

**Creating the token:**
- `jwt.NewWithClaims()`: Create unsigned token
- `jwt.SigningMethodHS256`: Use HMAC-SHA256 algorithm
  - HMAC: Hash-based Message Authentication Code
  - SHA256: Hash function
  - Symmetric: Same key signs and verifies

**Why HS256?**
- Fast
- Secure (when key is long enough)
- Simple (one key for sign and verify)

Alternative: RS256 (asymmetric, public/private key pair) - overkill for most apps

```go
    return token.SignedString(m.secretKey)
```

**Signing the token:**
1. Serialize header and claims to JSON
2. Base64 encode them
3. Create signature: `HMAC-SHA256(header.payload, secretKey)`
4. Combine: `header.payload.signature`
5. Return as string

**Result example:**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIiwiZW1haWwiOiJhbGljZUBleGFtcGxlLmNvbSIsImV4cCI6MTY5OTk5OTk5OX0.signature
```

Decode the payload (middle part) with base64:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "exp": 1699999999
}
```

#### ValidateToken Method

```go
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
```

- Takes JWT string from client
- Returns claims if valid, error if invalid

```go
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
```

**Breaking this down:**
- `jwt.ParseWithClaims()`: Parse and validate token
- `tokenString`: The JWT from the client
- `&Claims{}`: Empty struct to fill with parsed data
- `func(token *jwt.Token) (interface{}, error)`: Callback function (explained below)

```go
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, ErrInvalidToken
        }
        return m.secretKey, nil
    })
```

**The callback function:**
- Called by parser to get the secret key
- **Security check**: Verify algorithm is HMAC
  - `token.Method`: Algorithm from token header
  - `(*jwt.SigningMethodHMAC)`: Type assertion (is it HMAC?)
  - `!ok`: If not HMAC, reject

**Why check the algorithm?**

**The "none" attack:**
```json
// Attacker creates token:
{
  "alg": "none",  // ← Claims no signature needed!
  "typ": "JWT"
}
{
  "user_id": "admin-user-id",
  "exp": 9999999999
}
```

Without algorithm check:
```go
// Server: "Algorithm is 'none', so no signature needed. Token valid!"
// Attacker gets admin access!
```

With algorithm check:
```go
// Server: "Algorithm must be HMAC. This is 'none'. REJECTED!"
```

```go
    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }
```

**Error handling:**
- `jwt.ErrTokenExpired`: Token was valid but expired
  - Specific error helps client know to refresh
- Any other error: Invalid token
  - Malformed, wrong signature, etc.

```go
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }
```

**Final validation:**
- `token.Claims.(*Claims)`: Type assertion (convert to Claims struct)
- `ok`: Did type assertion succeed?
- `token.Valid`: Is token marked as valid by parser?

Both must be true, or reject the token.

```go
    return claims, nil
```

- Success! Return the claims
- Caller can now trust `claims.UserID` and `claims.Email`

#### Helper Functions

```go
func GenerateToken(secretKey string, duration time.Duration, userID uuid.UUID, email string) (string, error) {
    manager, err := NewJWTManager(secretKey, duration)
    if err != nil {
        return "", err
    }
    return manager.GenerateToken(userID, email)
}
```

**Convenience function:**
- Combines `NewJWTManager` + `GenerateToken`
- Useful for one-off token generation

```go
func ValidateToken(secretKey string, tokenString string) (*Claims, error) {
    manager, err := NewJWTManager(secretKey, 24*time.Hour)
    if err != nil {
        return nil, err
    }
    return manager.ValidateToken(tokenString)
}
```

**Convenience function:**
- Combines `NewJWTManager` + `ValidateToken`
- Duration doesn't matter for validation (only checks expiration)

#### Summary: jwt.go

**Generate a token:**
```go
manager, _ := token.NewJWTManager("my-secret-key-32-characters-min", 24*time.Hour)
token, _ := manager.GenerateToken(userID, "alice@example.com")
// token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Validate a token:**
```go
claims, err := manager.ValidateToken(token)
if err != nil {
    // Invalid or expired
}
// claims.UserID = user's ID
// claims.Email = user's email
```

**Security guarantees:**
1. **Integrity**: Signature ensures token wasn't modified
2. **Expiration**: Old tokens automatically rejected
3. **Algorithm enforcement**: Prevents "none" attack
4. **Stateless**: No database lookup needed

---

*Continuing in next sections with Repository, Service, Handler, Middleware, and Main...*

Would you like me to continue with the remaining layers?

## Repository Layer

**File:** `internal/repository/user_repository.go`

### Purpose
The **data access layer** - the ONLY place that talks to PostgreSQL. Executes SQL queries and converts database rows to Go structs.

### Repository Pattern

**Why separate database code?**
```
❌ WITHOUT Repository:
Handler → Writes SQL → Executes query
Service → Writes SQL → Executes query
(SQL scattered everywhere, hard to test, hard to change)

✅ WITH Repository:
Handler → Service → Repository → Database
(All SQL in one place, easy to test with mocks, single source of truth)
```

### Line-by-Line Explanation

```go
package repository
```

```go
import (
    "context"
    "errors"
    "fmt"

    "github.com/MohammedaminSa/FindIt/backend/internal/models"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
)
```

**Imports:**
- `models`: Our user structs
- `uuid`: For user IDs
- `pgx/v5`: PostgreSQL driver (for query execution)
- `pgconn`: PostgreSQL errors (to detect constraint violations)

#### Error Definitions

```go
var (
    ErrUserNotFound  = errors.New("user not found")
    ErrDuplicateEmail = errors.New("email already in use")
)
```

**Domain errors:**
- Hide database details from upper layers
- Service layer checks `errors.Is(err, repository.ErrUserNotFound)`
- Don't leak "SQL error: no rows" to API responses

#### DBTX Interface

```go
type DBTX interface {
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}
```

**What's this interface for?**

Both `*pgxpool.Pool` and `pgx.Tx` (transaction) implement this interface:

```go
// Works with pool (normal queries)
repo := NewUserRepository(pool)

// Works with transaction (atomic operations)
tx, _ := pool.Begin(ctx)
repo := NewUserRepository(tx)
```

**Why this matters:**
- Same repository works with and without transactions
- Easy to test (can mock DBTX)

**Interface methods:**
- `QueryRow`: Execute SQL, expect 1 row back (SELECT with WHERE id = ...)
- `Exec`: Execute SQL, no rows back (INSERT, UPDATE, DELETE)
- `Query`: Execute SQL, expect multiple rows (SELECT * FROM ...)

#### UserRepository Interface

```go
type UserRepository interface {
    CreateUser(ctx context.Context, user *models.User) error
    GetUserByEmail(ctx context.Context, email string) (*models.User, error)
    GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}
```

**Why define an interface?**
- Service layer depends on this interface (not the concrete type)
- Tests can use a mock repository (no real database)
- Could swap PostgreSQL for MySQL without changing service code

#### PostgresUserRepository Struct

```go
type PostgresUserRepository struct {
    db DBTX
}
```

**Implementation of UserRepository:**
- `db`: Can be a pool or a transaction
- Struct holds the database connection

```go
func NewUserRepository(db DBTX) *PostgresUserRepository {
    return &PostgresUserRepository{db: db}
}
```

**Constructor:**
- Takes any `DBTX` (pool or transaction)
- Returns pointer to repository

#### CreateUser Method

```go
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *models.User) error {
```

**Signature breakdown:**
- `(r *PostgresUserRepository)`: Method on repository
- `ctx context.Context`: For cancellation and timeouts
- `user *models.User`: Pointer to user struct (will be modified)
- `error`: Returns error if something fails

```go
    user.ID = uuid.New()
```

**Generate UUID:**
- `uuid.New()`: Creates random v4 UUID
- Assigns it to the user struct
- Could also use database's `gen_random_uuid()`, but doing it here is more testable

```go
    query := `
        INSERT INTO users (id, name, email, password_hash, phone, avatar_url, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
```

**SQL query breakdown:**

| Part | Explanation |
|------|-------------|
| `INSERT INTO users` | Add row to users table |
| `(id, name, ...)` | List of columns we're filling |
| `VALUES ($1, $2, ...)` | Placeholders for actual values |
| `$1, $2, $3` | Positional parameters (prevents SQL injection) |
| `NOW()` | PostgreSQL function for current timestamp |
| `RETURNING ...` | Give back these values after insert |

**Why use $1, $2 instead of string concatenation?**

```go
// ❌ DANGEROUS - SQL Injection vulnerability
email := "'; DROP TABLE users; --"
query := "INSERT INTO users (email) VALUES ('" + email + "')"
// Results in: INSERT INTO users (email) VALUES (''); DROP TABLE users; --')
// Database executes: Delete all users!

// ✅ SAFE - Parameterized query
query := "INSERT INTO users (email) VALUES ($1)"
db.Exec(ctx, query, email)
// Database treats email as DATA, not CODE
// Even malicious input is safe
```

**Why RETURNING?**
- Database generates `created_at` and `updated_at` automatically
- We need those values to fill the Go struct
- `RETURNING` gives them back without a second query

```go
    err := r.db.QueryRow(ctx, query,
        user.ID,
        user.Name,
        user.Email,
        user.PasswordHash,
        user.Phone,
        user.AvatarURL,
    ).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
```

**Breaking this down:**

1. **`r.db.QueryRow(ctx, query, ...)`**
   - Execute the query
   - `ctx`: Context for cancellation/timeout
   - `query`: SQL statement
   - `...`: Values for $1, $2, $3, etc.
   
2. **Parameter order:**
   ```
   $1 = user.ID
   $2 = user.Name
   $3 = user.Email
   $4 = user.PasswordHash
   $5 = user.Phone
   $6 = user.AvatarURL
   ```

3. **`.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)`**
   - Read the `RETURNING` values
   - `&` means "write to this variable's address"
   - Fills in the fields we don't have yet

**Pointer magic:**
```go
// Before:
user.CreatedAt // Zero value (0001-01-01 00:00:00)

// After Scan:
user.CreatedAt // 2024-11-11 10:30:45 (from database)
```

```go
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return ErrDuplicateEmail
        }
        return fmt.Errorf("failed to create user: %w", err)
    }
```

**Error handling:**

1. **Check if error is a PostgreSQL error:**
   ```go
   errors.As(err, &pgErr)
   ```
   - `errors.As`: Type assertion for wrapped errors
   - Checks if `err` contains a `*pgconn.PgError`

2. **Check error code:**
   ```go
   pgErr.Code == "23505"
   ```
   - `23505`: PostgreSQL code for unique constraint violation
   - Means: Duplicate email (our UNIQUE constraint)

3. **Return domain error:**
   ```go
   return ErrDuplicateEmail
   ```
   - Service layer will check for this specific error
   - Handler can return 409 Conflict with proper message

**Why not check email existence first?**
```go
// ❌ Race condition vulnerable:
if emailExists(email) {
    return error
}
// Another request inserts same email here! ⚠️
insert(user)

// ✅ Let database enforce uniqueness:
insert(user)
// Database rejects duplicates atomically
```

```go
    return nil
```
- Success! User created and struct updated with ID and timestamps

#### GetUserByEmail Method

```go
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
```

**Purpose:** Find user by email (used during login)

```go
    query := `
        SELECT id, name, email, password_hash, phone, avatar_url, created_at, updated_at
        FROM users
        WHERE LOWER(TRIM(email)) = LOWER(TRIM($1))
    `
```

**SQL breakdown:**

| Part | Explanation |
|------|-------------|
| `SELECT id, name, ...` | Get these columns |
| `FROM users` | From users table |
| `WHERE ...` | Filter condition |
| `LOWER(TRIM(email))` | Normalize: lowercase + remove spaces |
| `LOWER(TRIM($1))` | Normalize search term the same way |

**Why normalize?**
```sql
-- Without normalization:
"alice@example.com" ≠ "Alice@example.com " ≠ " ALICE@EXAMPLE.COM"
(Same person, 3 different accounts! Bad!)

-- With normalization:
LOWER(TRIM("alice@example.com")) = "alice@example.com"
LOWER(TRIM("Alice@example.com ")) = "alice@example.com"
LOWER(TRIM(" ALICE@EXAMPLE.COM")) = "alice@example.com"
(All match! Good!)
```

**Database constraint should also normalize:**
```sql
CREATE UNIQUE INDEX idx_users_email ON users (LOWER(TRIM(email)));
```

```go
    var user models.User
    err := r.db.QueryRow(ctx, query, email).Scan(
        &user.ID,
        &user.Name,
        &user.Email,
        &user.PasswordHash,
        &user.Phone,
        &user.AvatarURL,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
```

**Breaking this down:**

1. **`var user models.User`**
   - Create empty user struct
   - Will be filled by `Scan()`

2. **`QueryRow(ctx, query, email)`**
   - Execute query
   - `$1` is replaced with `email` parameter

3. **`.Scan(&user.ID, &user.Name, ...)`**
   - Read columns into struct fields
   - Order must match SELECT order!
   
   ```
   Column 1 (id) → user.ID
   Column 2 (name) → user.Name
   Column 3 (email) → user.Email
   etc.
   ```

4. **Why pointers (`&`)?**
   - `Scan` needs to modify the variables
   - Pointers let it write to the original struct

```go
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }
```

**Error handling:**

1. **`pgx.ErrNoRows`**
   - No user found with that email
   - Return our domain error (`ErrUserNotFound`)
   - Service layer checks for this to distinguish "not found" from "database error"

2. **Other errors:**
   - Database connection lost
   - Malformed query
   - Permission denied
   - Wrap and return

```go
    return &user, nil
```
- Success! Return pointer to user struct

**Why pointer return?**
```go
// If user doesn't exist:
return nil, ErrUserNotFound  // nil pointer signals "no user"

// If user exists:
return &user, nil  // pointer to actual user data
```

#### GetUserByID Method

```go
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
```

**Purpose:** Find user by ID (used for profile fetching)

```go
    query := `
        SELECT id, name, email, password_hash, phone, avatar_url, created_at, updated_at
        FROM users
        WHERE id = $1
    `
```

**Simpler than GetUserByEmail:**
- UUID comparison is exact (no need for LOWER/TRIM)
- Primary key lookup (very fast with index)

```go
    var user models.User
    err := r.db.QueryRow(ctx, query, id).Scan(
        &user.ID,
        &user.Name,
        &user.Email,
        &user.PasswordHash,
        &user.Phone,
        &user.AvatarURL,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
```

Same pattern as GetUserByEmail:
1. Execute query with `id` parameter
2. Scan results into struct

```go
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user by id: %w", err)
    }

    return &user, nil
}
```

Same error handling as GetUserByEmail.

### Summary: repository/user_repository.go

**Responsibilities:**
1. ✅ Execute SQL queries safely (parameterized)
2. ✅ Convert database rows to Go structs (and vice versa)
3. ✅ Handle database-specific errors (unique violations, not found)
4. ✅ Return domain errors (not database errors)

**Design patterns:**
- **Repository pattern**: Isolate data access
- **Interface**: Enable testing and swappable implementations
- **DBTX**: Support both pools and transactions

**Security:**
- ✅ Parameterized queries (no SQL injection)
- ✅ Email normalization (prevent duplicate accounts)
- ✅ UUID primary keys (prevent enumeration attacks)

---

## Service Layer

**File:** `internal/service/auth_service.go`

### Purpose
The **business logic layer** - orchestrates workflows, enforces rules, and coordinates between repositories and external services.

### Clean Architecture: The Service Layer

```
┌─────────────┐
│   Handler   │  "The Waiter" - Takes orders, serves responses
└──────┬──────┘
       │
┌──────▼──────┐
│   Service   │  "The Chef" - Business rules and workflows ← YOU ARE HERE
└──────┬──────┘
       │
┌──────▼──────┐
│ Repository  │  "The Pantry" - Database access only
└──────┬──────┘
       │
┌──────▼──────┐
│  Database   │
└─────────────┘
```

**Service layer rules:**
- ✅ Contains business logic (password validation, duplicate checks)
- ✅ Coordinates multiple repositories or services
- ✅ Returns domain models (not database rows)
- ❌ Never touches HTTP (no request/response objects)
- ❌ Never writes SQL (delegates to repository)

### Line-by-Line Explanation

```go
package service
```

```go
import (
    "context"
    "errors"
    "fmt"

    "github.com/MohammedaminSa/FindIt/backend/internal/models"
    "github.com/MohammedaminSa/FindIt/backend/internal/repository"
    "github.com/MohammedaminSa/FindIt/backend/pkg/hasher"
    "github.com/MohammedaminSa/FindIt/backend/pkg/token"
    "github.com/google/uuid"
)
```

**Dependencies:**
- `models`: User DTOs
- `repository`: Data access (our interface)
- `hasher`: Password hashing
- `token`: JWT generation
- `uuid`: User IDs

#### Error Definitions

```go
var (
    ErrEmailAlreadyInUse = errors.New("email already in use")
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrUserNotFound       = errors.New("user not found")
)
```

**Service-level errors:**
- Hide repository errors from handlers
- Provide user-friendly messages
- `ErrInvalidCredentials`: Generic message (security - doesn't reveal if email exists)

#### AuthService Interface

```go
type AuthService interface {
    Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error)
    Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error)
    GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error)
}
```

**Why an interface?**
- Handlers depend on interface (not concrete implementation)
- Easy to mock for testing
- Can swap implementations (add caching, rate limiting, etc.)

#### DefaultAuthService Struct

```go
type DefaultAuthService struct {
    userRepo     repository.UserRepository
    tokenManager token.TokenManager
}
```

**Dependencies (injected):**
- `userRepo`: Database access (could be real or mock)
- `tokenManager`: JWT creation (could be real or mock)

**Dependency injection = testability:**
```go
// Production:
service := NewAuthService(postgresRepo, jwtManager)

// Testing:
service := NewAuthService(mockRepo, mockTokenManager)
```

```go
func NewAuthService(userRepo repository.UserRepository, tokenManager token.TokenManager) *DefaultAuthService {
    return &DefaultAuthService{
        userRepo:     userRepo,
        tokenManager: tokenManager,
    }
}
```

**Constructor:**
- Takes dependencies as parameters
- Stores them in struct
- Returns pointer to service

#### Register Method

```go
func (s *DefaultAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
```

**Signature:**
- `ctx`: Request context (cancellation, timeout)
- `req`: Registration data from client (validated by handler)
- Returns: Auth response (token + user) or error

**Business flow:**
1. Check if email already exists
2. Hash the password
3. Create user in database
4. Generate JWT token
5. Return token and user profile

```go
    existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
    if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
        return nil, fmt.Errorf("failed to check existing user: %w", err)
    }
```

**Pre-check for duplicate email:**

1. **Try to find user by email:**
   ```go
   existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
   ```

2. **Handle possible outcomes:**
   ```go
   if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
   ```
   
   This condition means:
   - `err != nil`: Something went wrong
   - `!errors.Is(err, repository.ErrUserNotFound)`: But it's NOT "user not found"
   - So it must be a real error (database down, network issue, etc.)

3. **Error is NOT "not found":**
   ```go
   return nil, fmt.Errorf("failed to check existing user: %w", err)
   ```
   - Real database error, propagate it upwards

**Truth table:**
| Condition | existingUser | err | Action |
|-----------|-------------|-----|--------|
| User exists | `*User` | `nil` | Reject (duplicate) |
| User not found | `nil` | `ErrUserNotFound` | Continue (good!) |
| DB error | `nil` | Other error | Abort (return error) |

```go
    if existingUser != nil {
        return nil, ErrEmailAlreadyInUse
    }
```

**Duplicate detection:**
- If user was found, email is taken
- Return service-level error
- Handler will convert to 409 Conflict response

**Why check here AND in repository?**
```go
// Defense in depth:
1. Service checks (early rejection, better error message)
2. Database constraint (handles race conditions)

// Race condition example:
Request A: Check email (not found) → continues
Request B: Check email (not found) → continues
Request A: Insert user ✓
Request B: Insert user ✗ (database constraint catches it)
```

```go
    passwordHash, err := hasher.HashPassword(req.Password)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }
```

**Password hashing:**
- Convert plain password to bcrypt hash
- Never store plain passwords!
- If hashing fails (password > 72 bytes), return error

```go
    user := &models.User{
        Name:         req.Name,
        Email:        req.Email,
        PasswordHash: passwordHash,
        Phone:        nil,
        AvatarURL:    nil,
    }
```

**Create user model:**
- Fill in provided fields
- `Phone` and `AvatarURL` are `nil` (optional fields)
- `ID`, `CreatedAt`, `UpdatedAt` will be set by database

```go
    if err := s.userRepo.CreateUser(ctx, user); err != nil {
        if errors.Is(err, repository.ErrDuplicateEmail) {
            return nil, ErrEmailAlreadyInUse
        }
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
```

**Database insertion:**

1. **Call repository:**
   ```go
   s.userRepo.CreateUser(ctx, user)
   ```
   - Executes INSERT query
   - Updates `user` with ID and timestamps

2. **Handle duplicate error:**
   ```go
   if errors.Is(err, repository.ErrDuplicateEmail) {
       return nil, ErrEmailAlreadyInUse
   }
   ```
   - Caught race condition!
   - Return same error as pre-check

3. **Handle other errors:**
   ```go
   return nil, fmt.Errorf("failed to create user: %w", err)
   ```

```go
    jwtToken, err := s.tokenManager.GenerateToken(user.ID, user.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to generate token: %w", err)
    }
```

**Generate JWT:**
- Create token with user's ID and email
- Token valid for 24 hours (configured in manager)
- If generation fails (shouldn't happen), return error

```go
    return &models.AuthResponse{
        Token: jwtToken,
        User: &models.UserResponse{
            ID:        user.ID,
            Name:      user.Name,
            Email:     user.Email,
            Phone:     user.Phone,
            AvatarURL: user.AvatarURL,
            CreatedAt: user.CreatedAt,
            UpdatedAt: user.UpdatedAt,
        },
    }, nil
```

**Build response:**
- `Token`: JWT string
- `User`: Profile data (WITHOUT password hash!)

**Why separate UserResponse?**
```go
// ❌ Accidentally leak password hash:
return &models.AuthResponse{
    Token: token,
    User:  user,  // Has PasswordHash field!
}

// ✅ Explicit mapping:
return &models.AuthResponse{
    Token: token,
    User:  &models.UserResponse{
        // Only safe fields, no PasswordHash
    },
}
```

#### Login Method

```go
func (s *DefaultAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
```

**Business flow:**
1. Find user by email
2. Verify password matches hash
3. Generate JWT token
4. Return token and user profile

```go
    user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
    if err != nil {
        if errors.Is(err, repository.ErrUserNotFound) {
            return nil, ErrInvalidCredentials
        }
        return nil, fmt.Errorf("failed to retrieve user: %w", err)
    }
```

**Find user:**

1. **Query database:**
   ```go
   user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
   ```

2. **User not found:**
   ```go
   if errors.Is(err, repository.ErrUserNotFound) {
       return nil, ErrInvalidCredentials
   }
   ```
   
   **🔒 Security note:** Return generic error!
   - ❌ Don't say "email not found" (helps attackers enumerate accounts)
   - ✅ Say "invalid email or password" (ambiguous)

3. **Database error:**
   ```go
   return nil, fmt.Errorf("failed to retrieve user: %w", err)
   ```

**Account enumeration attack prevention:**
```go
// ❌ BAD: Reveals which emails exist
if userNotFound {
    return "Email not found"
}
if wrongPassword {
    return "Wrong password"
}

// Attacker: "Aha! alice@example.com exists but bob@example.com doesn't"

// ✅ GOOD: Same error for both cases
if userNotFound || wrongPassword {
    return "Invalid email or password"
}

// Attacker: "I learned nothing useful"
```

```go
    if !hasher.CheckPasswordHash(req.Password, user.PasswordHash) {
        return nil, ErrInvalidCredentials
    }
```

**Password verification:**

1. **Compare entered password with stored hash:**
   ```go
   hasher.CheckPasswordHash(req.Password, user.PasswordHash)
   ```
   - Hashes `req.Password` with same salt
   - Compares result with `user.PasswordHash`
   - Returns `true` if match, `false` if not

2. **Wrong password:**
   ```go
   if !matches {
       return nil, ErrInvalidCredentials
   }
   ```
   
   **🔒 Security:** Same error as "user not found"

**Why not just compare hashes directly?**
```go
// ❌ WRONG:
if hasher.HashPassword(req.Password) == user.PasswordHash {
    // This will NEVER match! Bcrypt generates new salt each time
}

// ✅ RIGHT:
if hasher.CheckPasswordHash(req.Password, user.PasswordHash) {
    // Bcrypt extracts salt from hash and compares correctly
}
```

```go
    jwtToken, err := s.tokenManager.GenerateToken(user.ID, user.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to generate token: %w", err)
    }
```

**Generate JWT** (same as Register)

```go
    return &models.AuthResponse{
        Token: jwtToken,
        User: &models.UserResponse{
            ID:        user.ID,
            Name:      user.Name,
            Email:     user.Email,
            Phone:     user.Phone,
            AvatarURL: user.AvatarURL,
            CreatedAt: user.CreatedAt,
            UpdatedAt: user.UpdatedAt,
        },
    }, nil
}
```

**Return response** (same as Register)

#### GetUserProfile Method

```go
func (s *DefaultAuthService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error) {
```

**Purpose:** Fetch user profile by ID (for /auth/me endpoint)

```go
    user, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        if errors.Is(err, repository.ErrUserNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user profile: %w", err)
    }
```

**Fetch user:**
- Query by ID (from JWT token)
- Convert repository error to service error
- Propagate database errors

```go
    return &models.UserResponse{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        Phone:     user.Phone,
        AvatarURL: user.AvatarURL,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }, nil
}
```

**Return profile** (without password hash)

### Summary: service/auth_service.go

**Responsibilities:**
1. ✅ Orchestrate business workflows (register, login)
2. ✅ Enforce business rules (duplicate emails, password verification)
3. ✅ Coordinate dependencies (repository + token manager)
4. ✅ Return domain errors (hide implementation details)
5. ✅ Prevent security vulnerabilities (account enumeration)

**Design patterns:**
- **Service layer pattern**: Business logic separation
- **Dependency injection**: Testable, flexible
- **Error wrapping**: Context-rich error messages

**Security measures:**
- ✅ Password hashing (bcrypt)
- ✅ Account enumeration protection (generic errors)
- ✅ Defense in depth (pre-check + database constraint)

---

*Continuing with Handler Layer, Middleware Layer, and Main Application...*

Let me know if you want me to continue with the remaining sections!

## Handler Layer

**File:** `internal/handler/auth.go`

### Purpose
The **presentation layer** - handles HTTP requests, validates input, calls services, and returns HTTP responses.

### MVC Equivalent
In MVC (Model-View-Controller) architecture:
- **Handler = Controller** (processes requests, coordinates actions)
- **Service = Model** (business logic)
- **JSON responses = View** (presentation format)

### Line-by-Line Explanation

```go
package handler
```

```go
import (
    "errors"
    "net/http"

    "github.com/MohammedaminSa/FindIt/backend/internal/models"
    "github.com/MohammedaminSa/FindIt/backend/internal/service"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)
```

**Imports:**
- `net/http`: HTTP status codes (200, 401, 500, etc.)
- `models`: DTOs for requests/responses
- `service`: Business logic interface
- `gin`: Web framework
- `uuid`: For user IDs

#### AuthHandler Struct

```go
type AuthHandler struct {
    authService service.AuthService
}
```

**Handler structure:**
- Holds reference to auth service
- Service does the work, handler just coordinates

```go
func NewAuthHandler(authService service.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}
```

**Constructor:**
- Dependency injection (receives service)
- Returns pointer to handler

#### RegisterRoutes Method

```go
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware ...gin.HandlerFunc) {
```

**Signature breakdown:**
- `rg *gin.RouterGroup`: Group of routes (e.g., `/api/v1`)
- `authMiddleware ...gin.HandlerFunc`: Variadic parameter (0 or more middleware functions)

**Variadic parameter:**
```go
// Can be called with:
h.RegisterRoutes(rg)                  // No middleware
h.RegisterRoutes(rg, authMw)          // One middleware
h.RegisterRoutes(rg, mw1, mw2)        // Multiple middleware
```

```go
    auth := rg.Group("/auth")
    {
        auth.POST("/register", h.Register)
        auth.POST("/login", h.Login)
        if len(authMiddleware) > 0 && authMiddleware[0] != nil {
            auth.GET("/me", authMiddleware[0], h.Me)
        } else {
            auth.GET("/me", h.Me)
        }
    }
}
```

**Route registration:**

1. **Create subgroup:**
   ```go
   auth := rg.Group("/auth")
   ```
   - If `rg` is `/api/v1`, then `auth` is `/api/v1/auth`

2. **Register public routes:**
   ```go
   auth.POST("/register", h.Register)
   auth.POST("/login", h.Login)
   ```
   - `POST /api/v1/auth/register` → `h.Register()`
   - `POST /api/v1/auth/login` → `h.Login()`

3. **Register protected route:**
   ```go
   if len(authMiddleware) > 0 && authMiddleware[0] != nil {
       auth.GET("/me", authMiddleware[0], h.Me)
   ```
   
   **Middleware chain:**
   ```
   Request → authMiddleware[0] (validate JWT) → h.Me (get profile)
   ```
   
   If middleware rejects (invalid token), request never reaches `h.Me`

4. **Fallback without middleware:**
   ```go
   } else {
       auth.GET("/me", h.Me)
   }
   ```
   - For testing without authentication
   - Production always passes middleware

**Curly braces trick:**
```go
auth := rg.Group("/auth")
{  // ← Not required, but groups routes visually
    auth.POST("/register", h.Register)
    auth.POST("/login", h.Login)
}
```
- Pure style/readability
- No functional difference

#### Register Handler

```go
func (h *AuthHandler) Register(c *gin.Context) {
```

**Signature:**
- `c *gin.Context`: Gin's context object (contains request, response, params, etc.)
- No return value (response written directly to `c`)

**Handler responsibilities:**
1. Parse and validate request body
2. Call service
3. Map service errors to HTTP status codes
4. Return JSON response

```go
    var req models.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation failed",
            "details": err.Error(),
        })
        return
    }
```

**Request binding and validation:**

1. **Declare variable:**
   ```go
   var req models.RegisterRequest
   ```

2. **Bind JSON body to struct:**
   ```go
   c.ShouldBindJSON(&req)
   ```
   
   **What happens:**
   - Reads JSON from request body
   - Parses JSON
   - Fills `req` struct fields
   - Validates `binding` tags (required, email, min, max)

3. **Validation tags reminder:**
   ```go
   type RegisterRequest struct {
       Name     string `json:"name" binding:"required,min=2,max=100"`
       Email    string `json:"email" binding:"required,email"`
       Password string `json:"password" binding:"required,min=6,max=72"`
   }
   ```

4. **Validation failure:**
   ```go
   if err != nil {
       c.JSON(http.StatusBadRequest, gin.H{...})
       return
   }
   ```
   
   - `http.StatusBadRequest`: 400 error
   - `gin.H{...}`: Shorthand for `map[string]interface{}{...}`
   - Returns JSON: `{"error": "Validation failed", "details": "Name is required"}`

**Example validation errors:**
```json
// Missing name:
{"error": "Validation failed", "details": "Key: 'RegisterRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"}

// Invalid email:
{"error": "Validation failed", "details": "Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"}

// Password too short:
{"error": "Validation failed", "details": "Key: 'RegisterRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag"}
```

```go
    res, err := h.authService.Register(c.Request.Context(), &req)
    if err != nil {
        if errors.Is(err, service.ErrEmailAlreadyInUse) {
            c.JSON(http.StatusConflict, gin.H{
                "error": err.Error(),
            })
            return
        }

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to register user",
        })
        return
    }
```

**Call service:**

1. **Invoke business logic:**
   ```go
   res, err := h.authService.Register(c.Request.Context(), &req)
   ```
   - `c.Request.Context()`: HTTP request context (for cancellation)
   - `&req`: Pointer to validated request

2. **Handle duplicate email:**
   ```go
   if errors.Is(err, service.ErrEmailAlreadyInUse) {
       c.JSON(http.StatusConflict, gin.H{...})
       return
   }
   ```
   - `http.StatusConflict`: 409 error (resource conflict)
   - Return specific error message
   - Client knows to try different email

3. **Handle other errors:**
   ```go
   c.JSON(http.StatusInternalServerError, gin.H{
       "error": "Failed to register user",
   })
   ```
   - `http.StatusInternalServerError`: 500 error
   - Generic message (don't leak internal errors to client)
   - Log the actual error (in production, add logging here)

**HTTP status code mapping:**
| Service Error | HTTP Status | Meaning |
|--------------|-------------|---------|
| `ErrEmailAlreadyInUse` | 409 Conflict | Email taken |
| Any other error | 500 Internal Server Error | Something went wrong |
| No error | 201 Created | Success |

```go
    c.JSON(http.StatusCreated, res)
}
```

**Success response:**
- `http.StatusCreated`: 201 (resource created)
- `res`: AuthResponse struct
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Alice",
      "email": "alice@example.com",
      "created_at": "2024-11-11T10:00:00Z",
      "updated_at": "2024-11-11T10:00:00Z"
    }
  }
  ```

#### Login Handler

```go
func (h *AuthHandler) Login(c *gin.Context) {
```

**Similar structure to Register, with different errors**

```go
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation failed",
            "details": err.Error(),
        })
        return
    }
```

**Bind and validate** (same pattern)

```go
    res, err := h.authService.Login(c.Request.Context(), &req)
    if err != nil {
        if errors.Is(err, service.ErrInvalidCredentials) {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": err.Error(),
            })
            return
        }

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to authenticate user",
        })
        return
    }
```

**Call service with different error handling:**

| Service Error | HTTP Status | Meaning |
|--------------|-------------|---------|
| `ErrInvalidCredentials` | 401 Unauthorized | Wrong email/password |
| Any other error | 500 Internal Server Error | Something went wrong |
| No error | 200 OK | Success |

**Why 401 instead of 404 or 403?**
- `401 Unauthorized`: Credentials required or invalid ✅
- `403 Forbidden`: Authenticated but not authorized (wrong use case)
- `404 Not Found`: Resource doesn't exist (would reveal if email exists)

```go
    c.JSON(http.StatusOK, res)
}
```

**Success response:**
- `http.StatusOK`: 200 (standard success)
- Same AuthResponse as Register

#### Me Handler (Get Profile)

```go
func (h *AuthHandler) Me(c *gin.Context) {
```

**Purpose:** Return authenticated user's profile

```go
    val, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
```

**Extract user ID from context:**

1. **Middleware sets context:**
   ```go
   // In auth middleware:
   c.Set("userID", claims.UserID)
   ```

2. **Handler reads context:**
   ```go
   val, exists := c.Get("userID")
   ```
   - `exists`: Did middleware set this value?
   - If not, middleware didn't run (shouldn't happen if routes configured correctly)

3. **No user ID:**
   ```go
   if !exists {
       c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
       return
   }
   ```

```go
    userID, ok := val.(uuid.UUID)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
        return
    }
```

**Type assertion:**

1. **Convert interface{} to UUID:**
   ```go
   val.(uuid.UUID)
   ```
   - `val` is type `interface{}` (can hold anything)
   - We know it should be `uuid.UUID`
   - Type assertion checks and converts

2. **Assertion syntax:**
   ```go
   userID, ok := val.(uuid.UUID)
   ```
   - `userID`: Converted value (if successful)
   - `ok`: Boolean (true if type matches, false otherwise)

3. **Wrong type:**
   ```go
   if !ok {
       c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
       return
   }
   ```
   - Shouldn't happen (middleware sets correct type)
   - Defense against bugs

**Why this is needed:**
```go
// Context stores values as interface{}
c.Set("userID", uuid.New())  // Stored as interface{}

// Later:
val, _ := c.Get("userID")
// val is interface{}, not uuid.UUID!
// Can't call uuid.String() on it

// Must convert:
userID := val.(uuid.UUID)  // Now it's uuid.UUID
fmt.Println(userID.String())  // ✓ Works
```

```go
    profile, err := h.authService.GetUserProfile(c.Request.Context(), userID)
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve profile"})
        return
    }
```

**Fetch profile:**

| Service Error | HTTP Status | Meaning |
|--------------|-------------|---------|
| `ErrUserNotFound` | 404 Not Found | User doesn't exist |
| Any other error | 500 Internal Server Error | Something went wrong |
| No error | 200 OK | Success |

**User not found scenario:**
- JWT valid but user deleted from database
- Rare, but possible

```go
    c.JSON(http.StatusOK, profile)
}
```

**Success response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Alice",
  "email": "alice@example.com",
  "created_at": "2024-11-11T10:00:00Z",
  "updated_at": "2024-11-11T10:00:00Z"
}
```

### Summary: handler/auth.go

**Responsibilities:**
1. ✅ Parse and validate HTTP requests
2. ✅ Call service layer
3. ✅ Map service errors to HTTP status codes
4. ✅ Return JSON responses

**HTTP Status Codes:**
| Code | Meaning | When Used |
|------|---------|-----------|
| 200 OK | Success | Login, Get Profile |
| 201 Created | Resource created | Register |
| 400 Bad Request | Invalid input | Validation failed |
| 401 Unauthorized | Auth required/failed | Invalid credentials, No token |
| 404 Not Found | Resource missing | User not found |
| 409 Conflict | Resource conflict | Email already exists |
| 500 Internal Server Error | Server error | Database down, etc. |

**Design patterns:**
- **Thin handlers**: Logic in service, coordination in handler
- **Error mapping**: Service errors → HTTP statuses
- **Validation**: Early rejection of bad input

---

## Middleware Layer

### CORS Middleware - `internal/middleware/cors.go`

#### Purpose
Allow cross-origin requests from browser-based clients (React, Flutter web, etc.)

#### What is CORS?

**The problem:**
```
Your Frontend: http://localhost:3000
Your Backend:  http://localhost:8080

Browser: "These are different origins! I'll block the request for security."
```

**The solution:**
```
Backend sends headers:
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Methods: GET, POST, PUT, DELETE
Access-Control-Allow-Headers: Content-Type, Authorization

Browser: "Okay, backend says it's fine. Request allowed!"
```

#### Line-by-Line Explanation

```go
package middleware
```

```go
import (
    "github.com/gin-gonic/gin"
)
```

```go
func CORS() gin.HandlerFunc {
```

**Middleware function:**
- Returns a `gin.HandlerFunc` (function that processes requests)
- Will be used like: `router.Use(middleware.CORS())`

```go
    return func(c *gin.Context) {
```

**Closure:**
- Returns an anonymous function
- This function runs on every request

```go
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
```

**Allow any origin:**
- `"*"`: Any domain can make requests
- Production: Set to specific domain (e.g., `"https://findit.com"`)

```go
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
```

**Allow cookies:**
- Enables sending cookies/auth credentials cross-origin
- Needed for cookie-based auth (we use JWT in headers, so technically optional)

```go
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
```

**Allowed request headers:**
- Client can send these headers
- Important: `Authorization` (for JWT tokens)

```go
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
```

**Allowed HTTP methods:**
- Client can use GET, POST, PUT, DELETE, PATCH
- `OPTIONS`: Preflight requests (explained below)

```go
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
```

**Preflight request handling:**

**What's a preflight?**
```
1. Browser: "I want to POST to /api/v1/auth/login with Authorization header"
2. Browser sends OPTIONS request first (preflight)
3. Server: "Yes, POST and Authorization are allowed"
4. Browser: "Okay, now I'll send the real POST request"
```

- `OPTIONS`: HTTP method for preflight
- `204 No Content`: Empty successful response
- `c.AbortWithStatus(204)`: Stop here, don't call next handlers

```go
        c.Next()
    }
}
```

**Continue to next handler:**
- For non-OPTIONS requests, continue to actual handler

**Request flow:**
```
Request → CORS middleware (sets headers) → Next middleware → Handler
```

### Auth Middleware - `internal/middleware/auth.go`

#### Purpose
Validate JWT tokens and protect routes from unauthorized access.

#### Line-by-Line Explanation

```go
package middleware
```

```go
import (
    "errors"
    "net/http"
    "strings"

    "github.com/MohammedaminSa/FindIt/backend/pkg/token"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)
```

#### Constants

```go
const (
    AuthorizationHeaderKey = "Authorization"
    AuthorizationTypeBearer = "bearer"
    ContextUserIDKey = "userID"
    ContextUserEmailKey = "userEmail"
    ContextClaimsKey = "claims"
)
```

**Why constants?**
- Avoid typos: `"Authorization"` vs `"Autorization"` 
- Refactoring: Change in one place
- Autocomplete: Editor helps with naming

#### Auth Middleware Function

```go
func Auth(tokenManager token.TokenManager) gin.HandlerFunc {
```

**Middleware factory:**
- Takes token manager (for JWT validation)
- Returns middleware function
- Enables dependency injection

```go
    return func(c *gin.Context) {
```

**The middleware function** (runs on each request)

```go
        authHeader := c.GetHeader(AuthorizationHeaderKey)
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "authorization header is required",
            })
            return
        }
```

**Check for Authorization header:**

1. **Read header:**
   ```go
   c.GetHeader("Authorization")
   ```
   - Gets value of `Authorization` header
   - Example: `"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

2. **Missing header:**
   ```go
   if authHeader == "" {
       c.AbortWithStatusJSON(401, gin.H{"error": "..."})
       return
   }
   ```
   - `AbortWithStatusJSON`: Stop processing, return error
   - 401 Unauthorized
   - Request never reaches handler

```go
        fields := strings.Fields(authHeader)
        if len(fields) != 2 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid authorization header format",
            })
            return
        }
```

**Parse header format:**

1. **Split by whitespace:**
   ```go
   strings.Fields("Bearer token123")
   // Returns: []string{"Bearer", "token123"}
   ```

2. **Validate format:**
   ```go
   if len(fields) != 2 {
       // Error: Must be exactly 2 words
   }
   ```

**Valid formats:**
```
✓ "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
✗ "Bearer" (only 1 word)
✗ "Bearer token extra" (3 words)
✗ "token" (only 1 word, missing Bearer)
```

```go
        authType := strings.ToLower(fields[0])
        if authType != AuthorizationTypeBearer {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "unsupported authorization type",
            })
            return
        }
```

**Validate authorization type:**

1. **Extract type:**
   ```go
   authType := strings.ToLower(fields[0])
   // "Bearer" → "bearer"
   // "BEARER" → "bearer"
   // "bearer" → "bearer"
   ```

2. **Case-insensitive check:**
   ```go
   if authType != "bearer" {
       // Reject
   }
   ```

**Valid types:**
```
✓ "Bearer token"
✓ "bearer token"
✓ "BEARER token"
✗ "Basic dXNlcjpwYXNz" (HTTP Basic auth)
✗ "Token abc123" (custom scheme)
```

```go
        tokenString := fields[1]
        claims, err := tokenManager.ValidateToken(tokenString)
        if err != nil {
            if errors.Is(err, token.ErrExpiredToken) {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                    "error": "token has expired",
                })
                return
            }

            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid or malformed token",
            })
            return
        }
```

**Validate JWT:**

1. **Extract token:**
   ```go
   tokenString := fields[1]
   ```

2. **Validate signature and claims:**
   ```go
   claims, err := tokenManager.ValidateToken(tokenString)
   ```
   - Checks signature (hasn't been tampered with)
   - Checks expiration (not expired)
   - Checks algorithm (is HMAC)

3. **Handle expired token:**
   ```go
   if errors.Is(err, token.ErrExpiredToken) {
       c.AbortWithStatusJSON(401, gin.H{
           "error": "token has expired",
       })
       return
   }
   ```
   - Specific error message
   - Client knows to refresh token

4. **Handle invalid token:**
   ```go
   c.AbortWithStatusJSON(401, gin.H{
       "error": "invalid or malformed token",
   })
   ```
   - Wrong signature
   - Malformed format
   - Wrong algorithm

```go
        c.Set(ContextUserIDKey, claims.UserID)
        c.Set(ContextUserEmailKey, claims.Email)
        c.Set(ContextClaimsKey, claims)
```

**Store user info in context:**

1. **Set values:**
   ```go
   c.Set("userID", claims.UserID)
   c.Set("userEmail", claims.Email)
   c.Set("claims", claims)
   ```

2. **Handlers can access:**
   ```go
   func SomeHandler(c *gin.Context) {
       userID, _ := c.Get("userID")
       // Now handler knows who the user is
   }
   ```

```go
        c.Next()
    }
}
```

**Continue to handler:**
- Token valid, continue processing
- Handler can trust the user is authenticated

#### Helper Functions

```go
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
    val, exists := c.Get(ContextUserIDKey)
    if !exists {
        return uuid.Nil, false
    }
    id, ok := val.(uuid.UUID)
    return id, ok
}
```

**Type-safe context getter:**

1. **Get value:**
   ```go
   val, exists := c.Get("userID")
   ```

2. **Check existence:**
   ```go
   if !exists {
       return uuid.Nil, false
   }
   ```
   - `uuid.Nil`: Zero value (all zeros)
   - `false`: Signal "not found"

3. **Type assert:**
   ```go
   id, ok := val.(uuid.UUID)
   return id, ok
   ```
   - Convert interface{} to uuid.UUID
   - `ok`: true if correct type

**Usage:**
```go
func Handler(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        // Shouldn't happen if middleware ran
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }
    
    // Use userID
    user := getUserFromDB(userID)
}
```

```go
func GetUserEmail(c *gin.Context) (string, bool) {
    val, exists := c.Get(ContextUserEmailKey)
    if !exists {
        return "", false
    }
    email, ok := val.(string)
    return email, ok
}
```

**Get user email** (same pattern)

```go
func GetClaims(c *gin.Context) (*token.Claims, bool) {
    val, exists := c.Get(ContextClaimsKey)
    if !exists {
        return nil, false
    }
    claims, ok := val.(*token.Claims)
    return claims, ok
}
```

**Get full claims** (same pattern)

### Summary: middleware/auth.go

**Middleware flow:**
```
1. Request arrives with "Authorization: Bearer <token>"
2. Extract token from header
3. Validate format ("Bearer <token>")
4. Validate token (signature, expiration)
5. Store user info in context
6. Call next handler (c.Next())
```

**Protection:**
```go
// Without middleware:
router.GET("/api/v1/items", handler.GetItems)
// Anyone can access

// With middleware:
protected := router.Group("")
protected.Use(middleware.Auth(tokenManager))
protected.GET("/api/v1/items", handler.GetItems)
// Only authenticated users
```

---

## Main Application

**File:** `cmd/api/main.go`

### Purpose
Application entry point - wires everything together and starts the server.

### Line-by-Line Explanation

```go
package main
```

**Special package:**
- `package main`: Creates executable
- `package <other>`: Creates library

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
    "github.com/MohammedaminSa/FindIt/backend/internal/repository"
    "github.com/MohammedaminSa/FindIt/backend/internal/service"
    "github.com/MohammedaminSa/FindIt/backend/pkg/token"
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
)
```

**All our dependencies**

#### setupRouter Function

```go
func setupRouter(db handler.Pinger, authHandler *handler.AuthHandler, authMiddleware ...gin.HandlerFunc) *gin.Engine {
```

**Router factory:**
- `db handler.Pinger`: For health checks
- `authHandler`: Auth endpoints
- `authMiddleware`: JWT validation middleware
- Returns: Configured Gin engine

```go
    router := gin.New()
```

**Create Gin instance:**
- `gin.New()`: Empty router (no default middleware)
- `gin.Default()`: Router with Logger and Recovery (alternative)

```go
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(middleware.CORS())
```

**Global middleware:**

1. **Logger:**
   ```
   [GIN] 2024/11/11 - 10:30:45 | 200 | 15.2ms | 127.0.0.1 | POST /api/v1/auth/login
   ```

2. **Recovery:**
   - Catches panics (crashes)
   - Returns 500 error instead of crashing server
   - Logs stack trace

3. **CORS:**
   - Allows cross-origin requests
   - Sets CORS headers on all responses

**Middleware order matters:**
```
Request → Logger → Recovery → CORS → Route Handler
```

```go
    healthHandler := handler.NewHealthHandler(db)
```

**Create health check handler:**
- Needs database connection for health checks

```go
    var authMw gin.HandlerFunc
    if len(authMiddleware) > 0 {
        authMw = authMiddleware[0]
    }
```

**Extract middleware:**
- Variadic parameters → Get first one
- Could be `nil` if not provided

```go
    v1 := router.Group("/api/v1")
    {
        v1.GET("/health", healthHandler.Check)
        if authHandler != nil {
            authHandler.RegisterRoutes(v1, authMw)
        }
    }
```

**Route registration:**

1. **Create route group:**
   ```go
   v1 := router.Group("/api/v1")
   ```
   - All routes in this group start with `/api/v1`

2. **Register health check:**
   ```go
   v1.GET("/health", healthHandler.Check)
   ```
   - `GET /api/v1/health`

3. **Register auth routes:**
   ```go
   authHandler.RegisterRoutes(v1, authMw)
   ```
   - `POST /api/v1/auth/register`
   - `POST /api/v1/auth/login`
   - `GET /api/v1/auth/me` (protected)

```go
    return router
}
```

**Return configured router**

#### main Function

```go
func main() {
```

**Application entry point:**
- Go calls this function when program starts

```go
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load application configuration: %v", err)
    }
```

**Load configuration:**
- Reads `.env` file
- Sets defaults
- `log.Fatalf`: Print error and exit if config fails

```go
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }
```

**Set Gin mode:**
- Development: Verbose logging, debug info
- Production: Less logging, faster

```go
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
```

**Create context with timeout:**
- `context.Background()`: Root context
- `WithTimeout(..., 10s)`: Auto-cancel after 10 seconds
- `defer cancel()`: Cleanup when main() exits

**Why timeout?**
- Database connection should be fast (< 10 seconds)
- If it takes longer, something's wrong
- Better to fail fast than hang forever

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

**Database connection:**

1. **Attempt connection:**
   ```go
   dbPool, err = database.NewPool(ctx, cfg)
   ```

2. **Connection failed:**
   ```go
   if err != nil {
       log.Printf("Warning: ...")
   }
   ```
   - Don't crash! Server can still handle health checks
   - "Degraded mode": Limited functionality

3. **Connection succeeded:**
   ```go
   defer dbPool.Close()
   ```
   - Close pool when program exits
   - Prevents connection leaks

**Degraded mode:**
```
Database down:
✓ GET /api/v1/health → Returns "database: DOWN"
✗ POST /api/v1/auth/login → Returns 500 error
```

```go
    var pinger handler.Pinger
    if dbPool != nil {
        pinger = dbPool
    }
```

**Handle nil pool:**
- If database failed, `dbPool` is `nil`
- `pinger` can be `nil` (health handler handles it)

```go
    jwtManager, err := token.NewJWTManager(cfg.JWTSecret, 24*time.Hour)
    if err != nil {
        log.Fatalf("Failed to initialize JWT manager: %v", err)
    }
```

**Create JWT manager:**
- Secret from config
- 24-hour token lifetime
- Fatal error if secret is too short (< 32 chars)

```go
    userRepo := repository.NewUserRepository(dbPool)
    authService := service.NewAuthService(userRepo, jwtManager)
    authHandler := handler.NewAuthHandler(authService)
    authMiddleware := middleware.Auth(jwtManager)
```

**Dependency injection chain:**

```
dbPool
  ↓
userRepo (needs dbPool)
  ↓
authService (needs userRepo + jwtManager)
  ↓
authHandler (needs authService)

jwtManager
  ↓
authMiddleware (needs jwtManager)
```

**This is dependency injection:**
- Each layer receives dependencies as parameters
- No global variables
- Easy to test (can inject mocks)

```go
    router := setupRouter(pinger, authHandler, authMiddleware)
```

**Create router with all dependencies**

```go
    serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
    srv := &http.Server{
        Addr:         serverAddr,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
```

**Configure HTTP server:**

| Setting | Value | Purpose |
|---------|-------|---------|
| `Addr` | `:8080` | Listen on port 8080 |
| `Handler` | router | Gin router handles requests |
| `ReadTimeout` | 15s | Max time to read request |
| `WriteTimeout` | 15s | Max time to write response |
| `IdleTimeout` | 60s | Max time for idle connections |

**Why timeouts?**

**Without timeouts:**
```
Slowloris attack: Open 1000 connections, send data veeeery slowly
Server: Waits forever for complete requests
Result: Server runs out of resources, crashes
```

**With timeouts:**
```
Slowloris: Opens connections, sends slow data
Server: "15 seconds passed, still no complete request? Closing connection."
Result: Attack fails, server stays up
```

```go
    go func() {
        log.Printf("FindIt API server listening on http://localhost:%s (%s mode)", cfg.ServerPort, cfg.Environment)
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            log.Fatalf("Server failed to listen: %v", err)
        }
    }()
```

**Start server in goroutine:**

1. **`go func() { ... }()`:**
   - Goroutine (concurrent function)
   - Runs in background
   - Main function continues immediately

2. **Why goroutine?**
   - `ListenAndServe()` blocks (never returns)
   - Without goroutine, code after this never runs
   - With goroutine, server runs while main() sets up shutdown handling

3. **Server startup:**
   ```go
   srv.ListenAndServe()
   ```
   - Opens port 8080
   - Starts accepting connections
   - Blocks until server stops

4. **Error handling:**
   ```go
   if err != nil && !errors.Is(err, http.ErrServerClosed) {
       log.Fatalf("Server failed to listen: %v", err)
   }
   ```
   - `http.ErrServerClosed`: Normal (graceful shutdown)
   - Other errors: Fatal (port in use, permission denied)

```go
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
```

**Graceful shutdown setup:**

1. **Create signal channel:**
   ```go
   quit := make(chan os.Signal, 1)
   ```
   - Channel receives OS signals
   - Buffer size 1 (can hold one signal)

2. **Register signal handlers:**
   ```go
   signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
   ```
   - `SIGINT`: Ctrl+C in terminal
   - `SIGTERM`: Kill command, Docker stop

3. **Wait for signal:**
   ```go
   <-quit
   ```
   - Blocks until signal received
   - When received, continues to shutdown code

**What's happening:**
```
Terminal: User presses Ctrl+C
OS: Sends SIGINT signal to process
Signal handler: Writes to quit channel
Main function: Unblocks from <-quit
Shutdown code: Runs
```

```go
    log.Println("Shutdown signal received, shutting down gracefully...")

    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer shutdownCancel()
```

**Shutdown preparation:**
- Log message
- Create context with 5-second timeout
- If shutdown takes > 5 seconds, force stop

```go
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
```

**Graceful shutdown:**

1. **`srv.Shutdown(shutdownCtx)`:**
   - Stop accepting new connections
   - Wait for active requests to finish
   - Close idle connections
   - Max 5 seconds (from context timeout)

2. **What "graceful" means:**
   ```
   ❌ Forceful shutdown:
   User mid-request → Server dies → User sees "connection reset"
   
   ✅ Graceful shutdown:
   User mid-request → Server finishes request → Returns response → Then stops
   ```

3. **Timeout:**
   - If requests still active after 5 seconds, force close
   - Better to drop a few requests than hang forever

```go
    log.Println("FindIt API server gracefully terminated")
}
```

**Shutdown complete**

### Summary: cmd/api/main.go

**Startup sequence:**
1. Load configuration (.env)
2. Connect to PostgreSQL
3. Create dependencies (repos, services, handlers)
4. Setup router and middleware
5. Start HTTP server
6. Wait for shutdown signal
7. Gracefully shutdown

**Key patterns:**
- **Dependency injection**: No global state
- **Graceful shutdown**: Finish active requests before stopping
- **Degraded mode**: Continue without database for health checks
- **Timeouts**: Protection against slow clients/attacks

---

## Testing Explained

### Why Test?

**Without tests:**
```go
func divide(a, b int) int {
    return a / b  // ← Bug: Division by zero crashes!
}

// Deploy to production
// User enters b=0
// 💥 Server crashes
```

**With tests:**
```go
func TestDivideByZero(t *testing.T) {
    result, err := divide(10, 0)
    if err == nil {
        t.Error("Expected error for division by zero")
    }
}

// Run: go test
// ❌ Test fails, bug caught before production
```

### Go Testing Basics

```go
// file: math.go
package math

func Add(a, b int) int {
    return a + b
}
```

```go
// file: math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
```

**Test conventions:**
- Test files end with `_test.go`
- Test functions start with `Test`
- Take `*testing.T` parameter
- `t.Error()` / `t.Fatal()` to fail tests

### Testing Our Code

#### Unit Test Example: `pkg/hasher/hasher_test.go`

```go
func TestHashPassword(t *testing.T) {
    password := "MySecretPassword123"
    
    hash, err := HashPassword(password)
    if err != nil {
        t.Fatalf("HashPassword failed: %v", err)
    }
    
    if !CheckPasswordHash(password, hash) {
        t.Error("Password hash verification failed")
    }
}
```

**What's tested:**
1. ✅ Hashing doesn't error
2. ✅ Hashed password can be verified

#### Integration Test Example: `internal/handler/auth_test.go`

```go
func TestAuthHandler_Register(t *testing.T) {
    // Setup: Create mock service
    mockService := &mockAuthService{
        registerFunc: func(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
            return &models.AuthResponse{
                Token: "fake-jwt-token",
                User: &models.UserResponse{
                    ID:    uuid.New(),
                    Name:  req.Name,
                    Email: req.Email,
                },
            }, nil
        },
    }
    
    // Setup: Create handler with mock
    handler := NewAuthHandler(mockService)
    router := gin.New()
    router.POST("/auth/register", handler.Register)
    
    // Execute: Send test request
    body := `{"name":"Alice","email":"alice@example.com","password":"password123"}`
    req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert: Check response
    if w.Code != http.StatusCreated {
        t.Errorf("Expected status 201, got %d", w.Code)
    }
}
```

**What's tested:**
1. ✅ Handler accepts valid JSON
2. ✅ Calls service correctly
3. ✅ Returns 201 status
4. ✅ Returns expected JSON format

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/handler/...

# Verbose output
go test ./... -v

# Run specific test
go test -run TestHashPassword ./pkg/hasher/...
```

---

## Conclusion

You now understand:

1. ✅ **Go basics**: Structs, interfaces, pointers, error handling
2. ✅ **PostgreSQL**: Tables, queries, constraints, indexes
3. ✅ **Clean architecture**: Handler → Service → Repository → Database
4. ✅ **Security**: Password hashing, JWT tokens, SQL injection prevention
5. ✅ **HTTP**: Status codes, middleware, request/response handling
6. ✅ **Testing**: Unit tests, integration tests, mocking

**Next steps:**
- Continue to Phase 3 (Items, Search, Claims)
- Practice modifying existing code
- Write your own tests
- Experiment with the API using curl/Postman

**Questions to solidify learning:**
1. Trace a request: Client sends `POST /auth/login` → Which functions are called in order?
2. Security: What happens if you remove `json:"-"` from `PasswordHash`?
3. Errors: What HTTP status should you return if a user tries to claim their own item?

Happy coding! 🚀
