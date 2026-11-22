// Turns a graph snapshot into React Flow nodes and edges laid out with dagre,
// the same way WorkSwarm's SwarmFlow graph does. Node heights are fixed by the
// stylesheet (see the matching numbers in style.css), so the layout is known
// before anything is drawn and never has to be measured and redone.
import dagre from "@dagrejs/dagre";
import type { Edge, Node } from "@xyflow/react";
import { condLabel, finished, started, type Graph, type GraphNode, type Tone } from "../types";

export const NODE_W = 212;
const nodeH = (running: boolean) => 74 + (running ? 22 : 0);

export type Line = { kind: "act" | "note"; text: string; tone?: Tone };
export interface ExpertData extends Record<string, unknown> {
  n: GraphNode; status: string; denied: boolean; line?: Line; delay: number;
}
export type EdgeState = "idle" | "active" | "done" | "skip" | "bad";
export interface EdgeData extends Record<string, unknown> { state: EdgeState; label?: string; labelOff?: boolean; delay: number }

export interface Built { nodes: Node[]; edges: Edge<EdgeData>[]; width: number; height: number; key: string }

// Enter delays are fixed the first time an element appears, so later
// snapshots never restart its entrance.
export class Delays {
  private readonly seen = new Map<string, number>();
  first = true;
  get(id: string, compute: () => number) {
    let d = this.seen.get(id);
    if (d === undefined) { d = compute(); this.seen.set(id, d); }
    return d;
  }
}

export function build(g: Graph, delays: Delays): Built {
  const shown = g.run ? g.nodes.filter((n) => started(n.status)) : g.nodes;
  const plan = new Map(g.nodes.map((n) => [n.id, n]));
  const byId = new Map(shown.map((n) => [n.id, n]));
  const order = new Map(g.nodes.map((n, i) => [n.id, i]));
  const layerOf = new Map<string, number>();
  g.layers.forEach((l, i) => l.forEach((id) => layerOf.set(id, i)));

  // Draw only direct needs: an edge implied through another dependency is noise.
  const up = new Map<string, Set<string>>();
  const ancestors = (id: string): Set<string> => {
    let s = up.get(id);
    if (s) return s;
    s = new Set();
    up.set(id, s);
    for (const d of plan.get(id)?.deps ?? []) { s.add(d); ancestors(d).forEach((a) => s!.add(a)); }
    return s;
  };
  const linksAmong = (nodes: GraphNode[]) => {
    const ids = new Set(nodes.map((n) => n.id));
    const out: [string, GraphNode][] = [];
    for (const n of nodes) {
      const deps = (n.deps ?? []).filter((d) => ids.has(d));
      for (const d of deps) {
        if (n.when?.step !== d && deps.some((o) => o !== d && ancestors(o).has(d))) continue;
        out.push([d, n]);
      }
    }
    return out;
  };
  const links = linksAmong(shown);

  const size = new Map(g.nodes.map(n => [n.id, { w: NODE_W, h: nodeH(!!g.run) }]));

  const dg = new dagre.graphlib.Graph();
  dg.setDefaultEdgeLabel(() => ({}));
  dg.setGraph({ rankdir: "LR", nodesep: 16, ranksep: 64, marginx: 0, marginy: 0 });
  // Lay out the whole plan even while a run shows only the experts that have
  // started: each appears in its final place and the card keeps its size.
  for (const n of g.nodes) dg.setNode(n.id, { width: size.get(n.id)!.w, height: size.get(n.id)!.h });
  for (const [d, n] of linksAmong(g.nodes)) dg.setEdge(d, n.id);
  dagre.layout(dg);

  // dagre reorders a rank to cut crossings; keep the plan's order instead so
  // nodes do not swap places between snapshots, stacked around the rank center.
  const pos = new Map<string, { x: number; y: number }>();
  // Keep the plan order within each rank to avoid snapshot-to-snapshot swaps.
  const ranks = new Map<number, string[]>();
  for (const n of g.nodes) {
    const cx = Math.round(dg.node(n.id).x);
    (ranks.get(cx) ?? ranks.set(cx, []).get(cx)!).push(n.id);
  }
  let height = 0, width = 0;
  const heights: number[] = [];
  for (const ids of ranks.values()) heights.push(ids.reduce((s, id) => s + size.get(id)!.h, 0) + (ids.length - 1) * 16);
  const tallest = Math.max(...heights, 0);
  let left = Infinity;
  for (const [cx, ids] of ranks) for (const id of ids) left = Math.min(left, cx - size.get(id)!.w / 2);
  for (const [cx, ids] of ranks) {
    ids.sort((a, b) => order.get(a)! - order.get(b)!);
    const total = ids.reduce((s, id) => s + size.get(id)!.h, 0) + (ids.length - 1) * 16;
    let y = Math.round((tallest - total) / 2);
    for (const id of ids) {
      const x = Math.round(cx - size.get(id)!.w / 2 - left);
      pos.set(id, { x, y });
      y += size.get(id)!.h + 16;
      width = Math.max(width, x + size.get(id)!.w);
      height = Math.max(height, pos.get(id)!.y + size.get(id)!.h);
    }
  }

  const first = delays.first;
  const enter = (id: string, layer: number, i: number, hasUp: boolean) =>
    delays.get(id, () => (first ? layer * 110 + i * 30 : hasUp ? 260 : 0));

  const nodes: Node[] = [];
  shown.forEach((n, i) => {
    const p = pos.get(n.id)!, s = size.get(n.id)!;
    const hasUp = links.some(([, t]) => t.id === n.id);
    const delay = enter(n.id, layerOf.get(n.id) ?? 0, i, hasUp);
    nodes.push({ id: n.id, type: "expert", position: p, width: s.w, height: s.h, data: expertData(g, n, delay) });
  });

  const edges: Edge<EdgeData>[] = [];
  for (const [d, n] of links) {
    const src = byId.get(d)!;
    let state: EdgeState = "idle";
    if (g.run) {
      if (n.status === "skipped") state = "skip";
      else if (finished(src.status) && n.status === "running") state = "active";
      else if (finished(src.status) && finished(n.status)) state = "done";
    } else if (!g.ok && n.denied) state = "bad";
    const id = `${d}>${n.id}`;
    const cond = n.when && n.when.step === d ? n.when : undefined;
    edges.push({
      id, source: d, target: n.id, type: "flow",
      data: {
        state, delay: delays.get(id, () => (first ? 200 + (layerOf.get(d) ?? 0) * 110 : 0)),
        label: cond && condLabel(cond), labelOff: n.status === "skipped",
      },
    });
  }
  delays.first = false;
  return { nodes, edges, width, height, key: JSON.stringify([width, height, [...pos].map(([id, p]) => [id, p.x, p.y])]) };
}

