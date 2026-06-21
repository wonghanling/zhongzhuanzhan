# Boluo API Gateway phase plan

## Current baseline

This repository is already close to your MVP. For phase 1, do not rewrite the core routing architecture.

Use the existing New API capabilities first:

- user registration
- token creation
- quota deduction
- logs
- model pricing ratio
- channel priority
- weighted channel selection
- channel model mapping
- automatic retry
- channel auto-disable

## Mapping from your requirements to current code

### Already available now

1. User creates API key
2. User calls `/v1/chat/completions`
3. Admin creates multiple upstream channels
4. One display model can map to multiple upstream channels
5. Channel priority and weight routing
6. Retry after failure
7. User quota and usage logs
8. Channel enable/disable
9. Model ratio and model price configuration

### Needs light customization

1. Admin wording should emphasize third-party aggregators only
2. Display-model naming templates for your marketing names
3. Cleaner reporting for upstream aggregator health
4. More explicit reason breakdown for `429`, `5xx`, and quota failures
5. Fal.ai and DashScope task-adapter placeholders

### Should wait until after MVP

1. Complex membership system
2. Full custom frontend redesign
3. Deep integration with other Boluolab systems
4. Large provider matrix
5. Full video task workflow

## Recommended implementation order

### Step 1

Deploy the unmodified baseline on Azure and verify:

- admin login
- user creation
- token creation
- one OpenAI-compatible aggregator works end to end

### Step 2

Configure three aggregator channels in admin:

- n1n.ai
- OpenRouter
- allapi

Use:

- same display model name on multiple channels
- same model mapping
- descending priority
- same or tuned weights

### Step 3

Set business-facing model names in channel model mapping strategy:

- `GPT-5.5 Pro`
- `Claude Creative`
- `Gemini Vision`
- `Nano Banana`
- `Veo Fast`
- `Kling Pro`

### Step 4

Tune retry and health strategy:

- `RetryTimes = 2`
- enable auto-ban on aggregator channels
- configure automatic retry status codes
- configure automatic disable status codes

### Step 5

Add a thin business layer for aggregator-focused admin UX:

- rename visible channel helper text
- add an internal guideline for allowed upstream types
- add dashboards/filters around third-party aggregators

## Proposed MVP data conventions

### Upstream channel naming

Use explicit names:

```text
n1n-gpt-primary
openrouter-gpt-backup
allapi-claude-primary
openrouter-claude-backup
```

### Model mapping convention

For each channel, map display name to real upstream name:

```json
{
  "GPT-5.5 Pro": "openai/gpt-4o",
  "Claude Creative": "anthropic/claude-3.5-sonnet",
  "Gemini Vision": "google/gemini-2.0-flash"
}
```

### Priority convention

```text
100 = primary
90 = first backup
80 = second backup
```

### Weight convention

Inside the same priority:

```text
100 = main traffic
50 = half share
10 = canary
```

## Minimal code changes I recommend next

1. Add deployment docs and Azure compose
2. Add a business setup guide for third-party aggregators
3. Add small admin/UI wording changes where they do not conflict with protected upstream attribution
4. Add fal.ai and DashScope adapter interfaces without full workflow yet
5. Add more explicit channel health counters if existing logs are not enough

## What I have not verified yet

Because this machine cannot run Docker right now, these remain unverified locally:

1. live container startup
2. admin login in browser
3. real aggregator upstream request
4. live quota deduction and log display

These should be verified either on:

- an Azure VM directly
- or a local machine with Docker installed
