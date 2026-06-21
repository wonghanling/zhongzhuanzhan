# chinallmapi.com upstream setup

Use `chinallmapi.com` as an OpenAI-compatible third-party aggregator channel.

## Channel type

In New API admin, use:

- channel type: `OpenAI`

Reason:

- this project already supports custom `base_url`
- OpenAI type already supports `model_mapping`
- it also participates in priority and retry selection logic

## Required fields

Fill these fields in admin:

```text
Name: chinallmapi-primary
Type: OpenAI
Base URL: <your chinallmapi base url>
Key: <your API key>
Models: GPT-5.5 Pro,Claude Creative,Gemini Vision
Group: default
Priority: 100
Weight: 100
Auto Ban: enabled
```

## Model mapping example

Use business-facing names on the left, real upstream names on the right:

```json
{
  "GPT-5.5 Pro": "gpt-4o",
  "Claude Creative": "claude-3-5-sonnet",
  "Gemini Vision": "gemini-2.0-flash"
}
```

Adjust right-side model IDs to whatever `chinallmapi.com` actually exposes.

## Base URL rule

For OpenAI-compatible providers in this project:

- prefer the provider root without trailing slash
- if the provider expects `/v1`, use the exact API root they document
- do not guess if their docs specify a different prefix

Examples:

```text
https://chinallmapi.com/v1
https://api.chinallmapi.com/v1
```

Use the one from the provider docs or dashboard.

## Suggested backup layout

For your MVP:

```text
GPT-5.5 Pro
- chinallmapi-primary      priority 100 weight 100
- openrouter-backup        priority 90  weight 100
- allapi-backup            priority 80  weight 100
```

## Validation checklist

After saving the channel:

1. Use the channel test action in admin.
2. Fetch upstream models if the provider supports model listing.
3. Create a user token.
4. Call:

```bash
curl https://your-domain/v1/chat/completions \
  -H "Authorization: Bearer USER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "GPT-5.5 Pro",
    "messages": [
      {"role": "user", "content": "hello"}
    ]
  }'
```

5. Confirm logs show the request used the `chinallmapi-primary` channel.

## Notes

I have not verified the exact `chinallmapi.com` base URL or model IDs yet.
You should plug in the real values from the provider dashboard or docs.
