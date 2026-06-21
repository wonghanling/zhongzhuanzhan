# OpenAI-compatible aggregator admin guide

This guide is for adding third-party OpenAI-compatible aggregator platforms in New API admin, including:

- `chinallmapi.com`
- `n1n.ai`
- `OpenRouter`
- `allapi`
- other OpenAI-compatible third-party platforms

Goal:

- fill fields in New API admin correctly
- save the channel successfully
- complete one real `/v1/chat/completions` call through a user API key

## Scope

Use this guide when the upstream platform is:

- third-party
- OpenAI-compatible
- uses API key auth
- exposes endpoints similar to:

```text
POST /v1/chat/completions
GET /v1/models
```

In New API admin, these platforms should usually be added as:

- channel type: `OpenAI`

## Where to add the channel

Go to:

```text
Admin -> Channels -> Add Channel
```

## Field-by-field guide

### 1. 渠道名称

- 字段名: `Name`
- 应该填什么: a clear internal name for the upstream channel
- 示例值:
  - `chinallmapi-primary`
  - `n1n-primary`
  - `openrouter-backup`
  - `allapi-gpt-main`
- 注意事项:
  - use a name that tells you provider and role
  - include `primary` / `backup` when multiple providers serve the same display model
- 常见错误:
  - using vague names like `channel1`
  - not indicating whether it is primary or backup

### 2. 渠道类型

- 字段名: `Type`
- 应该填什么: `OpenAI`
- 示例值: `OpenAI`
- 注意事项:
  - for third-party OpenAI-compatible platforms, use `OpenAI`
  - do not use provider-specific official channel types unless the upstream is really that official API
- 常见错误:
  - selecting `Azure`, `Anthropic`, `Gemini`, or other official provider types for an aggregator

### 3. Base URL

- 字段名: `Base URL`
- 应该填什么: the real upstream API root from the provider docs
- 示例值:
  - `<BASE_URL>`
  - `https://<PROVIDER_HOST>/v1`
- 注意事项:
  - use the exact API root required by that provider
  - some providers want the root with `/v1`
  - some providers want the root without `/v1`
  - do not guess, follow the provider docs or dashboard
- 常见错误:
  - base URL 多写 `/v1`
  - base URL 少写 `/v1`
  - adding a trailing slash when the provider is sensitive to path joins
  - copying the website homepage URL instead of the API base URL

### 4. API Key

- 字段名: `Key`
- 应该填什么: provider API key
- 示例值: `<PROVIDER_API_KEY>`
- 注意事项:
  - paste only the real upstream key
  - if the provider rotates keys, update the channel immediately
  - if one provider gives multiple keys, you can later manage them as separate channels or multi-key channels
- 常见错误:
  - pasting a user token from New API instead of the upstream key
  - extra spaces or line breaks
  - expired key

### 5. 模型列表

- 字段名: `Models`
- 应该填什么: the model names this channel exposes to New API routing
- 示例值:
  - `<MODEL_NAME>`
  - `GPT-5.5 Pro`
  - `GPT-5.5 Pro,Claude Creative,Gemini Vision`
- 注意事项:
  - these can be your display model names
  - if you use display names, pair them with `model_mapping`
  - separate multiple models with commas
- 常见错误:
  - model names and mapping keys do not match
  - adding spaces or separators inconsistently
  - using upstream real model names here but display names in curl requests

### 6. 分组

- 字段名: `Group`
- 应该填什么: which user group can use this channel
- 示例值:
  - `default`
  - `vip`
  - `default,vip`
- 注意事项:
  - if most users will use it, start with `default`
  - the user token group and channel group must match
- 常见错误:
  - New API 分组没匹配
  - channel is in `vip` but the user token only has `default`

### 7. 倍率

- 字段名: not primarily a channel field; usually configured in system pricing / model ratio settings
- 应该填什么: set charge ratio in pricing settings after channel creation
- 示例值:
  - display model `GPT-5.5 Pro` -> ratio `2`
