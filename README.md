# dbspin 🚀

[![Go Version](https://img.shields.io/github/go-mod/go-version/Harshidpatel12/dbspin)](https://github.com/Harshidpatel12/dbspin)
[![License](https://img.shields.io/github/license/Harshidpatel12/dbspin)](https://github.com/Harshidpatel12/dbspin/blob/main/LICENSE)
[![Build & Test](https://github.com/Harshidpatel12/dbspin/actions/workflows/test.yml/badge.svg)](https://github.com/Harshidpatel12/dbspin/actions)
[![Linter status](https://github.com/Harshidpatel12/dbspin/actions/workflows/lint.yml/badge.svg)](https://github.com/Harshidpatel12/dbspin/actions)

`dbspin` is a lightweight, zero-configuration CLI utility written in Go that lets you spin up fully pre-configured database instances instantly in Docker using a single command. 

No more writing `docker-compose.yml` configs or trying to remember complex Docker parameters for environment variables, ports, or volumes.

---

## ✨ Features

- **Instant Database Provisioning**: Start fully configured PostgreSQL, Redis, MongoDB, MySQL, MariaDB, Elasticsearch, RabbitMQ, pgvector, TimescaleDB, Kafka, Meilisearch, LocalStack, MinIO, or DynamoDB instances in seconds.
- **Interactive Mode Setup**: Don't remember CLI flags? Run `dbspin` without arguments to step through a user-friendly selection wizard.
- **Companion Web GUI Dashboards**: Spin up companion admin dashboards (Adminer, Redis Commander, Mongo Express, Elasticvue) with a single flag.
- **Port Conflict Auto-Detection**: Automatically detects if the default database port is already in use and binds to the next available port.
- **Persistence by Default**: Automatically maps data volumes locally so database states persist between restarts.
- **Custom Naming & Multi-Instances**: Run multiple isolated containers of the same database by specifying custom name suffixes.
- **Environment Overrides**: Easily pass custom environment variables directly via the CLI.
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

#### Available Engines:
- `postgres` (PostgreSQL)
- `redis` (Redis)
- `mysql` (MySQL)
- `mariadb` (MariaDB)
- `mongo` (MongoDB)
- `elasticsearch` (Elasticsearch)
- `rabbitmq` (RabbitMQ)
- `pgvector` (PostgreSQL with Vector Search extension)
- `timescaledb` (PostgreSQL with Time-series extension)
- `kafka` (Apache Kafka Broker - KRaft mode)
- `meilisearch` (Meilisearch search engine)
- `localstack` (LocalStack AWS mock services)
- `minio` (MinIO Object Storage)
- `dynamodb` (DynamoDB Local emulator)

#### Command Flags for `up`:
- `-port int`: Override default host port
- `-version string`: Override default docker image tag/version
- `-name string`: Override container name and volume name suffix (allows spinning up multiple instances)
- `-env string`: Additional environment variables comma-separated (e.g. `KEY1=VAL1,KEY2=VAL2`)
- `-seed string`: Path to a local SQL or JS initialization file or directory of scripts to seed the database on initial start
- `-gui`: Start a companion web GUI dashboard for the database (e.g. Adminer, Redis Commander, Mongo Express, Elasticvue, Kafka UI, or DynamoDB Admin)
- `-wait`: Wait for the database engine to be fully initialized and ready to accept connections before returning

#### Examples:
```bash
# Spin up PostgreSQL with a companion Adminer web client
dbspin up postgres -gui

# Spin up Redis on a custom port 6380 with Redis Commander web UI
dbspin up redis -port 6380 -gui

# Spin up an isolated PostgreSQL instance for a specific app named 'auth'
dbspin up postgres -name auth

# Spin up MongoDB with custom credentials
dbspin up mongo -env "MONGO_INITDB_ROOT_USERNAME=custom,MONGO_INITDB_ROOT_PASSWORD=secret"

# Spin up PostgreSQL and seed it with a local SQL file
dbspin up postgres -seed ./init.sql

# Spin up PostgreSQL and seed it with all SQL files in a directory
dbspin up postgres -seed ./migrations/

# Spin up PostgreSQL and wait for it to be fully ready before returning
dbspin up postgres -wait
```

### 2. List running instances
See all database containers managed by `dbspin`:
```bash
dbspin list

# Verbose mode: includes connection strings and active companion GUI URLs
dbspin list -v
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

### 5. Interactive Shell
Instantly connect to your running database container via its native interactive CLI client (e.g. `psql`, `redis-cli`, `mysql`, `mongosh`):
```bash
dbspin shell postgres
```

### 6. Export Database Data
Dump database data directly to a local SQL/JSON/RDB file using shell redirection or the `-f` flag:
```bash
dbspin export postgres > backup.sql
# Or using direct file flag
dbspin export postgres -f backup.sql
```

### 7. Import Database Data
Restore database data from a local file directly into the container using shell redirection or the `-f` flag:
```bash
dbspin import postgres < backup.sql
# Or using direct file flag
dbspin import postgres -f backup.sql
```

### 8. View Connection Info
Print detailed container status, ports, connection URLs, default credentials, and companion GUI web dashboard URLs:
```bash
dbspin info postgres
```

### 9. Docker Compose Exporter
Generate a valid, production-ready `docker-compose.yml` configuration on-the-fly for any combination of databases:
```bash
# Generate compose file for PostgreSQL and Redis
dbspin compose postgres redis > docker-compose.yml

# Include companion web GUI dashboards (e.g. Adminer, Redis Commander) as services in the compose file
dbspin compose -gui postgres redis > docker-compose.yml

# Generate a compose file containing all 14 supported database engines
dbspin compose > docker-compose-all.yml
```

### 10. Prune Environment
Reset your local development environment by stopping and removing all dbspin containers and data volumes:
```bash
dbspin prune
```

### 11. Shell Auto-Completion
`dbspin` automatically configures shell auto-completion for your current shell (`bash`, `zsh`, or `fish`) upon its first execution.

If you ever need to manually generate or output the completion scripts, you can run:
```bash
# For Bash
dbspin completion bash

# For Zsh
dbspin completion zsh

# For Fish
dbspin completion fish
```

### 12. Interactive Setup Wizard
Simply run `dbspin` (or `dbspin up` with no engine) without arguments to open an interactive configuration wizard:
```bash
dbspin
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
