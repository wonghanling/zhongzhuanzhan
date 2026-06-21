# Azure VM deployment runbook

This runbook assumes:

- Azure VM uses Ubuntu 22.04/24.04
- New API and Nginx run on the VM with Docker Compose
- MySQL uses Azure Database for MySQL Flexible Server
- Redis uses Azure Cache for Redis

Goal:

- bring up the service
- create the first admin
- add one `chinallmapi.com`-style OpenAI-compatible upstream
- complete one real `/v1/chat/completions` call

## Startup modes

### A. External managed MySQL and Redis

Use:

```bash
docker compose -f docker-compose.azure.yml up -d
```

This mode expects `.env.azure` to point `SQL_DSN` and `REDIS_CONN_STRING` to Azure managed services or other external endpoints.

### B. Single-VM temporary full stack

Use:

```bash
docker compose -f docker-compose.local-stack.yml up -d
```

This mode starts:

- New API
- Nginx
- MariaDB
- Redis

It is useful for first acceptance on one Azure VM, then you can later migrate only the connection variables to Azure Database for MySQL and Azure Cache for Redis.

Detailed single-VM acceptance steps:

- [docs/local-stack-acceptance-checklist.md](D:\项目2\zhongzhuanzhan\docs\local-stack-acceptance-checklist.md)

## 1. Full commands to run on Azure VM

### Install Docker and Compose

```bash
sudo apt update
sudo apt install -y ca-certificates curl gnupg lsb-release git
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
newgrp docker
docker --version
docker compose version
```

### Prepare project directory

```bash
mkdir -p ~/apps
cd ~/apps
git clone https://github.com/QuantumNous/new-api.git boluo-api-gateway
cd boluo-api-gateway
git remote -v
mkdir -p deploy/nginx/certs
```

### Create environment file

Choose one:

```bash
cp .env.azure.template .env.azure
```

or

```bash
cp .env.azure.production.example .env.azure
```

### Edit environment file

```bash
nano .env.azure
```

### Start service

```bash
docker compose -f docker-compose.azure.yml pull
docker compose -f docker-compose.azure.yml up -d
docker compose -f docker-compose.azure.yml ps
```

### Start single-VM temporary full stack

```bash
docker compose -f docker-compose.local-stack.yml pull
docker compose -f docker-compose.local-stack.yml up -d
docker compose -f docker-compose.local-stack.yml ps
```

## 2. Ports to open

### Azure VM inbound ports

Open these in the Azure NSG:

- `22/tcp` for SSH
- `80/tcp` for HTTP
- `443/tcp` for HTTPS if you later mount certificates into Nginx

Optional:

- `3000/tcp` only for temporary debugging

### What does not need VM inbound exposure

Do not expose these on the VM:

- MySQL `3306`
- Redis `6379` or `6380`

Those are Azure managed services, not local containers in this architecture.

## 3. Key variables for `.env.azure`

Minimum required:

```dotenv
SQL_DSN=boluo_admin:REPLACE_ME@tcp(boluo-mysql.mysql.database.azure.com:3306)/new-api?charset=utf8mb4&parseTime=true&loc=Local&tls=true
REDIS_CONN_STRING=rediss://:REPLACE_ME@boluo-redis.redis.cache.windows.net:6380/0
SESSION_SECRET=replace-with-a-long-random-string

BOLUO_SYSTEM_NAME=Boluo API Gateway
BOLUO_LOGO_URL=
BOLUO_FOOTER_HTML=Boluo API Gateway
BOLUO_DOCS_LINK=https://chinallmapi.com

BOLUO_RETRY_TIMES=2
BOLUO_AUTO_DISABLE_STATUS_CODES=429,500-503
BOLUO_AUTO_RETRY_STATUS_CODES=429,500-503
```

Field meaning:

- `SQL_DSN`: Azure MySQL connection string
- `REDIS_CONN_STRING`: Azure Redis TLS URL
- `SESSION_SECRET`: login/session secret
- `BOLUO_*`: deployment preset injected at app startup

## 4. How to start `docker-compose.azure.yml`

### Start

```bash
docker compose -f docker-compose.azure.yml up -d
```

### Restart

```bash
docker compose -f docker-compose.azure.yml restart
```

### Stop

```bash
docker compose -f docker-compose.azure.yml down
```

### View running containers

```bash
docker compose -f docker-compose.azure.yml ps
```

## 5. How to check New API, MySQL, Redis, Nginx

### Check containers

```bash
docker compose -f docker-compose.azure.yml ps
docker logs --tail 100 boluo-new-api
docker logs --tail 100 boluo-nginx
```

For the single-VM full stack:

```bash
docker compose -f docker-compose.local-stack.yml ps
docker logs --tail 100 boluo-mysql
docker logs --tail 100 boluo-redis
docker logs --tail 100 boluo-new-api
docker logs --tail 100 boluo-nginx
```

### Check New API health endpoint

```bash
curl http://127.0.0.1:3000/api/status
curl http://127.0.0.1/api/status
curl http://YOUR_VM_IP/api/status
```

Expected result includes:

```json
{"success":true}
```

### Check Nginx is proxying

```bash
curl -I http://127.0.0.1/
curl -I http://YOUR_VM_IP/
```

### Check MySQL connectivity from app logs