- 注意事项:
  - do not confuse provider cost with user billing ratio
  - if your policy is `用户扣费 = 上游成本 × 2`, set it in model price/ratio settings
- 常见错误:
  - expecting a single channel form field to control all billing logic
  - forgetting to configure model ratio after the channel works

### 8. 优先级

- 字段名: `Priority`
- 应该填什么: which upstream should be tried first
- 示例值:
  - `100` for primary
  - `90` for first backup
  - `80` for second backup
- 注意事项:
  - higher value means higher priority
  - for the same display model, use descending priority across providers
- 常见错误:
  - setting all providers to the same priority when you actually want clear fallback order

### 9. 权重

- 字段名: `Weight`
- 应该填什么: traffic share among channels with the same priority
- 示例值:
  - `100`
  - `50`
  - `10`
- 注意事项:
  - weight only matters when multiple channels share the same priority
  - use high weight for main traffic, low weight for canary/testing
- 常见错误:
  - expecting weight to override priority
  - using weight without understanding that different priority channels are chosen first by priority

### 10. 状态

- 字段名: `Status`
- 应该填什么: enabled
- 示例值: `Enabled`
- 注意事项:
  - save the channel as enabled for first acceptance
- 常见错误:
  - forgetting to enable the channel after saving

### 11. 自动禁用

- 字段名: `Auto Ban`
- 应该填什么: enabled
- 示例值: `Enabled`
- 注意事项:
  - recommended for third-party aggregators
  - helps remove temporarily broken upstreams from routing
- 常见错误:
  - disabling auto-ban and then wondering why a broken provider keeps being selected

### 12. 重试策略

- 字段名: not a single channel field; controlled by system retry settings plus channel priority
- 应该填什么:
  - retry count in deployment preset or settings
  - retryable status codes
  - multiple providers with descending priority
- 示例值:
  - `BOLUO_RETRY_TIMES=2`
  - `BOLUO_AUTO_RETRY_STATUS_CODES=429,500-503`
- 注意事项:
  - retry works best when you have at least one backup provider
  - priority decides fallback order
- 常见错误:
  - only one channel exists, so retry has nowhere useful to fail over
  - retryable status codes are not configured for your failure pattern

### 13. 备注

- 字段名: `Remark`
- 应该填什么: internal operator notes
- 示例值:
  - `third-party aggregator`
  - `primary GPT route`
  - `backup for GPT-5.5 Pro`
- 注意事项:
  - use it for provider role, account note, or migration note
- 常见错误:
  - leaving no note and forgetting the purpose of the channel later

## Model mapping

When users should see your marketing model names instead of real upstream model IDs, use `model_mapping`.

Example:

```json
{
  "GPT-5.5 Pro": "gpt-4o",
  "Claude Creative": "claude-3-5-sonnet",
  "Gemini Vision": "gemini-2.0-flash"
}
```

Rule:

- left side = display model name used by your users
- right side = real upstream model name required by the provider

## Template A: chinallmapi.com

Use this as a starting template:

```text
渠道名称 / Name: chinallmapi-primary
渠道类型 / Type: OpenAI
Base URL: <BASE_URL>
API Key: <PROVIDER_API_KEY>
模型列表 / Models: GPT-5.5 Pro
分组 / Group: default
优先级 / Priority: 100
权重 / Weight: 100
状态 / Status: Enabled
自动禁用 / Auto Ban: Enabled
备注 / Remark: chinallmapi third-party aggregator
```

Model mapping:

```json
{
  "GPT-5.5 Pro": "<MODEL_NAME>"
}
```

## Template B: n1n.ai

```text
渠道名称 / Name: n1n-primary
渠道类型 / Type: OpenAI
Base URL: <BASE_URL>
API Key: <PROVIDER_API_KEY>
模型列表 / Models: GPT-5.5 Pro,Claude Creative
分组 / Group: default
优先级 / Priority: 100
权重 / Weight: 100
状态 / Status: Enabled
自动禁用 / Auto Ban: Enabled
备注 / Remark: n1n.ai primary aggregator
```

