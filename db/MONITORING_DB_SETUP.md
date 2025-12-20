# Monitoring Database Setup

The monitoring database is a **separate Turso database** dedicated to tracking and analytics, keeping the main ride event database focused on ride data.

## Database Separation

- **Main Database** (`TURSO_DB_URL`): Stores all ride data, groups, routes, etc.
- **Monitoring Database** (`TURSO_MONITORING_DB_URL`): Stores sync logs, analytics, and monitoring data

## Setup Instructions

### 1. Create the Monitoring Database in Turso

```bash
# Create a new Turso database for monitoring
turso db create cycle-scene-monitoring
```

### 2. Get Database Credentials

```bash
# Get the database URL
turso db show cycle-scene-monitoring

# Get the read-write token
turso db tokens create cycle-scene-monitoring --read-write
```

### 3. Update Environment Variables

Add to your `.env` file or deployment environment:

```env
# Monitoring Database
TURSO_MONITORING_DB_URL=libsql://cycle-scene-monitoring-xxxxx.turso.io
TURSO_MONITORING_DB_RW_TOKEN=eyJ...
```

### 4. Run Database Migrations

The monitoring database migrations are in `db/migrations/monitoring/`.

To initialize the monitoring database with the schema:

```bash
# If using a migration tool
turso db shell cycle-scene-monitoring < db/migrations/monitoring/1734782400_create_sync_logs_table.up.sql
```

Or execute the migration manually in the Turso console.

## Tables

### sync_logs

Tracks all background sync events from clients:

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER | Auto-incrementing primary key |
| client_id | TEXT | UUID for device/browser identification |
| sync_type | TEXT | Type of sync: "periodic", "manual", "foreground" |
| status | TEXT | Result: "success" or "error" |
| error_msg | TEXT | Error message if sync failed |
| ride_count | INTEGER | Number of rides synced |
| duration | INTEGER | Sync duration in milliseconds |
| timestamp | DATETIME | When the sync occurred |
| city_code | TEXT | City code (pdx, slc, etc.) |

Indexes:
- `idx_sync_logs_client_id`: For querying a specific client
- `idx_sync_logs_timestamp`: For time-based queries
- `idx_sync_logs_status`: For filtering by success/error
- `idx_sync_logs_city_code`: For city-specific analytics

## API Endpoints

All sync logging endpoints use the monitoring database:

- `POST /v1/sync-logs` - Log a sync event (public)
- `GET /v1/sync-logs` - Get recent sync logs (admin only)
- `GET /v1/sync-logs/stats` - Get sync statistics (admin only)
- `GET /v1/sync-logs/client/{clientId}` - Get logs for specific client

## Benefits

✅ **Isolation**: Monitoring data doesn't impact ride data performance
✅ **Focused**: Main DB stays clean with just ride/event data
✅ **Archivable**: Old sync logs can be archived/deleted without affecting ride data
✅ **Scalable**: Can grow monitoring data independently
✅ **Analytics**: Easier to query without join complexity

## Rollback

If you need to remove the monitoring database:

```bash
# Delete the Turso database
turso db destroy cycle-scene-monitoring
```

Remove environment variables:
- `TURSO_MONITORING_DB_URL`
- `TURSO_MONITORING_DB_RW_TOKEN`

The main database and app will continue to work without monitoring.