const ended: Record<string, Line> = {
  done: { kind: "note", text: "已完成", tone: "ok" },
  reused: { kind: "note", text: "复用账本里的结果" },
  failed: { kind: "note", text: "执行失败", tone: "bad" },
  skipped: { kind: "note", text: "条件不满足，未执行", tone: "mute" },
};

// The first sentence of what a step produced, without the step tag and markdown.
function gist(output?: string): string {
  const text = (output ?? "").replace(/^\[step:[^\]]*\]\s*/, "");
  for (const raw of text.split("\n")) {
    const line = raw.replace(/\*\*/g, "").replace(/^\s*[-*#>]+\s*/, "").replace(/\s+/g, " ").trim();
    if (line) return line;
  }
  return "";
}

function lineOf(g: Graph, n: GraphNode, status: string): Line | undefined {
  if (!g.run) return undefined;
  if (status === "running") return { kind: "act", text: n.activity || "执行中" };
  if (status === "waiting") return { kind: "act", text: n.activity || "等待远端结果" };
  if (status === "unknown") return { kind: "note", text: n.activity || "结果待核对，不会自动重发", tone: "bad" };
  const said = finished(status) ? gist(n.output) : "";
  return said ? { kind: "note", text: said } : ended[status];
}

function expertData(g: Graph, n: GraphNode, delay: number): ExpertData {
  const status = n.status ?? "plan";
  return { n, status, delay, denied: !g.ok && !g.run && !!n.denied, line: lineOf(g, n, status) };
}
