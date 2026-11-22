// W04 /v1/workbench/tasks/{id} projection consumed by this W02 adapter.

export interface RunIf { step: string; contains: string; negate?: boolean }
export type { BusSummary, CapabilityBinding, ExecutionBinding, GatewayExchange, RemoteExecution } from "./flow/bus-detail";
import type { BusSummary, ExecutionBinding, RemoteExecution } from "./flow/bus-detail";
export interface GraphNode {
  id: string; name: string; kind: string; role?: string; task: string; deps?: string[]; when?: RunIf;
  contract?: string; denied?: string; status?: string; activity?: string; output?: string;
  // Set when the node runs through Agent Bus: the capability it is bound to and what its call recorded.
  execution?: ExecutionBinding; remote?: RemoteExecution;
}
export interface Check { name: string; ok: boolean; detail: string; items?: string[] }
export interface BlockedReason { code: string; message: string }
export interface Graph {
  plan: string; version: number; case: string; goal: string; layers: string[][]; nodes: GraphNode[];
  ok: boolean; problems?: string[]; hash: string; approve?: string; confirm?: string; superseded?: number;
  run?: { id: string; status: string; decision?: string; receipt?: string; leases: number; stalled?: boolean; url: string };
  verify?: { ok: boolean; summary: string; checks: Check[] };
  selection?: {id:string;digest:string;eligibility:unknown[];ranking:unknown[];selection:string[];composition:unknown};
  reports?: {node:string;artifact_id:string;digest:string;output:unknown;check_outcome:string;can_approve:boolean}[];
  task_status?: string; delivery_receipts?: unknown[];
  blocked_reason?: BlockedReason;
  human_review?: {url:string;checks:{node:string;check_artifact:string;object_digest:string}[]};
  live?: string;
  bus?: BusSummary;
  rev?: string; // state revision the snapshot was read at (see order.ts)
}
export interface Options {
  // Approve in the card: resolves to an error message, or undefined when approved.
  approve?: (g: Graph, user: string, password: string) => Promise<string | undefined>;
  onApproved?: () => void;
}

export const finished = (s?: string) => s === "done" || s === "reused";
// Once a run exists, an expert appears when it starts. A remote node may be
// waiting for its result, or have one that still has to be checked (unknown).
export const started = (s?: string) => s === "running" || s === "waiting" || s === "unknown" || s === "failed" || finished(s);

export type Tone = "ok" | "bad" | "mute";

export function condLabel(w: RunIf) {
  if (!w.negate && /GRANT/.test(w.contains)) return "同意时";
  return (w.negate ? "不含 " : "含 ") + w.contains;
}
