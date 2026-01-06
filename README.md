# Alerts Platform V2

Alerts Platform V2 is a modern dashboard enabling comprehensive JIRA alert management, component health monitoring, and rule governance. The project adopts a decoupled architecture, featuring a high-performance Go-based backend and a modern React-based frontend.

## Features

- **Global Overview**: Real-time monitoring of alert distribution and trends across components, tenants, and clusters.
- **Component Details**: In-depth analysis of specific component alert history, statistics, and top tenant/cluster rankings.
- **Rule Management**: manage Prometheus and Logging alert rules, supporting Diff previews for rule changes and GitOps workflow integration.
- **JIRA Integration**: Automatically sync JIRA Issues, categorize Critical/Major incidents based on rules, supporting both full and incremental updates.
- **High Performance**: Built on Go + Gin framework, supporting Air hot-reload for efficient development.
- **Modern UI**: Built with React + TypeScript + Vite, offering a responsive layout and smooth interactive experience.

## Tech Stack

### Backend
- **Language**: Go
- **Framework**: [Gin](https://gin-gonic.com/)
- **Database**: SQLite (via GORM)
- **Tools**: Air (Hot Reload), Go-JIRA

### Frontend
- **Framework**: React, TypeScript
- **Build Tool**: Vite
- **Styling**: TailwindCSS / CSS Modules
- **Charts**: Recharts

## Architecture

```text
+-------------------------------------------------------------+
|                     Frontend (React + Vite)                 |
|                                                             |
|   +---------------------+       +-----------------------+   |
|   |    Dashboard UI     |       |    Rule Manager UI    |   |
|   +----------+----------+       +-----------+-----------+   |
|              |                              |               |
+--------------|------------------------------|---------------+
               | HTTP / JSON                  |
               v                              v
+-------------------------------------------------------------+
|                      Backend (Go + Gin)                     |
|                                                             |
|   +-----------------------------------------------------+   |
|   |                      REST API                       |   |
|   +--------------------------+--------------------------+   |
|                              |                              |
|                 +------------v-------------+                |
|                 |       Core Services      |                |
|                 +-------+----------+-------+                |
|                         |          |                        |
|   +-------------+       |          |      +-------------+   |
|   | Async Tasks |<------+          +----->| JIRA Client |   |
|   +------+------+                         +------+------+   |
|          |                                       |          |
+----------|---------------------------------------|----------+
           |                                       |
           v                                       v
+-----------------------+               +---------------------+
|          DB           |               |    Atlassian JIRA   |
+-----------------------+               +---------------------+
```

## Quick Start

### 1. Prerequisites

Ensure your development environment has the following installed:
- Go (1.20+)
- Node.js (18+)
- Make

### 2. Configuration

The backend service relies on a `.env` file for configuration (e.g., JIRA credentials).

```bash
cd backend
cp .env.example .env
# Edit .env and fill in JIRA_SERVER, JIRA_USER, JIRA_TOKEN, etc.
```

#### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `JIRA_SERVER` | Yes | JIRA server URL (e.g., `https://tidb.atlassian.net`) |
| `JIRA_USER` | Yes | JIRA user email |
| `JIRA_TOKEN` | Yes | JIRA API token |
| `PORT` | No | Server port (default: `8818`) |
| `HOST` | No | Server host (default: empty) |
| `TIDB_DSN` | No | TiDB connection string for Name Service (cluster/tenant name lookup) |

#### TiDB Name Service (Optional)

The Name Service provides cluster and tenant name resolution by querying TiDB directly. To enable it:

```bash
# Format: user:password@tcp(host:port)/database?tls=tidb
export TIDB_DSN="admin:password@tcp(gateway01.us-east-1.prod.aws.tidbcloud.com:4000)/mydb?tls=tidb"
```

If `TIDB_DSN` is not configured, the service will start normally but name lookup functionality will be unavailable.

### 3. Running Locally

#### Backend

We recommend using the `dev.sh` script to start the backend, which automatically detects and installs `air` for hot reloading:

```bash
cd backend
./dev.sh
```
Alternatively, run manually:
```bash
cd backend
go run cmd/server/main.go
```

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend application is typically accessible at `http://localhost:5173`, and the backend API defaults to port `:8080`.

### 4. Build for Production

A `Makefile` is provided in the root directory to build production artifacts:

```bash
# Build both backend and frontend, and generate a release package
make release
```

Artifacts will be placed in the `generated/` directory.

## Directory Structure

```
.
├── backend/                # Go Backend Project
│   ├── cmd/server/         # Service Entry Point
│   ├── internal/           # Core Logic (API, Models, Services)
│   ├── config/             # Configuration
│   ├── DEV_GUIDE.md        # Backend Developer Guide
│   └── JIRA_SYNC_GUIDE.md  # JIRA Sync Mechansim Guide
├── frontend/               # React Frontend Project
│   ├── src/                # Source Cloud
│   └── dist/               # Build Artifacts
├── scripts/                # Build & Release Scripts
└── Makefile                # Project Management Commands
```

## Documentation

- [Frontend Developer Guide](frontend/README.md) (TBD)

## 📝 License

MIT
