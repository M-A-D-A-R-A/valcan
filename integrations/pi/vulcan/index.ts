/**
 * vulcan-pi — Vulcan v0 observer for Pi.
 *
 * Records every tool call as a structured run event and ships it to Vigil.
 * This is M1: observe only. It never blocks or mutates tool calls.
 *
 * Configure with environment variables:
 *   VIGIL_BASE_URL    default http://localhost:8080
 *   VIGIL_PROJECT_ID  required (proj_...)
 *   VIGIL_INGEST_KEY  required (vigil_...)
 *
 * If the project id or ingest key is missing, the extension stays disabled.
 *
 * Usage: pi -e /path/to/valcan/integrations/pi/vulcan/index.ts
 */

import { createHash } from "node:crypto";
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

const PROMPT_SUMMARY_CHARS = 400;
const MAX_RESULT_BYTES = 2_000_000; // never serialize more than this per step

interface RunState {
	id: string;
	startedAt: string;
	turn: number;
	prompt: string;
	steps: number;
	failedSteps: number;
	toolStarts: Map<string, { tool: string; startedMs: number }>;
}

interface PendingEvent {
	name: string;
	level: "info" | "warn" | "error";
	traceId?: string;
	attrs: Record<string, unknown>;
	body?: Record<string, unknown>;
}

function sha12(value: string): string {
	return createHash("sha256").update(value).digest("hex").slice(0, 12);
}

export default function (pi: ExtensionAPI) {
	const baseUrl = (process.env.VIGIL_BASE_URL ?? "http://localhost:8080").replace(/\/+$/, "");
	const projectId = process.env.VIGIL_PROJECT_ID ?? "";
	const ingestKey = process.env.VIGIL_INGEST_KEY ?? "";

	if (!projectId || !ingestKey) {
		// Disabled: no Vigil project configured. Pi runs normally.
		return;
	}

	let run: RunState | null = null;
	let currentModel = "";
	let emitFailures = 0;
	let notified = false;

	function emit(event: PendingEvent): void {
		const envelope = {
			schema_version: 1,
			project_id: projectId,
			kind: "log",
			ts: new Date().toISOString(),
			source: "vulcan-pi",
			level: event.level,
			name: event.name,
			...(event.traceId ? { trace_id: event.traceId } : {}),
			attrs: event.attrs,
			body: event.body ?? { message: event.name },
		};

		fetch(`${baseUrl}/api/ingest`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				Authorization: `Bearer ${ingestKey}`,
			},
			body: JSON.stringify(envelope),
		})
			.then((res) => {
				if (!res.ok) throw new Error(`vigil ${res.status}`);
				emitFailures = 0;
			})
			.catch(() => {
				emitFailures += 1;
				// Notify once per run so the user knows telemetry is failing,
				// but never interrupt the session for it.
				if (emitFailures >= 5 && !notified) {
					notified = true;
					console.error("[vulcan-pi] failed to reach Vigil; continuing without telemetry");
				}
			});
	}

	function newRunId(): string {
		return `run_${Date.now().toString(36)}${Math.random().toString(36).slice(2, 8)}`;
	}

	pi.on("model_select", (event) => {
		currentModel = `${event.model.provider}/${event.model.id}`;
	});

	pi.on("before_agent_start", (event) => {
		run = {
			id: newRunId(),
			startedAt: new Date().toISOString(),
			turn: 1,
			prompt: event.prompt,
			steps: 0,
			failedSteps: 0,
			toolStarts: new Map(),
		};
		emit({
			name: "run.started",
			level: "info",
			traceId: run.id,
			attrs: {
				source: "pi",
				model: currentModel || null,
				prompt_summary: event.prompt.slice(0, PROMPT_SUMMARY_CHARS),
				prompt_chars: event.prompt.length,
			},
			body: { message: "run started" },
		});
	});

	pi.on("tool_call", (event) => {
		if (!run) return;
		run.toolStarts.set(event.toolCallId, {
			tool: event.toolName,
			startedMs: Date.now(),
		});
	});

	pi.on("tool_result", (event, ctx) => {
		if (!run) return;
		const start = run.toolStarts.get(event.toolCallId);
		run.toolStarts.delete(event.toolCallId);
		run.steps += 1;
		if (event.isError) run.failedSteps += 1;

		const text = event.content
			.filter((block): block is { type: "text"; text: string } => block.type === "text")
			.map((block) => block.text)
			.join("\n");
		const resultBytes = Math.min(Buffer.byteLength(text, "utf8"), MAX_RESULT_BYTES);

		emit({
			name: "run.step",
			level: event.isError ? "error" : "info",
			traceId: run.id,
			attrs: {
				turn: run.turn,
				tool: event.toolName,
				args_hash: sha12(JSON.stringify(event.input ?? {})),
				result_bytes: resultBytes,
				result_hash: sha12(text).slice(0, 12),
				duration_ms: start ? Date.now() - start.startedMs : null,
				is_error: event.isError,
				model: currentModel || null,
				cwd: ctx.cwd,
			},
			body: { message: `${event.toolName} ${event.isError ? "failed" : "ok"}` },
		});
	});

	pi.on("agent_end", (event) => {
		if (!run) return;

		let inputTokens = 0;
		let outputTokens = 0;
		let cacheReadTokens = 0;
		const byModel: Record<string, number> = {};

		for (const message of event.messages) {
			if (message.role !== "assistant") continue;
			const anyMessage = message as {
				model?: string;
				usage?: { input?: number; output?: number; cacheRead?: number; totalTokens?: number };
			};
			const usage = anyMessage.usage;
			if (!usage) continue;
			const input = usage.input ?? 0;
			const output = usage.output ?? 0;
			inputTokens += input + (usage.cacheRead ?? 0);
			outputTokens += output;
			cacheReadTokens += usage.cacheRead ?? 0;
			const model = anyMessage.model ?? currentModel ?? "unknown";
			byModel[model] = (byModel[model] ?? 0) + (usage.totalTokens ?? input + output);
		}

		const level = run.failedSteps > 0 ? "warn" : "info";
		emit({
			name: run.failedSteps > 0 ? "run.failed" : "run.completed",
			level,
			traceId: run.id,
			attrs: {
				source: "pi",
				steps: run.steps,
				failed_steps: run.failedSteps,
				input_tokens: inputTokens,
				output_tokens: outputTokens,
				cache_read_tokens: cacheReadTokens,
				total_tokens: inputTokens + outputTokens,
				models: byModel,
				duration_ms: Date.now() - Date.parse(run.startedAt),
			},
			body: { message: "run finished" },
		});
		run = null;
	});

	// If the session dies mid-run, close it out so analysis sees a terminal state.
	pi.on("session_shutdown", () => {
		if (!run) return;
		emit({
			name: "run.aborted",
			level: "warn",
			traceId: run.id,
			attrs: { source: "pi", steps: run.steps, failed_steps: run.failedSteps },
			body: { message: "session ended before run completed" },
		});
		run = null;
	});
}
