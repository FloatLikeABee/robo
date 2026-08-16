# DataPulse Deployment Guide

## Requirements

### Server Requirements
- **CPU**: 2+ cores
- **RAM**: 4GB+ (Metabase requires 2GB+)
- **Disk**: 10GB+ free space
- **OS**: Linux (Ubuntu 22.04+, Debian 11+ recommended)

### Software Requirements
- Docker 20.10+
- Docker Compose 1.29+
- Java 17+ (for Metabase)

## Deployment Options

### 1. Docker Compose (Recommended)

#### Quick Start

```bash
# Clone the repository
git clone https://github.com/your-repo/datapulse.git
cd datapulse

# Copy environment file
cp .env.example .env

# Edit .env with your settings
nano .env

# Start the application
cd deploy
docker-compose up -d
```

#### Configuration

Edit the `.env` file with your settings:

```env
# Server
PORT=3000
HOST=0.0.0.0

# Database
DATABASE_URL=postgresql://datapulse:secret@db:5432/datapulse

# JWT
JWT_SECRET=your-very-secure-secret-key

# Metabase
METABASE_DB_TYPE=postgres
METABASE_DB_URL=postgresql://metabase:metabase-secret@db:5432/metabase
```

#### Environment Variables

See `.env.example` for all available configuration options.

### 2. Manual Deployment

#### Backend Setup

```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env

# Build the backend
cd backend
cargo build --release

# Run the backend
./target/release/datapulse
```

#### Frontend Setup

```bash
# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt-get install -y nodejs

# Build the frontend
cd frontend
npm install
npm run build

# Serve the frontend (in production, use Nginx or similar)
npm run preview
```

### 3. Kubernetes (Advanced)

Kubernetes manifests are available in the `deploy/k8s` directory (not yet implemented).

## Production Configuration

### Database

For production, we recommend PostgreSQL:

```env
DATABASE_URL=postgresql://datapulse:securepassword@db:5432/datapulse
```

### Metabase Configuration

```env
METABASE_DB_TYPE=postgres
METABASE_DB_URL=postgresql://metabase:metabase-secret@db:5432/metabase
METABASE_JVM_OPTS=-Xmx4g -Xms2g
```

### Security

- Use HTTPS with a valid certificate
- Set secure JWT secret
- Configure CORS properly
- Use a reverse proxy (Nginx, Caddy)

## Reverse Proxy Configuration

### Nginx Example

```nginx
server {
    listen 80;
    server_name datapulse.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name datapulse.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /metabase/ {
        proxy_pass http://localhost:3001/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Monitoring

### Health Checks

```bash
# Check backend health
curl http://localhost:3000/api/v1/health

# Check Metabase health
curl http://localhost:3001/api/health
```

### Logs

```bash
# View Docker logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f app
```

## Backup and Restore

### Database Backup

```bash
# For SQLite
docker-compose exec app sqlite3 /app/data/datapulse.db .dump > backup.sql

# For PostgreSQL
docker-compose exec db pg_dump -U datapulse datapulse > backup.sql
```

### Database Restore

```bash
# For SQLite
docker-compose exec -T app sqlite3 /app/data/datapulse.db < backup.sql

# For PostgreSQL
docker-compose exec -T db psql -U datapulse datapulse < backup.sql
```

## Upgrading

### Docker Compose

```bash
# Pull latest images
cd deploy
docker-compose pull

# Recreate containers
docker-compose up -d --force-recreate

# Run migrations if needed
docker-compose exec app ./datapulse migrate
```

### Manual

```bash
# Stop the application
pkill datapulse

# Pull latest code
git pull

# Rebuild
cd backend
cargo build --release

# Restart
./target/release/datapulse
```

## Troubleshooting

### Common Issues

#### Metabase fails to start
- Check Java version: `java -version`
- Ensure port 3001 is available
- Check logs: `docker-compose logs metabase`

#### Database connection issues
- Verify database credentials
- Check network connectivity
- Test connection manually

#### Authentication problems
- Verify JWT secret matches
- Check token expiration
- Ensure CORS is configured correctly

### Debugging

```bash
# Increase logging level
RUST_LOG=debug docker-compose up

# Check environment variables
docker-compose exec app env

# Test database connection
docker-compose exec app sqlite3 /app/data/datapulse.db "SELECT * FROM users;"
```

## Scaling

### Horizontal Scaling

For horizontal scaling, you'll need to:
1. Use a shared database (PostgreSQL)
2. Configure Redis for session sharing
3. Use a load balancer
4. Ensure Metabase is configured for multiple instances

### Vertical Scaling

Increase resources for the Docker containers:

```yaml
# In docker-compose.yml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
```

## Performance Tuning

### Metabase Performance

```env
# Increase Metabase memory
METABASE_JVM_OPTS=-Xmx8g -Xms4g

# Enable query caching
MB_ENABLE_QUERY_CACHING=true
```

### Database Performance

```env
# For PostgreSQL
DATABASE_URL=postgresql://datapulse:password@db:5432/datapulse?pool_size=20
```

## Support

For deployment issues, contact support@datapulse.app