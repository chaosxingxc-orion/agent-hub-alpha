import type { FieldConsumption } from "../types";

const digest = (v: unknown): v is string => typeof v === "string" && /^sha256:[0-9a-f]{64}$/.test(v);
const name = (v: unknown): v is string => typeof v === "string" && /^[A-Za-z0-9_][A-Za-z0-9_.:-]{0,127}$/.test(v);
const names = (v: unknown): v is string[] => Array.isArray(v) && v.length > 0 && v.length <= 32 && v.every(name) && new Set(v).size === v.length;
const esc = (v: string) => v.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]!));

function valid(value: unknown): value is FieldConsumption {
  if (!value || typeof value !== "object") return false;
  const v = value as FieldConsumption;
  if (!["pending", "input_bound", "committed"].includes(v.status) || !digest(v.execution_binding_sha256) || !digest(v.node_contract_sha256) || !names(v.input_fields) || !names(v.output_fields)) return false;
  if (v.status !== "pending" && !digest(v.input_digest)) return false;
  if (v.status === "committed" && !digest(v.artifact_digest)) return false;
  if (v.dependencies !== undefined && (!Array.isArray(v.dependencies) || v.dependencies.length > 64 || !v.dependencies.every((d) => d && name(d.id) && (d.kind === "material" ? digest(d.sha256) : d.kind === "step" && digest(d.node_execution_id) && digest(d.artifact_digest))))) return false;
  return true;
}

export function consumptionLabel(value: unknown): string | undefined {
  if (!value) return undefined;
  if (!valid(value)) return "字段记录待核验";
  return { pending: "字段待绑定", input_bound: "输入字段已绑定", committed: "字段已正式提交" }[value.status];
}

// Select finite identifiers and digests only. Extra fields, values and secrets
// never enter this projection, including its copy/rendered HTML surfaces.
export function consumptionDetailHTML(value: unknown): string {
  if (!value) return "";
  if (!valid(value)) return '<section class="h03-consumption"><h5>字段记录待核验</h5><p>缺少完整绑定证据。</p></section>';
  const row = (label: string, text: string) => `<div><dt>${esc(label)}</dt><dd>${esc(text)}</dd></div>`;
  const sources = (value.dependencies ?? []).map((d) => row(d.kind === "material" ? "材料来源" : "已提交上游", d.id + " · " + (d.kind === "material" ? d.sha256 : d.artifact_digest))).join("");
  return `<section class="h03-consumption" data-consumption-status="${value.status}"><h5>${consumptionLabel(value)}</h5><dl>${row("输入字段", value.input_fields.join("、"))}${row("输出字段", value.output_fields.join("、"))}${sources}${row("采纳绑定", value.execution_binding_sha256)}${row("节点合同", value.node_contract_sha256)}${value.input_digest ? row("实际输入", value.input_digest) : ""}${value.artifact_digest ? row("正式产物", value.artifact_digest) : ""}</dl><p>字段提交反映工程记录；专业业务验收需另行确认。</p></section>`;
}
