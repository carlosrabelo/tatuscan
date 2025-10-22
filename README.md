# TatuScan - Distributed Machine Inventory System

TatuScan is a distributed machine inventory system with a lightweight Go client and a Python Flask server for collecting and monitoring machine information across networks.

## Overview

TatuScan consists of two main components:
- **Client**: Lightweight Go agent that collects system information (Windows/Linux)
- **Server**: Flask-based API server with SQLite/PostgreSQL/MySQL support

The system collects machine data every 60 seconds and provides a web dashboard for visualization and reporting.

## Architecture

```
┌─────────────────┐    HTTP POST JSON    ┌─────────────────┐
│   Go Client     │ ───────────────────→ │  Flask Server   │
│  (Windows/Linux)│    /api/machines     │   + Database    │
└─────────────────┘                      └─────────────────┘
         │                                       │
         │ 60s interval                          │
         │                              Web Dashboard
         └───────────────────────────────────────┘
```

## Features

### Client (Go)
- **Cross-platform**: Windows 7-11, Linux (modern distributions)
- **Lightweight**: Minimal resource footprint
- **Service integration**: systemd (Linux) and Windows Services support
- **Secure identification**: Machine ID based on physical MAC addresses (SHA-256)
- **Configurable interval**: Adjustable collection frequency
- **Robust filtering**: Excludes virtual/cloud interfaces automatically

### Server (Python Flask)
- **RESTful API**: Clean JSON endpoints for data ingestion
- **Multiple databases**: SQLite (default), PostgreSQL, MySQL support
- **Web dashboard**: HTML interface for inventory visualization and reporting
- **Timezone support**: Configurable timezone (default: America/Cuiaba)
- **Docker ready**: Containerized deployment support
- **Production ready**: Supervisord and Nginx integration

## Quick Start

### Prerequisites
- **Go 1.23+** (for client)
- **Python 3.10+** (for server)
- **Docker** (optional, for containerized deployment)

### Installation

1. **Clone the repository**:
   ```bash
   git clone <REPOSITORY_URL>
   cd tatuscan
   ```

2. **Build the client**:
   ```bash
   make client-build
   ```

3. **Setup the server**:
   ```bash
   cd server
   python3 -m venv .venv
   source .venv/bin/activate
   pip install -r requirements.txt
   cd ..
   ```

4. **Start the server**:
   ```bash
   make server-start
   ```

5. **Configure and run the client**:
   ```bash
   export TATUSCAN_URL=http://localhost:8040
   ./client/tatuscan
   ```

## Project Structure

```
tatuscan/
├── bin/                   # Compiled binaries
│   ├── linux/            # Linux executables
│   └── windows/          # Windows executables
├── client/                # Go client application
│   ├── cmd/tatuscan/     # Main entry point
│   ├── internal/         # Internal packages
│   ├── tools/            # Utility scripts
│   ├── tatuscan.wxs      # Windows installer configuration
│   ├── .env.example      # Environment variables template
│   └── go.mod/go.sum     # Go dependencies
├── server/               # Python Flask server
│   ├── tatuscan/         # Main application package
│   │   ├── blueprint/    # Flask blueprints (api, home, report, charts)
│   │   ├── config/       # Configuration module
│   │   ├── errors/       # Error handlers
│   │   ├── logging/      # Logging configuration
│   │   ├── models/       # Database models
│   │   ├── services/     # Business logic layer
│   │   ├── utils/        # Shared utilities
│   │   └── templates/    # HTML templates
│   ├── scripts/          # Database utility scripts
│   ├── requirements.txt  # Python dependencies
│   ├── run.py            # Development entry point
│   ├── Dockerfile        # Container configuration
│   ├── docker-compose.yml # Local development compose
│   └── .env.example      # Environment variables template
├── deploy/               # Deployment configurations
│   ├── docker/           # Production Docker setup
│   ├── k8s/              # Kubernetes manifests
│   ├── systemd/          # Linux systemd services
│   └── README.md         # Deployment guide
├── scripts/              # Build and deployment scripts
│   ├── client-build.sh   # Client build script
│   ├── server-*.sh       # Server management scripts
│   └── clean*.sh         # Cleanup scripts
├── Makefile             # Main project Makefile
├── README.md            # This file (English)
└── README-PT.md         # Portuguese documentation
```

## Configuration

### Client Configuration

Create `.env` file in `client/` directory:

```bash
# Server URL (mandatory)
TATUSCAN_URL=http://localhost:8040

# Collection interval (optional, default: 60s)
TATUSCAN_INTERVAL=60s

# Log level (optional, default: warn)
TATUSCAN_LOG_LEVEL=warn
```

### Server Configuration

Create `.env` file in `server/` directory:

