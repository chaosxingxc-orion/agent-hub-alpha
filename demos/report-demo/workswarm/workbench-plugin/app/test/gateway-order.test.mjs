// Exercise the real RPC client and revision gate with a controlled transport.
// Unlike the 25s deployment experiment, the first case keeps BOTH RPCs alive.
import { build } from 'esbuild';
import { test } from 'node:test';
import assert from 'node:assert/strict';
const bundle = await build({ stdin: { contents: 'export { Gateway } from "./src/workbench/gateway"; export { newer } from "./src/order";', resolveDir: process.cwd() }, bundle: true, write: false, format: 'esm', platform: 'node' });
const { Gateway, newer } = await import('data:text/javascript;base64,' + Buffer.from(bundle.outputFiles[0].text).toString('base64'));
const layoutBundle = await build({ stdin: { contents: 'export { build, Delays } from "./src/flow/layout";', resolveDir: process.cwd() }, bundle: true, write: false, format: 'esm', platform: 'node' });
const { build: layout, Delays } = await import('data:text/javascript;base64,' + Buffer.from(layoutBundle.outputFiles[0].text).toString('base64'));

const reportGraph = () => ({plan:'T-fixture',version:1,case:'annual-report',goal:'report',ok:true,hash:'fixture',
  layers:[['extract'],['business','quality'],['draft'],['check']],nodes:[
    {id:'extract',name:'extract',kind:'extract',task:'fixed material',deps:[]},
    {id:'business',name:'business',kind:'business',task:'business',deps:['extract']},
    {id:'quality',name:'quality',kind:'quality',task:'quality',deps:['extract']},
    {id:'draft',name:'draft',kind:'write',task:'join both branches',deps:['business','quality']},
    {id:'check',name:'check',kind:'check',task:'check draft',deps:['draft']}
  ]});

test('annual-report plan retains both analysis branches, join and stable positions',()=>{
  const g=reportGraph(), planned=layout(g,new Delays());
  assert.deepEqual(planned.nodes.map(n=>n.id),g.nodes.map(n=>n.id));
  assert.equal(planned.nodes.every(n=>n.type==='expert'),true);
  assert.deepEqual(planned.edges.map(e=>e.id),['extract>business','extract>quality','business>draft','quality>draft','draft>check']);
  const running=layout({...g,run:{id:g.plan,status:'running',leases:1,url:''},nodes:g.nodes.map(n=>({...n,status:'running'}))},new Delays());
  assert.deepEqual(running.nodes.map(n=>n.position.x),planned.nodes.map(n=>n.position.x));
  const node=id=>planned.nodes.find(n=>n.id===id);
  assert.ok(node('draft').position.x>node('business').position.x);
  assert.ok(node('draft').position.x>node('quality').position.x);
});

test('annual-report projection preserves unknown, authorized skip and condition labels',()=>{
  const g=reportGraph();g.run={id:g.plan,status:'waiting',leases:1,url:''};
  g.nodes=g.nodes.map(n=>({...n,status:n.id==='quality'?'unknown':'done',activity:n.id==='quality'?'retain remote task':undefined}));
  g.nodes.find(n=>n.id==='check').when={step:'draft',contains:'REVISE'};
  const built=layout(g,new Delays());
  assert.equal(built.nodes.find(n=>n.id==='quality').data.line.text,'retain remote task');
  assert.equal(built.edges.find(e=>e.id==='draft>check').data.label,'含 REVISE');
  g.nodes.find(n=>n.id==='check').status='skipped';
  const skipped=layout(g,new Delays());
  assert.equal(skipped.nodes.some(n=>n.id==='check'),false);
});

function transport(t) {
  const sent = [], timers = new Map();
  let socket, seq = 0;
  class Socket {
    static OPEN = 1;
    readyState = 1;
    constructor() { socket = this; queueMicrotask(() => this.onopen()); }
    send(data) { sent.push(JSON.parse(data)); }
  }
  for (const [key, value] of Object.entries({
    WebSocket: Socket, location: { protocol: 'http:', host: 'fixture' },
    localStorage: { getItem: () => null },
    window: { setTimeout: (fn) => { timers.set(++seq, fn); return seq; } },
    clearTimeout: (id) => timers.delete(id),
  })) {
    const descriptor = Object.getOwnPropertyDescriptor(globalThis, key);
    Object.defineProperty(globalThis, key, { value, configurable: true, writable: true });
    t.after(() => descriptor ? Object.defineProperty(globalThis, key, descriptor) : delete globalThis[key]);
  }
  return { sent, timers, reply: (id, rev) => socket.onmessage({ data: JSON.stringify({ type: 'res', id, ok: true, payload: { plan:"T-fixture",version:1,rev:String(rev), nodes: [] } }) }) };
}

test('two live RPCs resolve out of order; revision gate rejects the older graph', async (t) => {
  const wire = transport(t), gw = new Gateway();
  let shown, deliveries = 0;
  const accept = (g) => { deliveries++; if (newer(shown, g)) shown = g; };
  const older = gw.call('plugin.credit_workbench.graph').then(accept);
  const latest = gw.call('plugin.credit_workbench.graph').then(accept);
  await new Promise(setImmediate);
  assert.equal(wire.timers.size, 2); // neither request has expired
  wire.reply(wire.sent[1].id, 296);
  await latest;
  assert.equal(wire.timers.size, 1); // older request still active
  wire.reply(wire.sent[0].id, 247);
  await older;
  assert.equal(deliveries, 2); // both reached the revision gate
  assert.equal(shown.rev, '296');
});

test('expired RPC response is discarded before it can reach the revision gate', async (t) => {
  const wire = transport(t), gw = new Gateway();
  let deliveries = 0;
  const pending = gw.call('plugin.credit_workbench.graph').then(() => deliveries++);
  const rejected = assert.rejects(pending, { code: 'TIMEOUT' });
  await new Promise(setImmediate);
  wire.timers.values().next().value();
  await rejected;
  wire.reply(wire.sent[0].id, 247);
  assert.equal(deliveries, 0);
});

test('decimal generations above Number precision are ordered without loss',()=>{
 const g=rev=>({plan:'T-fixture',version:1,rev});
 assert.equal(newer(g('9007199254740993'),g('9007199254740992')),false);
 assert.equal(newer(g('9007199254740993'),g('9007199254740994')),true);
 assert.equal(newer(g('3'),g('invalid')),false);
});
