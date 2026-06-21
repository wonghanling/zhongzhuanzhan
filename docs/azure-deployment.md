# Azure deployment for Boluo API Gateway

This project can run on Azure with a simple split:

- Azure VM: `New API` + `Nginx`
- Azure Database for MySQL: application database
- Azure Cache for Redis: cache, rate limit, session, sync support

Detailed step-by-step VM commands and acceptance checks:

- [docs/azure-vm-runbook.md](D:\项目2\zhongzhuanzhan\docs\azure-vm-runbook.md)

## 1. What New API already supports

The current codebase already supports the core pieces needed for phase 1:

- MySQL via `SQL_DSN`
- Redis via `REDIS_CONN_STRING`
- OpenAI-compatible upstream channels with custom `base_url`
- channel `priority`
- channel `weight`
- channel `model_mapping`
- automatic retry controlled by `RetryTimes`
- auto-disable channel logic for retryable upstream failures
- user quota, token management, logs, and billing ratios

Relevant code:

- [model/main.go](D:\项目2\zhongzhuanzhan\model\main.go)
- [common/redis.go](D:\项目2\zhongzhuanzhan\common\redis.go)
- [model/channel.go](D:\项目2\zhongzhuanzhan\model\channel.go)
- [model/ability.go](D:\项目2\zhongzhuanzhan\model\ability.go)
- [service/channel_select.go](D:\项目2\zhongzhuanzhan\service\channel_select.go)
- [controller/relay.go](D:\项目2\zhongzhuanzhan\controller\relay.go)

## 2. Recommended first deployment

Use these files:

- [docker-compose.azure.yml](D:\项目2\zhongzhuanzhan\docker-compose.azure.yml)
- [.env.azure.template](D:\项目2\zhongzhuanzhan\.env.azure.template)
- [.env.azure.production.example](D:\项目2\zhongzhuanzhan\.env.azure.production.example)
- [deploy/nginx/default.conf](D:\项目2\zhongzhuanzhan\deploy\nginx\default.conf)
- [docs/azure-preflight-checklist.md](D:\项目2\zhongzhuanzhan\docs\azure-preflight-checklist.md)

Suggested VM layout:

1. Install Docker Engine and Docker Compose plugin on the Azure VM.
2. Copy this repo to the VM.
3. Copy `.env.azure.template` or `.env.azure.production.example` to `.env.azure` and fill real values.
4. Run:

```bash
docker compose -f docker-compose.azure.yml up -d
```

5. Open `http://<vm-ip>/api/status` and `http://<vm-ip>/`.

This repo also includes a small deployment preset layer. On startup it can
persist these values into New API options automatically:

- `BOLUO_SYSTEM_NAME`
- `BOLUO_LOGO_URL`
- `BOLUO_FOOTER_HTML`
- `BOLUO_DOCS_LINK`
- `BOLUO_RETRY_TIMES`
- `BOLUO_AUTO_DISABLE_STATUS_CODES`
- `BOLUO_AUTO_RETRY_STATUS_CODES`

## 3. MySQL notes

New API directly supports MySQL DSN strings. The code auto-adds `parseTime=true` if missing.

Expected DSN shape:

```text
user:password@tcp(host:3306)/dbname?charset=utf8mb4&parseTime=true&loc=Local
```

For Azure Database for MySQL Flexible Server:

- host is usually `xxx.mysql.database.azure.com`
- username is often `user@server-name`
- keep `charset=utf8mb4`
- keep `parseTime=true`
- if TLS is enforced, use `tls=true` first

Practical example:

```text
boluo_admin:password@tcp(boluo-mysql.mysql.database.azure.com:3306)/new-api?charset=utf8mb4&parseTime=true&loc=Local&tls=true
```

## 4. Redis notes

New API uses `redis.ParseURL`, so standard Redis URLs work.

Recommended Azure Redis URL:

```text
rediss://:password@boluo-redis.redis.cache.windows.net:6380/0
```

Notes:

- prefer `rediss://` because Azure Redis commonly requires TLS
- database index `/0` is fine for phase 1
- if Redis is not set, New API still runs, but you lose important cache and sync behavior

## 5. Phase 1 admin setup

After first login:

1. Sign in with root account if auto-created, or complete setup wizard.
2. Verify deployment preset values:
   - `SystemName = Boluo API Gateway`
   - `Logo = your logo URL`
   - `Footer = your footer HTML/text`
   - `DocsLink = your external docs URL`
3. Create at least three third-party aggregator channels:
   - `n1n.ai`
   - `OpenRouter`
   - `allapi` or another OpenAI-compatible aggregator
4. For each channel:
   - use `OpenAI` compatible channel type
   - set custom `base_url`
   - set upstream API key
   - set `priority`
   - set `weight`
   - enable `auto ban`
   - configure `model_mapping`
5. Create user account and user token.
6. Call `/v1/chat/completions` with the user token.

## 6. How to implement your routing strategy with existing capability

Your target:

```text
display model -> multiple upstream aggregators -> priority retry -> weighted split
```

This mostly maps to current New API behavior:

- display model: user-facing model name
- real model: channel `model_mapping`
- primary/backup: `priority`
- load split inside same priority: `weight`
- retry count: `RetryTimes`
- unhealthy channel auto-disable: auto ban + retryable status codes

Example:

```text
Display model: GPT-5.5 Pro

Channel A
- name: n1n-primary
- base_url: https://api.n1n.ai/v1
- models: GPT-5.5 Pro
- model_mapping: {"GPT-5.5 Pro":"openai/gpt-4o"}
- priority: 10
- weight: 100

Channel B
- name: openrouter-backup
- base_url: https://openrouter.ai/api/v1
- models: GPT-5.5 Pro
- model_mapping: {"GPT-5.5 Pro":"openai/gpt-4o"}
- priority: 9
- weight: 100
```

Behavior:

- priority 10 is tried first
- if it fails with retryable status, relay retry selects lower-priority backup
- channels with the same priority are chosen by weight

## 7. What still needs custom development

The following items are not fully productized for your business naming and reporting needs:

- a dedicated `aggregator` channel type label in admin UI
- cleaner admin wording focused on third-party aggregators only
- explicit per-channel metrics for:
  - `429 count`
  - `insufficient quota count`
  - `5xx count`
  - average latency by upstream aggregator
- business-facing model catalog for names like `Claude Creative`, `Nano Banana`, `Veo Fast`
- task-based adapter shells for `fal.ai` and `DashScope`
- clearer cost/profit views by:
  - upstream cost
  - user charge
  - profit

## 8. Brand replacement approach

Do not hard-delete upstream project attribution in source files.

For your commercial deployment, use runtime site settings first:

- `SystemName`
- `Logo`
- `Footer`
- docs link

These are already exposed by the backend status/config API and frontend store.

Relevant files:

- [controller/misc.go](D:\项目2\zhongzhuanzhan\controller\misc.go)
- [model/option.go](D:\项目2\zhongzhuanzhan\model\option.go)
- [web/default/src/main.tsx](D:\项目2\zhongzhuanzhan\web\default\src\main.tsx)

## 9. Current local blockers

This machine is not ready for full local acceptance yet:

- `docker` command is missing
- GitHub access is unstable
- I have not been able to run a real local container boot
- I have not verified a live `/v1/chat/completions` call in this environment

So the deployment files are prepared, but local runtime verification is still pending.