```bash
# Server port (default: 8040)
TATUSCAN_PORT=8040

# Database configuration
# SQLite (default)
TATUSCAN_DB_DIR=/data
TATUSCAN_DB_FILE=tatuscan.db

# PostgreSQL (optional)
# SQLALCHEMY_DATABASE_URI=postgresql+psycopg2://user:pass@host:5432/dbname

# MySQL (optional)
# SQLALCHEMY_DATABASE_URI=mysql+pymysql://user:pass@host:3306/dbname

# Timezone
TIMEZONE=America/Cuiaba

# Flask secret key
SECRET_KEY=change-me-in-production
```

### Environment Variables

Main variables read by the application:

- `TATUSCAN_PORT`: Internal app port. Falls back to `PORT` if not defined. Default: `8040`
- `PORT`: Fallback for port when `TATUSCAN_PORT` is not defined
- `SQLALCHEMY_DATABASE_URI`: Complete database URI (takes priority if defined)
- `TATUSCAN_DB_DIR`: SQLite directory when `SQLALCHEMY_DATABASE_URI` is not defined. Default: `/data`
- `TATUSCAN_DB_FILE`: SQLite filename. Default: `tatuscan.db`
- `TIMEZONE`: Timezone for display. Default: `America/Cuiaba`
- `SECRET_KEY`: Flask secret key. Default: `dev` (change in production)

## Usage

### Building the Project

```bash
# Build client for current platform
make client-build

# Build client for all platforms
make client-build-all

# Build server Docker image
make server-build
```

### Running the Applications

#### Development

**Server:**
```bash
cd server
source .venv/bin/activate
./run.py
```

Access:
- Dashboard: `http://localhost:8040/`
- Report: `http://localhost:8040/report`
- Charts: `http://localhost:8040/charts`
- API endpoint: `http://localhost:8040/api/machines`

**Client:**
```bash
# Run client (single collection)
make client-run

# Run client as daemon
make client-daemon
```

#### Docker Deployment

Use Makefile targets to manage the container:

```bash
make server-start   # Start with docker compose up -d
make server-stop    # Stop with docker compose down
make server-restart # Restart only the app service
make server-logs    # Follow logs
make server-ps      # Container status
```

**Notes:**
- The `docker-compose.yml` automatically reads `.env`
- Port mapping: `${HOST_PORT:-8040}:${TATUSCAN_PORT:-8040}`

**Overriding host port (HOST_PORT):**

Option 1 — via `.env` (recommended):
```env
# .env
HOST_PORT=18040       # Port exposed on host
TATUSCAN_PORT=8040    # Internal app port (don't change normally)
```

Then:
```bash
make server-start
# Access at http://localhost:18040
```

Option 2 — ad-hoc via shell:
```bash
HOST_PORT=18040 make server-start
```

### Production Deployment

1. **Server Setup**:
   ```bash
   # Create directories
   sudo mkdir -p /opt/tatuscan
   sudo mkdir -p /var/lib/tatuscan

   # Copy files to production directory
   sudo cp -r server/* /opt/tatuscan/

   # Setup permissions
   sudo chown -R www-data:www-data /opt/tatuscan
   sudo chown -R www-data:www-data /var/lib/tatuscan
   sudo chmod 750 /opt/tatuscan
   sudo chmod 750 /var/lib/tatuscan

   # Setup supervisord
   sudo cp deploy/tatuscan.conf /etc/supervisor/conf.d/
   sudo supervisorctl reread
   sudo supervisorctl update
   sudo supervisorctl start tatuscan
   ```

2. **Client Installation**:
   ```bash
   # Install as service (Linux)
   sudo make client-install

   # Install as service (Windows)
   ./bin/tatuscan-windows-amd64.exe install
   ```

3. **Nginx Configuration**:
   ```bash
   # Copy Nginx config
   sudo cp deploy/nginx/tatuscan.conf /etc/nginx/sites-available/
   sudo ln -s /etc/nginx/sites-available/tatuscan.conf /etc/nginx/sites-enabled/
   sudo nginx -t
   sudo systemctl reload nginx
   ```

### Development Commands

```bash
# Format Go code
make client-fmt

# Run Go tests
make client-test

# Run Python tests
make server-test

# Lint code
make lint

# Clean build artifacts
make clean
```

## API Endpoints

### GET /api/machines
List all registered machines.

**Response:**
```json
{
  "items": [
    {
      "machine_id": "sha256-hash",
      "hostname": "server-01",
      "ip": "192.168.1.100",
      "os": "linux",
      "os_version": "Ubuntu 22.04",
      "cpu_percent": 15.5,
      "memory_total_mb": 8192,
      "memory_used_mb": 4096,
      "created_at": "2025-01-01T12:00:00-04:00",
      "updated_at": "2025-01-01T12:30:00-04:00"
    }
  ],
  "count": 1
}
```

### POST /api/machines
Receive data from TatuScan clients.

