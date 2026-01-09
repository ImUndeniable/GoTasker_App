# GoTasker – Task Management API (v1)

GoTasker is a production-ready task management application backend built using Golang, PostgreSQL, and JWT-based authentication.
It is designed to demonstrate real-world backend engineering practices, including authentication, authorization, database persistence, middleware, and clean project structure.

This project focuses on backend correctness, scalability, and clarity, with a minimal frontend used only for integration testing.

🚀 Features
✅ Core Features

User registration & login

Secure password hashing (bcrypt)

JWT-based authentication & authorization

CRUD operations for tasks

User-scoped data access (each user sees only their tasks)

PostgreSQL persistence

Clean, idiomatic Go project structure

Middleware-based request handling

🔐 Security

Passwords are never stored in plaintext

JWT tokens are signed and validated on every protected request

Unauthorized access is blocked at middleware level

No API-key fallback (JWT-only auth)

⚙️ Engineering Highlights

Stateless authentication using JWT

Context-based user identity propagation

Database-driven ID generation (no in-memory hacks)

Clean separation of concerns (handlers, middleware, models)

Ready for scaling and further extensions

🧱 Tech Stack
Backend

Go (Golang)

Chi Router

PostgreSQL

pgx (Postgres driver)

bcrypt (password hashing)

JWT (github.com/golang-jwt/jwt/v5)

Frontend (Minimal / Optional)

React

Vite

Used only to validate backend integration

Tooling

Docker (for PostgreSQL)

SQL migrations

Postman (API testing)

📁 Project Structure
gotasker/
├── cmd/
│   └── api/
│       └── main.go        # Application entry point
├── internal/
│   ├── auth/              # JWT logic & password utilities
│   ├── handlers/          # HTTP handlers (tasks, auth)
│   ├── middleware/        # Logging, JWT middleware
│   ├── models/            # Request/response & DB models
│   └── db/                # Database setup & helpers
├── migrations/            # SQL migrations
├── go.mod
└── go.sum

🧠 Authentication Flow (JWT)

User registers with email + password

Password is hashed using bcrypt

User logs in with credentials

Server verifies password hash

Server issues a JWT token

Client sends JWT in Authorization: Bearer <token>

JWT middleware:

Validates token

Extracts user_id

Injects it into request context

All task operations use user_id from context

🗃️ Database Schema
Users Table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

Tasks Table
CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

🔌 API Endpoints
Auth
Method	Endpoint	Description
POST	/register	Create a new user
POST	/login	Authenticate and receive JWT
Tasks (JWT Required)
Method	Endpoint	Description
GET	/tasksdb	List user tasks
POST	/tasksdb	Create a task
PATCH	/tasksdb/{id}	Update task
DELETE	/tasksdb/{id}	Delete task
🔑 Authorization Header

All protected endpoints require:

Authorization: Bearer <JWT_TOKEN>

🛠️ Setup & Installation
1️⃣ Clone Repository
git clone https://github.com/<your-username>/gotasker.git
cd gotasker

2️⃣ Start PostgreSQL (Docker)
docker run -d \
  --name gotasker-db \
  -e POSTGRES_USER=gotasker \
  -e POSTGRES_PASSWORD=gotasker \
  -e POSTGRES_DB=gotasker \
  -p 5432:5432 postgres:16

3️⃣ Run Migrations

Apply SQL files inside migrations/ in order.

4️⃣ Set Environment Variables

Windows (PowerShell)

setx JWT_SECRET "super-secret-jwt-key"


Linux / macOS

export JWT_SECRET="super-secret-jwt-key"

5️⃣ Run Server
go run cmd/api/main.go


Server runs at:

http://localhost:8080

🧪 Testing

Use Postman or curl

Login → copy JWT → attach to Authorization header

Test full CRUD lifecycle

📌 Status

✅ GoTasker v1 – COMPLETED

What’s Done

Auth system (Register + Login)

JWT security

User-based task isolation

Full CRUD

Clean backend architecture

Planned for v2 (Optional)

Refresh tokens

Role-based access

Pagination & filtering

Background workers

Deployment (Docker Compose / Cloud)

🧑‍💻 Why This Project Matters

GoTasker demonstrates:

Backend-first thinking

Security-aware API design

Production-grade Go patterns

Real-world authentication flows

This project was built with learning + correctness as the primary goals.

📜 License

MIT License
