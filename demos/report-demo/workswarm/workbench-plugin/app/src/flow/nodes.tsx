// Annual-report nodes and data edges, using the W02 workbench styling.
import { memo, type CSSProperties } from "react";
import { BaseEdge, EdgeLabelRenderer, getBezierPath, Handle, Position, type EdgeProps, type NodeProps } from "@xyflow/react";
import { Ban, Circle, CircleCheck, CircleDashed, CircleDot, CircleHelp, CircleX, Clock, Loader2 } from "lucide-react";
import type { EdgeData, ExpertData, Line } from "./layout";

// A denied node shakes once it has landed: the second delay is for that.
const enter = (delay: number, shake = false): CSSProperties => ({ animationDelay: shake ? `${delay}ms, ${delay + 900}ms` : `${delay}ms` });

export function StatusIcon({ status, denied }: { status: string; denied?: boolean }) {
  // Keyed by status so a change remounts the icon and plays its pop.
  if (denied) return <Ban key="denied" className="ci pop t-bad" />;
  switch (status) {
    case "running": return <Loader2 key={status} className="ci spin t-run" />;
    case "waiting": return <Clock key={status} className="ci pop t-warn" />;
    case "unknown": return <CircleHelp key={status} className="ci pop t-warn" />;
    case "done": return <CircleCheck key={status} className="ci pop t-ok" />;
    case "reused": return <CircleCheck key={status} className="ci pop t-reuse" />;
    case "failed": return <CircleX key={status} className="ci pop t-bad" />;
    case "skipped": return <CircleDashed key={status} className="ci t-mute" />;
    case "pending": return <CircleDot key={status} className="ci t-mute" />;
  }
  return <Circle key="plan" className="ci t-mute" />;
}

function StatusLine({ line }: { line: Line }) {
  // Keyed by its text so a new activity slides in instead of just swapping.
  const k = `${line.kind}:${line.text}`;
  switch (line.kind) {
    case "act": return <div key={k} className="cn-row act" title={line.text}><i /><span>{line.text}</span></div>;
  }
  return <div key={k} className={`cn-row note t-${line.tone ?? "mute"}`} title={line.text}><span>{line.text}</span></div>;
}

export const ExpertNode = memo(({ data }: NodeProps) => {
  const d = data as ExpertData;
  const { n, status } = d;
  const tags: JSX.Element[] = [];
  if (n.execution?.kind === "agent_bus" || n.remote) tags.push(<span key="bus" className="tag bus" title="经 Agent Bus 网关 DIRECT 调用远端 Agent">总线</span>);
  if (status === "unknown") tags.push(<span key="uk" className="tag warn pop">待核对</span>);
  if (status === "reused") tags.push(<span key="re" className="tag reuse pop">复用</span>);
  if (d.denied) tags.push(<span key="dn" className="tag bad pop">无权限</span>);
  return (
    <div className={`cn s-${status} k-${n.kind}${d.denied ? " denied" : ""}`} style={enter(d.delay, d.denied)}>
      <Handle type="target" position={Position.Left} isConnectable={false} />
      <Handle type="source" position={Position.Right} isConnectable={false} />
      <div className="cn-head">
        <StatusIcon status={status} denied={d.denied} />
        <span className="cn-name">{n.name}</span>
        {tags}
      </div>
      <div className="cn-task" title={n.task}>{n.task}</div>
      {d.line && <StatusLine line={d.line} />}
    </div>
  );
});

export const FlowEdge = memo((p: EdgeProps) => {
  const d = p.data as EdgeData;
  const [path, lx, ly] = getBezierPath(p);
  return (
    <>
      <BaseEdge id={p.id} path={path} className={`ce s-${d.state}`} style={{ animationDelay: `${d.delay}ms` }} pathLength={1} />
      {d.state === "active" && <path className={`ce-flow s-${d.state}`} d={path} />}
      {d.label && (
        <EdgeLabelRenderer>
          <div className={`ce-label${d.labelOff ? " off" : ""}`}
            style={{ transform: `translate(-50%, -50%) translate(${lx}px, ${ly}px)`, animationDelay: `${d.delay + 300}ms` }}>
            {d.label}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
});

export const nodeTypes = { expert: ExpertNode };
export const edgeTypes = { flow: FlowEdge };
