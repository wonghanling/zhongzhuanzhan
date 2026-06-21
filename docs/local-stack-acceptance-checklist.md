# Local stack acceptance checklist

This checklist is specifically for:

```bash
docker compose -f docker-compose.local-stack.yml up -d
```

It assumes one Azure VM runs the full temporary stack:

- New API
- Nginx
- MariaDB
- Redis

Goal:

- finish first-stage real acceptance on one VM
- verify admin login, upstream channel creation, user key creation, and one real API call

## 1. VM initial environment checks

### Check Ubuntu version

```bash
cat /etc/os-release
uname -a
```

Expected:

- Ubuntu 22.04 or 24.04 is preferred

### Check Docker

```bash
docker --version
docker info
```

Expected:

- Docker command exists
- Docker daemon is running

### Check Docker Compose

```bash
docker compose version
```

Expected:

- `docker compose` works without error

### Check current directory

```bash
pwd
ls
```

Expected:

- you are inside the project directory
- `docker-compose.local-stack.yml` exists

### Check `.env.azure`

```bash
test -f .env.azure && echo ".env.azure exists" || echo ".env.azure missing"
grep -E '^(SESSION_SECRET|LOCAL_DB_PASSWORD|LOCAL_DB_ROOT_PASSWORD|LOCAL_REDIS_PASSWORD)=' .env.azure
```

Expected:

- `.env.azure` exists
- required variables are present

## 2. Start full stack

```bash
docker compose -f docker-compose.local-stack.yml up -d
docker compose -f docker-compose.local-stack.yml ps
```

Expected containers:

- `boluo-mysql`
- `boluo-redis`
- `boluo-new-api`
- `boluo-nginx`

## 3. View logs

```bash
docker compose -f docker-compose.local-stack.yml logs -f new-api
docker compose -f docker-compose.local-stack.yml logs -f nginx
docker compose -f docker-compose.local-stack.yml logs -f mysql
docker compose -f docker-compose.local-stack.yml logs -f redis
```

Useful combined log:

```bash
docker compose -f docker-compose.local-stack.yml logs -f
```

## 4. Check container health status

Use:

```bash
docker compose -f docker-compose.local-stack.yml ps
```

Expected status by service:

- `New API`: `Up ... (healthy)`
- `MariaDB`: `Up ... (healthy)`
- `Redis`: `Up ... (healthy)`
- `Nginx`: `Up`

Notes:

- `nginx` currently has no explicit healthcheck in Compose, so `Up` is expected
- `new-api` should only start after `mysql` and `redis` are healthy

If you want exact health JSON:

```bash
docker inspect --format='{{json .State.Health}}' boluo-mysql
docker inspect --format='{{json .State.Health}}' boluo-redis
docker inspect --format='{{json .State.Health}}' boluo-new-api
```

## 5. Check ports

Run:

```bash
ss -tulpn | grep -E '80|443|3000|3306|6379'
```

Interpretation:

- `80` should be exposed by Nginx
- `3000` should be exposed by New API
- `443` may be published by Docker, but it will not be useful until you add TLS certs and Nginx HTTPS config
- `3306` and `6379` are internal container ports in this Compose file and are not published to the host by default

So the important host ports for first acceptance are:

- `80`
- `3000`

## 6. Browser access

Open in browser:

```text
http://服务器公网IP
http://绑定域名
```

For direct API status check:

```text
http://服务器公网IP/api/status
http://绑定域名/api/status
```

Expected:

- home page or setup/login page can load
- `/api/status` returns JSON with `"success": true`

## 7. First administrator account creation

### Actual behavior in this version

This version should not be treated as "auto-create default root user" during normal first boot acceptance.

Use the setup flow instead.

### Recommended method: setup API

Check whether setup is pending:

```bash
curl http://服务器公网IP/api/setup
```

If setup is not finished, create the first admin:

```bash
curl -X POST http://服务器公网IP/api/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "ChangeThisPass123!",
    "confirmPassword": "ChangeThisPass123!",
    "SelfUseModeEnabled": false,
    "DemoSiteEnabled": false
  }'
```

Then log in from browser.

### Fallback method: promote an existing user in MariaDB

If you already created a normal user and need to promote it:

```bash
docker exec -it boluo-mysql mariadb -uroot -p"$LOCAL_DB_ROOT_PASSWORD" -e "USE new-api; SELECT id, username, role, status FROM users;"
```

Promote that user to root admin:

```bash
docker exec -it boluo-mysql mariadb -uroot -p"$LOCAL_DB_ROOT_PASSWORD" -e "USE new-api; UPDATE users SET role=100, status=1, quota=100000000 WHERE username='admin';"
```

Verify:

```bash
docker exec -it boluo-mysql mariadb -uroot -p"$LOCAL_DB_ROOT_PASSWORD" -e "USE new-api; SELECT id, username, role, status, quota FROM users WHERE username='admin';"
```

## 8. Add the first third-party aggregator channel

Use an OpenAI-compatible third-party aggregator in the admin UI.

Go to:

```text
Admin -> Channels -> Add Channel
```

Fill these fields:

