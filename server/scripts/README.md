# TatuScan Server Utility Scripts

This directory contains utility scripts for database management and maintenance of the TatuScan inventory system.

## Available Scripts

### 1. `add_manual_inventory.py`
**Purpose**: Manually add inventory entries via API

**Description**:
Sends manual inventory records to the TatuScan API. Useful for adding machines that don't have the client installed or for testing purposes.

**Usage**:
```bash
# Add single entry
./add_manual_inventory.py --hostname "server-01" --os "linux" --os-version "Ubuntu 22.04"

# Add entry with custom machine_id
./add_manual_inventory.py --hostname "server-01" --os "linux" \
  --os-version "Ubuntu 22.04" --machine-id "custom-id-123"

# Use custom API endpoint
export TATUSCAN_URL=http://tatuscan.example.com:8040
./add_manual_inventory.py --hostname "server-01" --os "linux" --os-version "Ubuntu 22.04"
```

**Options**:
- `--hostname`: Machine hostname (required)
- `--os`: Operating system (required)
- `--os-version`: OS version (required)
- `--machine-id`: Custom machine ID (optional, auto-generated if not provided)
- `--ip`: IP address (default: 0.0.0.0)
- `--cpu-percent`: CPU usage percentage (default: 0.0)
- `--memory-total-mb`: Total memory in MB (default: 0)
- `--memory-used-mb`: Used memory in MB (default: 0)
- `--api-base`: API base URL (default: http://localhost:8040/api)

### 2. `delete_older.py`
**Purpose**: Remove duplicate inventory entries keeping the most recent

**Description**:
Identifies and removes duplicate inventory records based on hostname, keeping only the most recently updated entry. Useful for cleaning up after migrations or bulk imports.

**Usage**:
```bash
# Dry run (show what would be deleted)
./delete_older.py --dry-run

# Actually delete duplicates
./delete_older.py

# Use custom API endpoint
export TATUSCAN_URL=http://tatuscan.example.com:8040
./delete_older.py
```

**Options**:
- `--dry-run`: Show what would be deleted without actually deleting
- `--api-base`: API base URL (default: http://localhost:8040/api)

**Output**:
- Shows statistics of duplicates found
- Lists which records will be kept vs deleted
- Displays deletion results

### 3. `update_activation.py`
**Purpose**: Update computer activation dates from CSV file

**Description**:
Updates the `computer_activation` field in inventory records by importing data from a CSV file. Also calculates `activation_days` based on the activation date.

**Usage**:
```bash
# Use default CSV file (inventario.csv)
./update_activation.py

# Use custom CSV file
./update_activation.py --csv-path /path/to/custom.csv

# Dry run to preview changes
./update_activation.py --dry-run

# Use custom API endpoint
export TATUSCAN_URL=http://tatuscan.example.com:8040
./update_activation.py --csv-path activation_data.csv
```

**CSV Format**:
The CSV file should contain at least these columns:
- Hostname or machine identifier
- Activation date (various formats supported: DD/MM/YYYY, YYYY-MM-DD, etc.)

Example CSV:
```csv
hostname,activation_date,model
server-01,15/01/2024,Dell PowerEdge R740
server-02,2024-02-20,HP ProLiant DL380
```

**Options**:
- `--csv-path`: Path to CSV file (default: inventario.csv)
- `--dry-run`: Preview changes without updating database
- `--api-base`: API base URL (default: http://localhost:8040/api)

### 4. `convert_db.py`
**Purpose**: Convert legacy database to new TatuScan format

**Description**:
Converts data from legacy database format to the current TatuScan schema. Handles timezone normalization (America/Cuiaba) and schema updates.

**Fixed Paths** (edit script to change):
- **Source**: `/tmp/tatuscan_legacy.db` (legacy SQLite database)
- **Destination**: `/tmp/tatuscan_new.db` (new format database)

**Usage**:
```bash
# Ensure PYTHONPATH includes src directory
export PYTHONPATH="$PWD/src"
python scripts/convert_db.py
```

**Features**:
- Drops and recreates destination schema
- Normalizes all dates to America/Cuiaba timezone
- Migrates all inventory records
- Preserves data integrity

**Warning**: This script will **overwrite** the destination database. Make backups before running.

## Environment Variables

All scripts support the following environment variable:

- `TATUSCAN_URL`: Base URL of the TatuScan server (e.g., http://localhost:8040)
  - Scripts automatically append `/api` to this URL
  - Overrides the `--api-base` command-line option

## Prerequisites

All scripts require:
- Python 3.8+
- TatuScan server running and accessible
- Dependencies from `server/requirements.txt` installed

For database scripts (`convert_db.py`):
- SQLAlchemy
- Flask
- pytz

## Common Workflows

### Cleaning Up After Migration
```bash
# 1. Check for duplicates
./delete_older.py --dry-run

# 2. Remove duplicates
./delete_older.py

# 3. Verify results
curl http://localhost:8040/api/machines | jq '.count'
```

### Importing Activation Data
```bash
# 1. Prepare CSV file with hostname and activation dates
# 2. Preview changes
./update_activation.py --csv-path activation_data.csv --dry-run

# 3. Apply changes
./update_activation.py --csv-path activation_data.csv

# 4. Verify in dashboard
# Visit http://localhost:8040/dashboard
```

### Adding Test Data
```bash
# Add multiple test machines
for i in {1..5}; do
  ./add_manual_inventory.py \
    --hostname "test-server-$(printf '%02d' $i)" \
    --os "linux" \
    --os-version "Ubuntu 22.04" \
    --ip "192.168.1.$((100+i))" \
    --cpu-percent "$((RANDOM % 100))" \
    --memory-total-mb 8192 \
    --memory-used-mb "$((RANDOM % 8192))"
done
```

## Troubleshooting

### Connection Refused
```bash
# Check if server is running
curl http://localhost:8040/api/health

# Check server status
cd server && docker compose ps
# or
make server-ps
```

### API Authentication Errors
- Ensure the API endpoint is correct
- Check server logs for authentication requirements
- Verify `.env` configuration

### Database Lock Errors (convert_db.py)
```bash
# Stop the server before running conversion
make server-stop

# Run conversion
export PYTHONPATH="$PWD/src"
python scripts/convert_db.py

# Restart server
make server-start
```

### CSV Import Failures
- Verify CSV encoding is UTF-8
- Check date formats are consistent
- Ensure hostname column matches database records
- Use `--dry-run` to preview without making changes

## Support

For issues with these scripts:
1. Check script output and error messages
2. Verify server is running and accessible
3. Check server logs: `make server-logs`
4. Consult main project documentation