**Request Body:**
```json
{
  "machine_id": "sha256-hash",
  "hostname": "server-01",
  "ip": "192.168.1.100",
  "os": "linux",
  "os_version": "Ubuntu 22.04",
  "cpu_percent": 15.5,
  "memory_total_mb": 8192,
  "memory_used_mb": 4096,
  "timestamp": "2025-01-01T12:30:00-04:00"
}
```

### GET /api/health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy"
}
```

## Data Collected

The client collects the following information:

| Field | Type | Description |
|-------|------|-------------|
| `machine_id` | string | SHA-256 hash of physical MAC addresses |
| `hostname` | string | Machine hostname |
| `ip` | string | Primary IPv4 address |
| `os` | string | Operating system (linux/windows/darwin) |
| `os_version` | string | Human-readable OS version |
| `cpu_percent` | float | CPU usage percentage |
| `memory_total_mb` | integer | Total memory in MB |
| `memory_used_mb` | integer | Used memory in MB |
| `timestamp` | string | ISO 8601 timestamp |

## Database Structure

The `Inventory` table contains:

- `machine_id`: Primary key (SHA-256 hash, unique per machine)
- `hostname`: Host name
- `ip`: IP address
- `os`: Operating system
- `os_version`: Operating system version
- `cpu_percent`: CPU usage percentage
- `memory_total_mb`: Total memory in MB
- `memory_used_mb`: Used memory in MB
- `created_at`: Creation date (UTC)
- `updated_at`: Last update date (configured timezone)

To inspect the database:

```bash
# Development (when using local SQLite)
sqlite3 /path/to/tatuscan/tatuscan.db

# Production
sqlite3 /var/lib/tatuscan/tatuscan.db

# Query
SELECT * FROM inventory;
```

## Using External Database (Postgres/MySQL)

To use PostgreSQL or MySQL instead of SQLite:

1. Define `SQLALCHEMY_DATABASE_URI` in `.env`:

```env
# PostgreSQL (driver psycopg2)
SQLALCHEMY_DATABASE_URI=postgresql+psycopg2://user:pass@host:5432/dbname

# MySQL/MariaDB (driver PyMySQL)
SQLALCHEMY_DATABASE_URI=mysql+pymysql://user:pass@host:3306/dbname
```

2. Drivers already included in `requirements.txt`:
   - PostgreSQL: `psycopg2-binary`
   - MySQL/MariaDB: `PyMySQL`

If you're not using an external database, you can remove them from `requirements.txt` for a leaner image.

### About the `/data` volume in Docker

When `SQLALCHEMY_DATABASE_URI` is defined for Postgres/MySQL, the volume mapped to `/data` is not used, but it doesn't interfere either. If you want to remove it:

**Option 1** — Edit `docker-compose.yml` and comment out the volume line:
```yaml
services:
  tatuscan:
    # ...
    # volumes:
    #   - tatuscan_data:/data
```

**Option 2** — Use an override file to not map `/data`:

Create `docker-compose.no-sqlite.yml`:
```yaml
services:
  tatuscan:
    volumes: []
```

And start with:
```bash
docker compose -f docker-compose.yml -f docker-compose.no-sqlite.yml up -d
```

## Testing with TatuScan Client

1. **Build the client**:
   ```bash
   cd client
   go build -o tatuscan cmd/tatuscan/main.go
   ```

2. **Configure environment variable**:
   ```bash
   # Development
   export TATUSCAN_URL=http://localhost:8040

   # Production
   export TATUSCAN_URL=http://tatuscan.example.com
   ```

3. **Run the client**:
   ```bash
   ./tatuscan
   ```

4. **Verify data in the report**:
   ```bash
   curl http://localhost:8040/report
   ```

   Or access via browser.

## Security Considerations

- **Machine ID**: Based on physical MAC addresses only (excludes virtual interfaces)
- **HTTPS**: Use HTTPS in production environments
- **Authentication**: Consider adding API authentication for production
- **Firewall**: Configure appropriate firewall rules
- **Secrets**: Never commit `.env` files or secrets to version control
- **Database**: Use strong passwords and restrict access to database servers
- **File permissions**: Set appropriate permissions on configuration files and databases

## Troubleshooting

### Client Issues

**Problem**: "No valid physical network interface found"
- **Solution**: Check if machine has physical network interfaces connected

**Problem**: "Environment variable TATUSCAN_URL not defined"
- **Solution**: Set the TATUSCAN_URL environment variable

**Problem**: Connection refused
- **Solution**: Verify server is running and URL is correct

### Server Issues

**Problem**: Database connection errors
- **Solution**: Check database configuration and permissions

**Problem**: Port already in use
- **Solution**: Change TATUSCAN_PORT or stop conflicting services

**Problem**: Permission denied on `/var/lib/tatuscan`
- **Solution**: Check ownership and permissions (www-data:www-data, 750)

**Problem**: Supervisord not starting
- **Solution**: Check logs in `/var/log/tatuscan.err.log` and `/var/log/tatuscan.out.log`

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue in the repository
- Contact: contato@carlosrabelo.com.br