- `渠道名称 / Name`: `chinallmapi-primary`
- `类型 / Type`: `OpenAI`
- `Base URL`: `https://<UPSTREAM_HOST>/v1`
- `API Key`: `<UPSTREAM_API_KEY>`
- `模型名 / Models`: `GPT-5.5 Pro`
- `分组 / Group`: `default`
- `状态 / Status`: `enabled`
- `优先级 / Priority`: `100`
- `权重 / Weight`: `100`
- `Auto Ban`: `enabled`

### About ratio / multiplier

The model charge multiplier is not primarily a channel field.

Configure it in the backend model/ratio settings after the channel is created:

```text
Admin -> System Settings -> Models / Ratio / Pricing
```

Example idea:

- display model: `GPT-5.5 Pro`
- charge ratio: `2`

### Model mapping example

In channel model mapping JSON:

```json
{
  "GPT-5.5 Pro": "gpt-4o"
}
```

Use the real upstream model name on the right side.

## 9. Create user API key

Recommended flow:

1. Create or log in as a normal user
2. Open:

```text
Console -> Token management
```

3. Create a new token
4. Copy the generated token value immediately

This token is the `Authorization: Bearer ...` value used for OpenAI-compatible requests.

## 10. curl test for `/v1/chat/completions`

```bash
curl http://服务器公网IP/v1/chat/completions \
  -H "Authorization: Bearer <USER_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "<MODEL_NAME>",
    "messages": [
      {"role": "user", "content": "hello"}
    ]
  }'
```

Example:

```bash
curl http://服务器公网IP/v1/chat/completions \
  -H "Authorization: Bearer sk-user-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "GPT-5.5 Pro",
    "messages": [
      {"role": "user", "content": "hello"}
    ]
  }'
```

Expected:

- JSON response from the upstream model
- no authentication or model-mapping error

## 11. Acceptance success criteria

All of these should be true:

- page can open in browser
- administrator can log in
- channel can be saved successfully
- user API key can be created
- `curl` returns a model reply
- backend shows the request in logs
- quota deduction / usage record is visible

## 12. Common failure troubleshooting

### Nginx 502

Check:

```bash
docker compose -f docker-compose.local-stack.yml logs -f nginx
docker compose -f docker-compose.local-stack.yml logs -f new-api
```

Usually means:

- `new-api` is not healthy
- upstream container name or port is wrong

### New API cannot connect to MySQL

Check:

```bash
docker compose -f docker-compose.local-stack.yml logs -f new-api
docker compose -f docker-compose.local-stack.yml logs -f mysql
```

Common causes:

- `LOCAL_SQL_DSN` password does not match `LOCAL_DB_PASSWORD`
- database name in DSN is wrong
- MySQL container has not become healthy yet

### New API cannot connect to Redis

Check:

```bash
docker compose -f docker-compose.local-stack.yml logs -f new-api
docker compose -f docker-compose.local-stack.yml logs -f redis
```

Common causes:

- `LOCAL_REDIS_CONN_STRING` password does not match `LOCAL_REDIS_PASSWORD`
- Redis URL host is not `redis`

### Upstream API key is wrong

Symptoms:

- upstream returns `401` or `403`
- channel test fails

Check the channel key and provider dashboard.

### Base URL is wrong

Symptoms:

- request fails immediately
- upstream returns 404 or connection errors

Check whether you used the correct provider API root, usually something like:

```text
https://provider.example.com/v1
```

### Model name does not match

Symptoms:

- `model not found`
- request routed but upstream rejects model id

Check:

- left side of `model_mapping` is your public model name
- right side is the real upstream model id

### Port not open

Check:

```bash
ss -tulpn | grep -E '80|3000'
```

Also verify Azure NSG allows inbound `80/tcp`.

### DNS not effective

Check:

```bash
nslookup 你的域名
ping 你的域名
```

If DNS still points elsewhere, browser access will fail even if containers are healthy.

### Container exits immediately after start

Check:

```bash
docker compose -f docker-compose.local-stack.yml ps
docker compose -f docker-compose.local-stack.yml logs --tail 200
```

Most common reasons:

- missing required environment variables
- wrong DB or Redis password
- invalid DSN / Redis URL

## 13. Stop and restart commands

```bash
docker compose -f docker-compose.local-stack.yml restart
docker compose -f docker-compose.local-stack.yml down
docker compose -f docker-compose.local-stack.yml up -d
```

## 14. Backup and restore MariaDB data

### Backup

```bash
mkdir -p ~/db-backups
docker exec boluo-mysql sh -c 'mariadb-dump -uroot -p"$LOCAL_DB_ROOT_PASSWORD" --single-transaction --quick new-api' > ~/db-backups/new-api-$(date +%F-%H%M%S).sql
gzip ~/db-backups/new-api-*.sql
ls -lh ~/db-backups
```

### Restore

```bash
gunzip -c ~/db-backups/new-api-YYYY-MM-DD-HHMMSS.sql.gz | docker exec -i boluo-mysql sh -c 'mariadb -uroot -p"$LOCAL_DB_ROOT_PASSWORD" new-api'
```

### Volume-level persistence note

Even without a manual dump, MariaDB data persists across container recreation because this stack uses:

- `local_mysql_data`

Redis persistence uses:

- `local_redis_data`

But manual SQL dumps are still recommended before major upgrades or migration.
