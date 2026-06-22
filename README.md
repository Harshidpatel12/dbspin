# dbspin 🚀

[![Go Version](https://img.shields.io/github/go-mod/go-version/Harshidpatel12/dbspin)](https://github.com/Harshidpatel12/dbspin)
[![License](https://img.shields.io/github/license/Harshidpatel12/dbspin)](https://github.com/Harshidpatel12/dbspin/blob/main/LICENSE)
[![Build & Test](https://github.com/Harshidpatel12/dbspin/actions/workflows/test.yml/badge.svg)](https://github.com/Harshidpatel12/dbspin/actions)
[![Linter status](https://github.com/Harshidpatel12/dbspin/actions/workflows/lint.yml/badge.svg)](https://github.com/Harshidpatel12/dbspin/actions)

`dbspin` is a lightweight, zero-configuration CLI utility written in Go that lets you spin up fully pre-configured database instances instantly in Docker using a single command. 

No more writing `docker-compose.yml` configs or trying to remember complex Docker parameters for environment variables, ports, or volumes.

---

## ✨ Features

- **Instant Database Provisioning**: Start fully configured PostgreSQL, Redis, MongoDB, or MySQL instances in seconds.
- **Port Conflict Auto-Detection**: Automatically detects if the default database port is already in use and binds to the next available port.
- **Persistence by Default**: Automatically maps data volumes locally so database states persist between restarts.
- **Copy-Paste Connection Strings**: Instantly prints ready-to-use connection string URLs (e.g. `postgresql://...`) to drop into your `.env` configuration files.
- **Zero Config Files**: Works entirely out of the box using smart command-line defaults.

---

## 📦 Installation

Install easily using the one-line installer:

```bash
curl -fsSL https://raw.githubusercontent.com/Harshidpatel12/dbspin/main/install.sh | bash
```

Alternatively, you can compile and install from source if you have Go installed:

```bash
git clone https://github.com/Harshidpatel12/dbspin.git
cd dbspin
go build -o dbspin main.go
sudo mv dbspin /usr/local/bin/
```

---

## 🛠️ Usage

### 1. Spin up a Database
Specify the database engine you want to start:
```bash
dbspin up postgres
# or
dbspin up redis
```

Available engines:
- `postgres` (PostgreSQL)
- `redis` (Redis)
- `mysql` (MySQL)
- `mongo` (MongoDB)

### 2. List running instances
See all database containers managed by `dbspin`:
```bash
dbspin list
```

### 3. Stop an instance
Stop and remove a database container:
```bash
dbspin down postgres
```

### 4. View logs
Display standard output logs of the database engine container:
```bash
dbspin logs postgres
```

---

## ⚙️ Development

### Local Build and Testing
To run and test the program locally:
```bash
go run main.go --help
```

### Pre-commit hooks
We use pre-commit to run linting and code quality audits locally. Register the hooks with:
```bash
pre-commit install
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
