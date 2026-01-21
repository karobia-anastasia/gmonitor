# GitHub Repository Monitor

## Overview

This Go-based service fetches and monitors GitHub repository data, including commits and repository metadata, storing it
in a persistent database. The service ensures efficient querying, prevents duplicate commits, and allows resetting data
from a specified point in time.

## Features

- Fetches commit history and repository metadata from GitHub's API.
- Stores commit messages, authors, timestamps, and URLs in a database.
- Ensures commits in the database mirror GitHub's repository commits.
- Monitors repositories for new commits at a configurable interval.
- Allows resetting commit collection from a specific date.
- Supports efficient querying of commit data.
- Includes unit tests for core functionalities.

## Technologies Used

- **Golang**
- **GitHub REST API**
- **SQLite** (Configurable)
- **GORM** (ORM for database operations)

## Project Structure

```
├── .github
│   ├── workflows
│   │   └── go.yml        # github CI configurations
├── cmd
│   └── main.go          # Entry point of the service
├── config
│   └── config.go        # Configuration management
├── internal
│   ├── db
│   │   └── db.go        # Database connection setup
│   ├── fetcher
│   │   ├── client.go    # HTTP client for GitHub API
│   │   ├── fetcher.go   # Fetching logic for commits and repositories
│   │   └── types.go     # API response structures
│   ├── models
│   │   ├── commit.go    # Commit model definition
│   │   └── repository.go # Repository model definition
│   ├── monitor
│   │   ├── monitor.go   # Scheduler for monitoring GitHub repositories
│   │   └── worker.go    # Worker handling commit updates
│   ├── repository
│   │   ├── commit.go    # CRUD operations for commits
│   │   └── repository.go # CRUD operations for repositories
│   ├── server
│   │   ├── handlers.go    # CRUD endpoints for commits and repository
│   │   └── server.go      # Servemux and handlers configurations
├── pkg
│   ├── cache
│   │   └── cache.go        # In memory cache implementation
├── go.mod               # Go module file
├── go.sum               # Dependency lock file
├── readme.md            # Project documentation
├── Makefile             # Optional commands to run, test and build project 
```

## Setup Instructions

### Prerequisites

- Go 1.23+
- SQLite (configured via `config/config.go`)
- GitHub API Token (for authenticated requests)

### Installation

1. Clone the repository:
   ```shell
   git clone https://github.com/karobia-anastasia/gmonitor.git
   cd gmonitor
   ```
2. Install dependencies:
   ```shell
   go mod tidy
   ```
3. Create `.env` file in the root project and add required configurations as shown below
   ```shell
    # Database Configuration
    DB_DSN="gmonitor.db"
    SERVER_PORT=8000

    # GitHub API Configuration
    GITHUB_TOKEN="***************************************************"
   ```
4. Set up the database and Run the service:
   ```shell
   make run
   ```

## Setting Up a Repository to be Monitored

To set up a repository for monitoring, make a `POST` request to the following API endpoint:

```
POST http://localhost:8000/api/v1/repos
```

### Request Body:

The request body should include the repository details in the following format:

```json
{
  "repo": "chromium",
  "owner": "chromium",
  "date": "2025-01-01T00:00:00Z"
}
```

### Parameters:

- **`repo`** (required): The name of the repository to monitor (e.g., `chromium`).
- **`owner`** (required): The GitHub username of the repository owner (e.g., `chromium`).
- **`date`** (required): The date from which commit monitoring should start, specified in ISO 8601 format (e.g.,
  `2025-01-01T00:00:00Z`).

### Example Response:

```json
{
  "message": "Repository successfully added for monitoring."
}
```

## Getting Repository Details by Repository Name

To retrieve the details of a specific GitHub repository, make a `GET` request to the following endpoint:

```
GET http://localhost:8000/api/v1/repos?repo=chromium/chromium
```

### Query Parameters:

- **`repo`** (required): The repository name in the format `owner/repository`, e.g., `chromium/chromium`.

### Example Request:

```bash
GET http://localhost:8000/api/v1/repos?repo=chromium/chromium
```

### Example Response:

```json
{
  "repo": "chromium",
  "owner": "chromium",
  "commit_count": 1500,
  "last_commit": "2025-03-30T12:30:00Z"
}
```

## Querying Top Commit Authors

To retrieve the top commit authors for a specific repository, make a `GET` request to the following endpoint:

```
GET http://localhost:8000/api/v1/repos/commit-authors?limit=10&repo=chromium/chromium
```

### Query Parameters:

- **`repo`** (required): The repository name in the format `owner/repository`, e.g., `chromium/chromium`.
- **`limit`** (optional): The number of top authors to return (default: `10` if not provided).

### Example Request:

```bash
GET http://localhost:8000/api/v1/repos/commit-authors?limit=10&repo=chromium/chromium
```

### Example Response:

```json
{
  "authors": [
    {
      "name": "Author1",
      "commit_count": 500
    },
    {
      "name": "Author2",
      "commit_count": 450
    },
    {
      "name": "Author3",
      "commit_count": 350
    }
  ]
}
```

## Querying Repository Commits

To retrieve commits for a specific GitHub repository, make a `GET` request to the following endpoint:

```
GET http://localhost:8000/api/v1/repos/commits?repo=chromium/chromium&page=1&size=20
```

### Query Parameters:

- **`repo`** (required): The repository name in the format `owner/repository`, e.g., `chromium/chromium`.
- **`page`** (optional): The page number for pagination (default: `1` if not provided).
- **`size`** (optional): The number of commits per page (default: `20` if not provided).

### Example Request:

```bash
GET http://localhost:8000/api/v1/repos/commits?repo=chromium/chromium&page=1&size=20
```

### Example Response:

```json
{
  "commits": [
    {
      "commit_hash": "abcd1234",
      "message": "Initial commit",
      "author": "Author1",
      "timestamp": "2025-01-01T00:00:00Z"
    },
    {
      "commit_hash": "efgh5678",
      "message": "Added new feature",
      "author": "Author2",
      "timestamp": "2025-01-02T12:00:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 10,
    "total_commits": 200
  }
}
```

## Running Tests

```sh
make coverage
```

## Docker Setup

```sh
docker compose up
```
