# Authentication Service with Golang

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-DC382D?style=flat&logo=redis&logoColor=white)](https://redis.io/)

## Table of Contents

- [Introduction](#introduction)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Running the Service](#running-the-service)
- [API Endpoints](#api-endpoints)
    - [Health Check](#health-check)
    - [Authentication](#authentication)
    - [Protected Routes](#protected-routes)
    - [WebSocket & Real-Time Messaging](#websocket--real-time-messaging)
- [Project Structure](#project-structure)
- [Technologies Used](#technologies-used)
- [Database Schema](#database-schema)
- [Service Methods](#service-methods)
- [Configuration](#configuration)
- [Security Features](#security-features)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

## Introduction

This repository contains a production-ready authentication service built with Golang. The service demonstrates modern authentication practices using cookie-based sessions, JWT tokens, and PostgreSQL for user data persistence. It includes real-time messaging capabilities via Pusher WebSocket and comprehensive security features.

## Architecture

The architecture is composed of several core components:

- **Authentication Handler**: Manages user signup, signin, and logout operations
- **Session Management**: Cookie-based session storage with JWT tokens
- **User Management**: PostgreSQL database for persistent user data
- **Real-Time Messaging**: Pusher integration for WebSocket communication
- **Health Monitoring**: Service health check endpoints
- **Validation Layer**: Input validation and sanitization

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- [Golang](https://golang.org/doc/install) 1.19 or higher
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) (for generating type-safe SQL code)

### Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/KunalKumar-1/auth-golang-cookies.git
    cd auth-golang-cookies
    ```

2. Create an `.env` file in the root directory and set the necessary environment variables:

    ```sh
    cp .env.example .env
    ```

3. Example `.env` file:

    ```env
    # PostgreSQL Configuration
    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=postgres
    POSTGRES_HOST=localhost
    POSTGRES_PORT=5432
    
    # Database URL
    DB_URL=postgres://postgres:postgres@localhost:5432/golangcookies?sslmode=disable

    # Application Configuration
    PORT=4000
    
    # JWT Configuration
    JWT_SECRET=your_secure_jwt_secret_here
    
    # Pusher Configuration (for real-time messaging)
    PUSHER_APP_ID=your_pusher_app_id
    PUSHER_APP_KEY=your_pusher_app_key
    PUSHER_APP_SECRET=your_pusher_app_secret
    PUSHER_APP_CLUSTER=ap2
    ```

### Running the Service

1. Build and run the service using Docker Compose:

    ```sh
    docker-compose up --build
    ```

2. For local development without Docker:

    ```sh
    # Install dependencies
    go mod download
    
    # Run database migrations
    psql -U your_db_user -d auth_db -f sql/schema/0001_users.sql
    
    # Generate database code
    sqlc generate
    
    # Run the application
    go run cmd/main.go
    ```

3. The service will be available at:
    - API Server: `http://localhost:4000`

## API Endpoints

### Health Check

#### Check Service Health
```http
GET /health-check
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-23T10:30:00Z",
  "services": {
    "database": "connected",
    "redis": "connected"
  }
}
```

### Authentication

#### Sign Up (Register User)
```http
POST /signup
Content-Type: application/json

{
  "name": "Kunal Kumar",
  "username": "kunal",
  "email": "kk@gmail.com",
  "password": "securePassword123"
}
```

**Response:**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Kunal Kumar",
    "username": "kunal",
    "email": "kk@gmail.com"
  }
}
```

#### Sign In (Login)
```http
POST /sign-in
Content-Type: application/json

{
  "email": "kk@gmail.com",
  "password": "securePassword123"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "kunal",
    "email": "kk@gmail.com"
  }
}
```
*Sets session cookie in response headers*

#### Logout
```http
POST /logout
Content-Type: application/json
Cookie: session_token=<token>

{}
```

**Response:**
```json
{
  "message": "Logout successful"
}
```

### Protected Routes

#### Access Protected Route
```http
GET /auth-route
Accept: application/json
Cookie: session_token=<token>
```

**Response:**
```json
{
  "message": "You are authenticated!",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "kunal",
    "email": "kk@gmail.com"
  }
}
```

### WebSocket & Real-Time Messaging

#### Check WebSocket Connection
```http
GET /check-ws
Accept: application/json
```

**Response:**
```json
{
  "status": "WebSocket available",
  "endpoint": "ws://localhost:4000/ws",
  "pusher": {
    "cluster": "ap2",
    "connected": true
  }
}
```

#### Send Message (WebSocket/SSE)
```http
POST /send-message
Content-Type: application/json

{
  "message": "Hello world!",
  "username": "kunal"
}
```

**Response:**
```json
{
  "status": "Message sent successfully",
  "message": "Hello world!",
  "username": "kunal",
  "timestamp": "2025-11-23T10:30:00Z"
}
```

## Project Structure

```
auth-golang-cookies/
├── cmd/
│   └── main.go                    # Application entry point
├── handlers/
│   ├── handler_health.go          # Health check endpoints
│   ├── handler_auth.go            # Authentication handlers (signup, signin, logout)
│   ├── handler_pusher.go          # WebSocket/SSE handlers for real-time messaging
│   └── handler_user.go            # User management handlers
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management & environment variables
│   └── database/
│       ├── db.go                  # Database connection setup
│       ├── models.go              # Generated database models (sqlc)
│       └── user.sql.go            # Generated SQL queries (sqlc)
├── models/
│   └── user_model.go              # Application-level user models
├── sql/
│   ├── queries/
│   │   └── user.sql               # SQL queries for sqlc
│   └── schema/
│       └── 0001_users.sql         # Database schema migrations
├── utils/
│   └── validation.go              # Input validation utilities
├── tmp/                           # Temporary build files (gitignored)
│   ├── build-errors.log
│   └── main
├── docker-compose.yml             # Docker Compose configuration
├── Dockerfile                     # Docker build configuration
├── sqlc.yaml                      # sqlc configuration
├── go.mod                         # Go module dependencies
├── go.sum                         # Go module checksums
├── .env.example                   # Environment variables template
└── README.md                      # Project documentation
```

## Technologies Used

- **Golang**: Primary language for building the service
- **Docker**: Containerization of the application
- **Docker Compose**: Orchestrating multi-container Docker application
- **PostgreSQL**: Database for persistent user data storage
- **sqlc**: Generate type-safe Go code from SQL queries
- **bcrypt**: Password hashing library
- **JWT**: JSON Web Tokens for authentication
- **Pusher**: Real-time bidirectional communication and WebSocket management
- **HTTP Cookies**: Secure session management

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

### Running Migrations

```sh
# Apply migrations
psql -U your_db_user -d auth_db -f sql/schema/0001_users.sql

# Generate type-safe code from SQL queries
sqlc generate
```

## Service Methods

### Handler Methods

> ### Health Handler (`handler_health.go`)
```go
func (h *HealthHandler) CheckHealth(c *gin.Context)
```

> ### Authentication Handler (`handler_auth.go`)
```go
func (h *AuthHandler) SignUp(c *gin.Context)
func (h *AuthHandler) SignIn(c *gin.Context)
func (h *AuthHandler) Logout(c *gin.Context)
func (h *AuthHandler) ValidateSession(c *gin.Context)
```

> ### User Handler (`handler_user.go`)
```go
func (h *UserHandler) GetCurrentUser(c *gin.Context)
func (h *UserHandler) UpdateUserProfile(c *gin.Context)
func (h *UserHandler) DeleteUser(c *gin.Context)
```

> ### Pusher Handler (`handler_pusher.go`)
```go
func (h *PusherHandler) ConnectWebSocket(c *gin.Context)
func (h *PusherHandler) SendMessage(c *gin.Context)
func (h *PusherHandler) BroadcastMessage(message string)
func (h *PusherHandler) CheckWebSocketStatus(c *gin.Context)
```

### Database Methods (Generated by sqlc)

> ### User Queries (`user.sql.go`)
```go
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error)
func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error)
func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
func (q *Queries) DeleteUser(ctx context.Context, id uuid.UUID) error
```

### Utility Methods

> ### Validation (`validation.go`)
```go
func ValidateEmail(email string) error
func ValidatePassword(password string) error
func ValidateUsername(username string) error
func SanitizeInput(input string) string
```

### Redis Session Methods

```go
func (r *RedisClient) SetSession(sessionID string, userID uuid.UUID, expiry time.Duration) error
func (r *RedisClient) GetSession(sessionID string) (uuid.UUID, error)
func (r *RedisClient) DeleteSession(sessionID string) error
func (r *RedisClient) ExtendSession(sessionID string, expiry time.Duration) error
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `POSTGRES_USER` | PostgreSQL username | `postgres` | Yes |
| `POSTGRES_PASSWORD` | PostgreSQL password | `postgres` | Yes |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` | Yes |
| `POSTGRES_PORT` | PostgreSQL port | `5432` | Yes |
| `DB_URL` | Full database connection URL | - | Yes |
| `PORT` | Application port | `4000` | Yes |
| `JWT_SECRET` | Secret key for JWT token generation | - | Yes |
| `PUSHER_APP_ID` | Pusher application ID | - | Yes |
| `PUSHER_APP_KEY` | Pusher application key | - | Yes |
| `PUSHER_APP_SECRET` | Pusher secret key | - | Yes |
| `PUSHER_APP_CLUSTER` | Pusher cluster region | `ap2` | Yes |

## Security Features

- **Password Hashing**: Bcrypt algorithm with configurable cost
- **Secure Cookies**: HttpOnly, Secure, and SameSite flags enabled
- **Session Tokens**: Cryptographically secure random token generation
- **Input Validation**: Comprehensive validation for all user inputs
- **SQL Injection Prevention**: Type-safe queries using sqlc
- **CORS Protection**: Configurable CORS middleware
- **Session Expiration**: Automatic cleanup of expired sessions in Redis
- **Rate Limiting**: Protection against brute force attacks
- **XSS Protection**: Input sanitization and output encoding

## Development

### Live Reload with Air

Install Air for development with live reload:

```sh
go install github.com/air-verse/air@latest
```

Run with live reload:

```sh
air
```

### Code Generation

After modifying SQL queries in `sql/queries/user.sql`:

```sh
sqlc generate
```

### Project Commands

```sh
# Install dependencies
go mod download

# Run tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Build binary
go build -o auth-service cmd/main.go

# Run linter
golangci-lint run

# Format code
go fmt ./...
```

## Testing

### Running Tests

```sh
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific package tests
go test ./handlers/... -v

# Run with race detection
go test -race ./...
```

### Test Structure

```
auth-golang-cookies/
├── handlers/
│   ├── handler_auth_test.go
│   ├── handler_user_test.go
│   └── handler_pusher_test.go
├── utils/
│   └── validation_test.go
└── internal/
    └── database/
        └── db_test.go
```

## Deployment

### Docker Deployment

Build and deploy using Docker Compose:

```sh
# Build and start services
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Remove volumes (caution: deletes data)
docker-compose down -v
```

### Production Deployment

1. Build the production binary:

```sh
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service cmd/main.go
```

2. Set environment variables for production:

```env
ENVIRONMENT=production
SECRET_KEY=<generate-secure-random-key>
DB_URL=<production-database-url>
REDIS_HOST=<production-redis-host>
```

3. Run with systemd or process manager:

```sh
./auth-service
```

### Production Considerations

- Use HTTPS/TLS for all communications
- Enable secure cookie flags in production
- Set up database connection pooling
- Configure Redis persistence (RDB/AOF)
- Implement proper logging and monitoring
- Set up automated backups for PostgreSQL
- Use environment-specific configuration
- Enable rate limiting on all endpoints
- Set up log rotation
- Use a reverse proxy (nginx/Caddy)

## Roadmap

- [x] Cookie-based authentication
- [x] Session management with Redis
- [x] User signup and signin
- [x] WebSocket support for real-time messaging
- [x] Health check endpoints
- [ ] OAuth2 integration (Google, GitHub, Facebook)
- [ ] Two-factor authentication (2FA/TOTP)
- [ ] Password reset functionality
- [ ] Email verification system
- [ ] JWT token support (optional alternative)
- [ ] Role-based access control (RBAC)
- [ ] API rate limiting with Redis
- [ ] Comprehensive API documentation (Swagger/OpenAPI)
- [ ] Integration tests
- [ ] Performance benchmarks
- [ ] Admin dashboard
- [ ] User profile management
- [ ] Account deletion workflow
- [ ] Audit logging

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Coding Standards

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Write meaningful commit messages
- Add tests for new features
- Update documentation as needed
- Run `go fmt` and `golangci-lint` before committing

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE.md) file for details.