Model mapping:

```json
{
  "GPT-5.5 Pro": "<MODEL_NAME>",
  "Claude Creative": "<MODEL_NAME>"
}
```

## Template C: OpenRouter

```text
渠道名称 / Name: openrouter-backup
渠道类型 / Type: OpenAI
Base URL: <BASE_URL>
API Key: <PROVIDER_API_KEY>
模型列表 / Models: GPT-5.5 Pro,Claude Creative,Gemini Vision
分组 / Group: default
优先级 / Priority: 90
权重 / Weight: 100
状态 / Status: Enabled
自动禁用 / Auto Ban: Enabled
备注 / Remark: OpenRouter backup aggregator
```

Model mapping:

```json
{
  "GPT-5.5 Pro": "<MODEL_NAME>",
  "Claude Creative": "<MODEL_NAME>",
  "Gemini Vision": "<MODEL_NAME>"
}
```

## Template D: generic third-party aggregator

```text
渠道名称 / Name: aggregator-generic
渠道类型 / Type: OpenAI
Base URL: <BASE_URL>
API Key: <PROVIDER_API_KEY>
模型列表 / Models: <MODEL_NAME>
分组 / Group: default
优先级 / Priority: 100
权重 / Weight: 100
状态 / Status: Enabled
自动禁用 / Auto Ban: Enabled
备注 / Remark: generic openai-compatible aggregator
```

Model mapping:

```json
{
  "<MODEL_NAME>": "<MODEL_NAME>"
}
```

## Suggested first routing layout

Example for one display model:

```text
GPT-5.5 Pro
- chinallmapi-primary   priority 100 weight 100
- n1n-backup            priority 90  weight 100
- openrouter-backup     priority 80  weight 100
- allapi-backup         priority 70  weight 100
```

## curl test example

```bash
curl http://服务器公网IP/v1/chat/completions \
  -H "Authorization: Bearer <USER_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "<DISPLAY_MODEL_OR_REAL_MODEL>",
    "messages": [
      {"role": "user", "content": "hello"}
    ]
  }'
```

Recommended for your platform:

- use the display model name if you configured `model_mapping`
- use the real model name only if you intentionally expose the upstream name directly

## Common error troubleshooting

### Base URL 多写 `/v1`

Symptoms:

- upstream returns 404
- request path becomes duplicated like `/v1/v1/chat/completions`

Fix:

- remove the extra `/v1`
- use exactly the provider-documented API root

### Base URL 少写 `/v1`

Symptoms:

- upstream returns 404 or endpoint not found

Fix:

- add `/v1` if the provider requires it

### 模型名和上游不一致

Symptoms:

- model not found
- upstream returns invalid model

Fix:

- check `Models`
- check `model_mapping`
- ensure the right-side model is the real upstream model id

### API Key 错误

Symptoms:

- 401 or 403 from upstream

Fix:

- replace with the correct upstream key
- remove accidental spaces

### 上游余额不足

Symptoms:

- upstream returns insufficient quota / insufficient balance

Fix:

- recharge the upstream provider
- let New API fall back to a backup channel if configured

### 上游限流 429

Symptoms:

- upstream returns 429

Fix:

- ensure retry count and retryable status codes are configured
- reduce pressure or add backup providers

### New API 分组没匹配

Symptoms:

- user request cannot find an available channel

Fix:

- check channel `Group`
- check user/token group

### 用户 Key 没有权限

Symptoms:

- access denied
- token has no model access

Fix:

- check token permissions
- check token model restrictions
- check user group and channel group alignment

### 渠道被自动禁用

Symptoms:

- channel shows auto-disabled
- requests stop routing to it

Fix:

- inspect logs
- fix the upstream cause
- manually re-enable the channel after correction

## First channel acceptance standard

All of these should be true:

- channel saves successfully
- test request succeeds
- user key can call the model
- backend logs show the request
- quota deduction record is visible
- if a request fails, backend shows a readable failure reason
