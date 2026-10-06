# Vulcan v0 — Quickstart

The first draft of M1 (Observe) is implemented. This is the small harness on top
of Pi: it records every tool call to Vigil and reports where tokens were
wasted. See [`v0-scope.md`](./v0-scope.md) for the full scope.

## Layout

```
cmd/vulcan/                 Go CLI
  main.go                     `vulcan runs` and `vulcan waste`
internal/vigil/client.go    Vigil read client (GET /api/logs)
internal/waste/waste.go     run reconstruction + waste classification
integrations/pi/vulcan/     Pi extension (observer)
  index.ts                    emits run.started / run.step / run.completed|failed|aborted
```

## 1. Start Vigil

```sh
vigil serve
# create a project, capture the ingest key
curl -X POST http://localhost:8080/api/projects \
  -H 'Content-Type: application/json' -d '{"name":"vulcan"}'
# → {"project":{"id":"proj_..."},"ingest_key":"vigil_..."}
```

## 2. Run Pi with the extension

```sh
VIGIL_PROJECT_ID=proj_... VIGIL_INGEST_KEY=vigil_... \
  pi -e /path/to/valcan/integrations/pi/vulcan/index.ts
```

The extension is **observe-only**: it never blocks or mutates tool calls. If the
env vars are missing it stays disabled and Pi runs normally.

## 3. Analyze waste

```sh
export VIGIL_PROJECT_ID=proj_...

go run ./cmd/vulcan runs                 # list runs
go run ./cmd/vulcan waste --window 24h   # waste report
```

Example output:

```
runs               1
steps              4
result bytes seen  55800
addressable bytes  30800

CLASS      COUNT  BYTES
repeat     1      5000
failed     1      800
oversized  1      25000
```

## What M1 classifies (deterministic, from the trace)

| Class | Rule |
|---|---|
| `repeat` | same tool + `args_hash` called again within a run |
| `failed` | tool call returned an error |
| `oversized` | result larger than `--size-cap` (default 20 KB) |

`unused result` (result never referenced by a later edit) needs edit
correlation and lands in M2+.

## Next milestones

- [ ] M2 — router v0 (rules + escalation) + per-model cost
- [ ] M3 — bandit router, A/B vs rules
- [ ] M4 — policy v1 (repeat-block, size caps, nudges) + Δtokens
- [ ] M5 — evidence-gated fork

## Typecheck the extension (optional)

```sh
cd integrations/pi/vulcan
ln -s /opt/homebrew/lib/node_modules/@earendil-works/pi-coding-agent \
  node_modules/@earendil-works/pi-coding-agent
# then run tsc --noEmit with a tsconfig, or rely on Pi's jiti loader at runtime
```