```bash
docker logs boluo-new-api | grep -i mysql
docker logs boluo-new-api | grep -i "database migration started"
```

### Check Redis connectivity from app logs

```bash
docker logs boluo-new-api | grep -i redis
```

Expected log hints:

- `using MySQL as database`
- `Redis is enabled`

### Optional direct check from VM

Install client tools:

```bash
sudo apt install -y mysql-client redis-tools
```

Test MySQL:

```bash
mysql -h boluo-mysql.mysql.database.azure.com -u boluo_admin -p --ssl-mode=REQUIRED -D new-api -e "SELECT 1;"
```

Test Redis:

```bash
redis-cli -u "rediss://:YOUR_PASSWORD@boluo-redis.redis.cache.windows.net:6380/0" --tls PING
```

### Direct checks for single-VM temporary full stack

Test MariaDB:

```bash
docker exec -it boluo-mysql mariadb -uroot -p"$LOCAL_DB_ROOT_PASSWORD" -e "SHOW DATABASES;"
```

Test Redis:

```bash
docker exec -it boluo-redis redis-cli -a "$LOCAL_REDIS_PASSWORD" PING
```

## 6. How to create the first admin account

This version uses the setup API and setup page.

### Check setup state

```bash
curl http://YOUR_VM_IP/api/setup
```

If setup is not finished, create the first admin:

```bash
curl -X POST http://YOUR_VM_IP/api/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "ChangeThisPass123!",
    "confirmPassword": "ChangeThisPass123!",
    "SelfUseModeEnabled": false,
    "DemoSiteEnabled": false
  }'
```

Then sign in from browser:

```text
http://YOUR_VM_IP/sign-in
```

## 7. How to add a `chinallmapi.com`-style OpenAI-compatible upstream

Reference:

- [docs/channels/chinallmapi-setup.md](D:\项目2\zhongzhuanzhan\docs\channels\chinallmapi-setup.md)

### In admin UI

Go to:

```text
Admin -> Channels -> Add Channel
```

Fill:

```text
Name: chinallmapi-primary
Type: OpenAI
Base URL: <provider actual base url>
Key: <provider API key>
Models: GPT-5.5 Pro,Claude Creative,Gemini Vision
Group: default
Priority: 100
Weight: 100
Auto Ban: enabled
```

Model mapping example:

```json
{
  "GPT-5.5 Pro": "gpt-4o",
  "Claude Creative": "claude-3-5-sonnet",
  "Gemini Vision": "gemini-2.0-flash"
}
```

Important:

- the left side is your public model name
- the right side is the real upstream model id

## 8. How to test `/v1/chat/completions` with curl

### Step 1

Create a normal user in the UI, or use an existing user.

### Step 2

Create a token in:

```text
Console -> Token management
```

### Step 3

Run:

```bash
curl http://YOUR_VM_IP/v1/chat/completions \
  -H "Authorization: Bearer USER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "GPT-5.5 Pro",
    "messages": [
      {"role": "user", "content": "hello"}
    ]
  }'
```

If you place a domain on Nginx later, replace `http://YOUR_VM_IP` with your real domain.

## 9. If request fails, how to view logs

### App and Nginx logs

```bash
docker logs --tail 200 boluo-new-api
docker logs --tail 200 boluo-nginx
```

### Follow logs live

```bash
docker logs -f boluo-new-api
docker logs -f boluo-nginx
```

### Docker Compose combined logs

```bash
docker compose -f docker-compose.azure.yml logs -f
```

### Common failure directions

- `401/403`: upstream key invalid, or user token invalid
- `429`: upstream rate limited, should trigger retry if configured
- `500/502/503`: upstream error or proxy issue
- connection refused: service not up or Nginx not proxying
- MySQL TLS/auth error: DSN wrong
- Redis TLS/auth error: Redis URL wrong

### App-side record checks

After a successful or failed request, inspect in admin:

```text
Admin -> Logs / Usage logs / Channel logs
```

Check:

- which channel was used
- whether quota was deducted
- failure reason

## 10. How to back up MySQL data

Azure Database for MySQL already has managed automated backups, but you should also keep a manual logical dump process.

### Install client

```bash
sudo apt install -y mysql-client
```

### Manual dump

```bash
mkdir -p ~/mysql-backups
mysqldump \
  -h boluo-mysql.mysql.database.azure.com \
  -u boluo_admin \
  -p \
  --ssl-mode=REQUIRED \
  --single-transaction \
  --quick \
  --set-gtid-purged=OFF \
  new-api > ~/mysql-backups/new-api-$(date +%F-%H%M%S).sql
```

### Compress backup

```bash
gzip ~/mysql-backups/new-api-*.sql
ls -lh ~/mysql-backups
```

### Restore example

```bash
gunzip -c ~/mysql-backups/new-api-YYYY-MM-DD-HHMMSS.sql.gz | \
mysql -h boluo-mysql.mysql.database.azure.com -u boluo_admin -p --ssl-mode=REQUIRED new-api
```

## Final acceptance checklist

1. `docker compose ps` shows `boluo-new-api` and `boluo-nginx` as running
2. `curl http://YOUR_VM_IP/api/status` returns success
3. first admin account can sign in
4. `chinallmapi` channel can be added and tested
5. user token can be created
6. `/v1/chat/completions` returns a valid response
7. admin logs show channel usage and quota consumption
