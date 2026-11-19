const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const web = path.join(__dirname, '../control-go/web');
const guard = fs.readFileSync(path.join(web, 'projection.js'), 'utf8');
function deferred() { let resolve; const promise = new Promise(r => resolve = r); return {promise, resolve}; }
function snapshot(generation, task_id='T-ui') {return {task_id, generation, status:generation, graph_revision:'g1', graph_digest:'g', graph:{goal:'fixture',nodes:[]},nodes:{}};}
function page(name) {
  const elements = new Map();
  function element() {return {textContent:'',value:'',children:[],style:{},append(...items){this.children.push(...items);},appendChild(item){this.children.push(item);},replaceChildren(...items){this.children=items;},insertBefore(){}};}
  const intervals=[];
  const host={postMessage(){}};
  const context=vm.createContext({console,BigInt,Map,Promise,setTimeout:()=>1,clearTimeout(){},setInterval:fn=>intervals.push(fn),
    location:{pathname:'/tasks/T-ui'},sessionStorage:{getItem:()=> 'fixture-csrf',setItem(){}},window:{parent:host,innerWidth:1000,addEventListener(){}},
    document:{createElement:element,body:element(),documentElement:{scrollHeight:100},querySelector:element,getElementById(id){if(!elements.has(id))elements.set(id,element());return elements.get(id);}},
    fetch:async()=>({ok:true,json:async()=>snapshot('1')})});
  const html=fs.readFileSync(path.join(web,name),'utf8');
  const source=html.match(/<script>([\s\S]*?)<\/script>/)[1].replace('/* PROJECTION_GUARD */',guard);
  if(name==='task.html') context.fetch=async url=>({ok:true,json:async()=>url.endsWith('/T-ui')?snapshot('1'):[]});
  vm.runInContext(source,context);
  return {context,elements,intervals,run:code=>vm.runInContext(code,context)};
}
test('actual graph renderer rejects old large generations and foreign tasks',()=>{
  const p=page('graph.html');p.context.fresh=snapshot('9007199254740993');p.context.old=snapshot('9007199254740992');
  p.run("projection.selectTask('T-ui'); render(fresh); render(old)");
  assert.match(p.elements.get('summary').textContent,/9007199254740993/);
  p.context.foreign=snapshot('9999999999999999','T-other');p.run('render(foreign)');
  assert.match(p.elements.get('summary').textContent,/T-ui/);
});
test('actual graph polling discards response after selected task changes',async()=>{
  const p=page('graph.html'), response=deferred();
  p.run("projection.selectTask('T-ui');liveURL='/projection';viewGrant='fixture'");
  p.context.fetch=async()=>({ok:true,json:()=>response.promise});
  const poll=p.intervals[0]();
  p.run("projection.selectTask('T-new')");response.resolve(snapshot('20'));
  await poll;assert.equal(p.elements.has('summary'),false);
});
test('actual confirmation page keeps newer task when delayed request returns',async()=>{
  const p=page('task.html');await new Promise(r=>setImmediate(r));
  const old=deferred();let reads=0;
  p.context.fetch=async url=>({ok:true,json:async()=>url.endsWith('/T-ui')?(++reads===1?old.promise:snapshot('9007199254740993')):[]});
  const slow=p.run('refresh()');const fast=p.run('refresh()');await fast;
  old.resolve(snapshot('9007199254740992'));await slow;
  assert.equal(p.elements.get('status').textContent,'9007199254740993');
});
test('actual confirmation handler coalesces double-click command',async()=>{
  const p=page('task.html');await new Promise(r=>setImmediate(r));
  const accepted=deferred();let writes=0;
  p.context.fetch=async(url,opts)=>({ok:true,json:async()=>opts.method==='POST'?(writes++,await accepted.promise):url.endsWith('/T-ui')?snapshot('2'):[]});
  const handler=p.elements.get('admit').onclick;
  const one=handler(),two=handler();await new Promise(r=>setImmediate(r));assert.equal(writes,1);
  accepted.resolve({status:'QUEUED'});await Promise.all([one,two]);
});
test('failed command releases in-flight guard for explicit retry',async()=>{
  const p=page('graph.html');let calls=0;p.context.command=()=>{calls++;return Promise.reject(Error('fixture'));};
  await assert.rejects(p.run("projection.runCommand('retry',command)"));
  await assert.rejects(p.run("projection.runCommand('retry',command)"));assert.equal(calls,2);
});

test('trusted review preserves the current draft over polling and a rejected submission', async () => {
  const p = page('task.html');
  await new Promise(resolve => setImmediate(resolve));

  let generation = 1;
  let artifact = 'check-a';
  let digest = 'sha256:report-a';
  let review = null;
  const writes = [];
  const snapshotForReview = () => ({
    ...snapshot(String(++generation)),
    nodes: {check: {result: {valid: true, artifact_id: artifact, output: {outcome: 'HUMAN', object_digest: digest}}, review}},
  });
  const form = () => p.run('reviewList.children[0]?.children ?? []');
  p.context.fetch = async (url, options) => {
    if (options.method === 'POST') {
      writes.push(JSON.parse(options.body));
      return {ok: false, json: async () => ({error: {code: 'RETRY', message: 'try again'}})};
    }
    return {ok: true, json: async () => url.endsWith('/T-ui') ? snapshotForReview() : []};
  };

  await p.run('refresh()');
  const reason = form()[2];
  reason.value = '已对照原材料的第十四页，核验完成';
  reason.oninput?.();
  for (let i = 0; i < 2; i++) {
    p.intervals[0]();
    await new Promise(resolve => setImmediate(resolve));
  }
  assert.equal(form()[2], reason);
  assert.equal(form()[2].value, reason.value);
  await form()[3].onclick();
  assert.deepEqual(writes, [{
    check_artifact: 'check-a', object_digest: 'sha256:report-a', decision: 'ACCEPT', reason: reason.value,
  }]);
  await p.run('refresh()');
  assert.equal(form()[2].value, reason.value);

  artifact = 'check-b';
  digest = 'sha256:report-b';
  await p.run('refresh()');
  assert.equal(form()[2].value, '');
  review = {receipt: 'human-review-b'};
  await p.run('refresh()');
  assert.equal(form().length, 2);
});
