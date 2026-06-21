# Azure preflight checklist

This checklist is for the exact deployment flow:

```bash
cp .env.azure.production.example .env.azure
nano .env.azure
docker compose -f docker-compose.azure.yml up -d
```

Alternative temporary full-stack flow:

```bash
cp .env.azure.production.example .env.azure
nano .env.azure
docker compose -f docker-compose.local-stack.yml up -d
```

## 1. DNS should point to the Azure VM

If you use a domain such as `api.yourdomain.com`:

- create an `A` record
- point it to the Azure VM public IPv4 address

Example:

```text
Host: api
Type: A
Value: <AZURE_VM_PUBLIC_IP>
TTL: 300
```

If you use the root domain:

```text
Host: @
Type: A
Value: <AZURE_VM_PUBLIC_IP>
TTL: 300
```

Check after DNS update:

```bash
nslookup api.yourdomain.com
ping api.yourdomain.com
```

## 2. Azure firewall / NSG ports

Open inbound:

- `22/tcp` for SSH
- `80/tcp` for HTTP
- `443/tcp` for HTTPS

Optional only during debugging:

- `3000/tcp`

Do not expose on the VM:

- `3306`
- `6379`
- `6380`

## 3. Variables that must be checked before Docker Compose start

Mandatory:

- `SQL_DSN`
- `REDIS_CONN_STRING`
- `SESSION_SECRET`

Mandatory for local full stack:

- `LOCAL_DB_PASSWORD`
- `LOCAL_DB_ROOT_PASSWORD`
- `LOCAL_SQL_DSN`
- `LOCAL_REDIS_PASSWORD`
- `LOCAL_REDIS_CONN_STRING`

Recommended:

- `BOLUO_SYSTEM_NAME`
- `BOLUO_DOCS_LINK`
- `BOLUO_RETRY_TIMES`
- `BOLUO_AUTO_DISABLE_STATUS_CODES`
- `BOLUO_AUTO_RETRY_STATUS_CODES`

Sanity checks:

- `SQL_DSN` includes `parseTime=true`
- MySQL database name exists, or the account can create/migrate tables
- `REDIS_CONN_STRING` matches actual Redis password and port
- Azure Redis should normally use `rediss://` on `6380`
- `SESSION_SECRET` is not a weak short string

## 4. If MySQL / Redis run as local containers first

Use these values in `.env.azure`:

### MySQL container

```dotenv
SQL_DSN=root:<CHANGE_ME>@tcp(mysql:3306)/new-api?charset=utf8mb4&parseTime=true&loc=Local
```

### Redis container

```dotenv
REDIS_CONN_STRING=redis://:<CHANGE_ME>@redis:6379/0
```

Notes:

- host is the Docker service name, not `localhost`
- no `tls=true` for a plain local MySQL container
- no `rediss://` for a plain local Redis container

## 5. If later switching to Azure Database for MySQL / Azure Cache for Redis

Only replace connection variables.

### Move from local MySQL container to Azure MySQL

From:

```dotenv
SQL_DSN=root:<CHANGE_ME>@tcp(mysql:3306)/new-api?charset=utf8mb4&parseTime=true&loc=Local
```

To:

```dotenv
SQL_DSN=boluo_admin@boluo-mysql:<CHANGE_ME>@tcp(boluo-mysql.mysql.database.azure.com:3306)/new-api?charset=utf8mb4&parseTime=true&loc=Local&tls=true
```

### Move from local Redis container to Azure Redis

From:

```dotenv
REDIS_CONN_STRING=redis://:<CHANGE_ME>@redis:6379/0
```

To:

```dotenv
REDIS_CONN_STRING=rediss://:<CHANGE_ME>@boluo-redis.redis.cache.windows.net:6380/0
```

Restart after changes:

```bash
docker compose -f docker-compose.azure.yml down
docker compose -f docker-compose.azure.yml up -d
```

### If using the single-VM temporary full stack now

Keep:

```dotenv
LOCAL_SQL_DSN=boluo:<CHANGE_ME>@tcp(mysql:3306)/new-api?charset=utf8mb4&parseTime=true&loc=Local
LOCAL_REDIS_CONN_STRING=redis://:<CHANGE_ME>@redis:6379/0
```

Later, when moving to Azure managed services:

- stop using `docker-compose.local-stack.yml`
- switch to `docker-compose.azure.yml`
- replace only `SQL_DSN` and `REDIS_CONN_STRING`
- keep the rest of the application preset variables unchanged

## 6. Final pre-start commands

```bash
test -f .env.azure
grep -E '^(SQL_DSN|REDIS_CONN_STRING|SESSION_SECRET)=' .env.azure
docker --version
docker compose version
docker compose -f docker-compose.azure.yml config > /tmp/boluo-compose-rendered.yml
tail -n 50 /tmp/boluo-compose-rendered.yml
docker compose -f docker-compose.azure.yml up -d
docker compose -f docker-compose.azure.yml ps
```
