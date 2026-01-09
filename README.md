# GoTasker 🚀

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)
![License](https://img.shields.io/badge/License-MIT-green)

**GoTasker** is a production-ready task management API built to demonstrate robust backend engineering practices in Golang. It features secure JWT authentication, strict clean architecture, and user-scoped data persistence using PostgreSQL.

This project focuses on **backend correctness, scalability, and clarity**, moving beyond simple tutorials to implement real-world patterns like middleware-based context propagation and concurrent request handling.

---

## 🌟 Key Features

### ✅ Core Functionality
* **User Management:** Secure registration and login flows.
* **Task CRUD:** Create, Read, Update, and Delete tasks with persistence.
* **Data Isolation:** Strict user-scoped access (users can only manage their own tasks).
* **Robust Persistence:** Leverages PostgreSQL for reliable data storage (no in-memory shortcuts).

### 🔐 Security First
* **JWT Authentication:** Stateless auth using `golang-jwt/v5`.
* **Password Security:** Industry-standard **bcrypt** hashing; passwords are never stored in plaintext.
* **Middleware Protection:** Protected routes verify signatures and inject User IDs into the request context.

### ⚙️ Engineering Highlights
* **Clean Architecture:** Clear separation of concerns (Handlers ↔ Services ↔ Repositories).
* **Context Propagation:** Uses Go's `context` package to manage request lifecycles and user identity.
* **Database-Driven IDs:** Utilizes PostgreSQL `BIGSERIAL` for scalable ID generation.
* **Containerized:** Includes Docker instructions for rapid database setup.

---

## 🧱 Tech Stack

| Category | Technology | Usage |
| :--- | :--- | :--- |
| **Language** | **Golang** | Core backend logic |
| **Router** | **Chi** | Lightweight, idiomatic HTTP routing |
| **Database** | **PostgreSQL** | Relational data persistence |
| **Driver** | **pgx** | High-performance Postgres driver for Go |
| **Auth** | **JWT (v5)** | Stateless token-based authentication |
| **Security** | **bcrypt** | Password hashing |
| **DevOps** | **Docker** | Containerization for database services |

---

## 📂 Project Structure

GoTasker follows a modular standard project layout to ensure maintainability:

```bash
gotasker/
├── cmd/
│   └── api/
│       └── main.go        # Application entry point
├── internal/
│   ├── auth/              # JWT logic & password hashing utilities
│   ├── handlers/          # HTTP handlers (Controller layer)
│   ├── middleware/        # Auth middleware & request logging
│   ├── models/            # Data structures & Database models
│   └── db/                # Database connection & helpers
├── migrations/            # SQL migration files
├── go.mod                 # Go module definition
└── go.sum
```

## 🧠 Authentication Flow

**Sign Up:** User sends Email + Password → Password hashed via bcrypt → Saved to DB.
**Login:** User sends credentials → Hash compared → Server signs a JWT.
**Access:** Client sends Authorization: Bearer <token> in headers.
**Validation:** Middleware validates the token, extracts user_id, and injects it into the Request Context.
**Execution:** Handlers retrieve user_id from context to execute isolated queries.

## 🗃️ Database Schema
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

## 🔌 API Endpoints
**All Tasks endpoints require the Authorization header.**
Method,Endpoint,Description,Auth
POST,/register,Register a new user,❌
POST,/login,Authenticate and receive JWT,❌
GET,/tasksdb,Get all tasks for logged-in user,✅
POST,/tasksdb,Create a new task,✅
PATCH,/tasksdb/{id},Update task status/title,✅
DELETE,/tasksdb/{id},Delete a specific task,✅

## 🛠️ Setup & Installation
**1. Clone the Repository**
git clone [https://github.com/ImUndeniable/gotasker.git]
cd gotasker

**2. Start PostgreSQL (via Docker)**
docker run -d \
  --name gotasker-db \
  -e POSTGRES_USER=gotasker \
  -e POSTGRES_PASSWORD=gotasker \
  -e POSTGRES_DB=gotasker \
  -p 5432:5432 postgres:16

**3. Run Migrations**
Apply the SQL files found in the /migrations folder to your database using your preferred SQL client (e.g., DBeaver, pgAdmin) or CLI.

**4. Set Environment Variables**
You must set the JWT_SECRET before running the app.

**Linux / macOS:**
Bash
export JWT_SECRET="super-secret-jwt-key"

**Windows (PowerShell):**
PowerShell
$env:JWT_SECRET="super-secret-jwt-key"

**5. Run the Server**
**Bash**
go run cmd/api/main.go

**Server runs at http://localhost:8080**
## 📌 Roadmap & Future Improvements (v2)

    [ ] Refresh Tokens: Implement sliding sessions for better UX.

    [ ] RBAC: Add Role-Based Access Control (Admin vs User).

    [ ] Pagination: Add cursor-based pagination for task lists.

    [ ] Background Workers: Move email notifications to a worker queue.

## 📜 License

Distributed under the MIT License. See LICENSE for more information.
