// Native cases use the original Hub and live backend; no intercepted routes/RPC.
// The explicitly named component case uses synthetic ordering and never qualifies native evidence.
import { createRequire } from 'node:module';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { dirname, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
import assert from 'node:assert/strict';
const { chromium }=createRequire(import.meta.url)('playwright');
const options=new Map();
for(let i=2;i<process.argv.length;i+=2){
 if(!['--hub-url','--output','--alias','--case'].includes(process.argv[i])||!process.argv[i+1]||options.has(process.argv[i]))throw new Error('invalid finite browser arguments');
 options.set(process.argv[i],process.argv[i+1]);
}
const mode=options.get('--case')||'build';
assert(['build','disabled','denied','completed','revoked','component'].includes(mode));
const root=resolve(dirname(fileURLToPath(import.meta.url)),'../../../../../');
const evidence=resolve(root,'.specify/feature/HUB-NEXT/evidence'),output=resolve(options.get('--output')||'');
assert(output.startsWith(evidence+sep)&&!existsSync(output),'new owned evidence directory required');
const native=mode!=='component',alias=options.get('--alias');let url,password;
if(native){
 url=new URL(options.get('--hub-url'));
 assert(['127.0.0.1','localhost','[::1]'].includes(url.hostname)&&url.protocol==='http:'&&!url.username&&!url.password&&!url.search&&!url.hash,'only operator-owned loopback Hub');
 assert(typeof alias==='string'&&/^[A-Za-z0-9][A-Za-z0-9._-]{0,119}$/.test(alias));
 password=process.env.BUILDER_E2E_PASSWORD;assert(password&&password!=='demo','explicit local login required');
}
await mkdir(output,{recursive:true});
const report={schema:'builder-build-live-page/v1',mode,alias,result:'FAIL',native,browser_posts:0,post_counts:{login:0,build_alias:0,build_off_probe:0,preflight:0},original_backend:native?'live; no route interception':'COMPONENT ONLY; synthetic fetch ordering; no HTTP/RPC',checks:[],responses:[]};
let browser;
const save=(name,value)=>writeFile(resolve(output,name),JSON.stringify(value,null,2)+'\n');
const pause=ms=>new Promise(r=>setTimeout(r,ms));
async function login(context){
 const page=await context.newPage();page.setDefaultTimeout(15000);
 page.on('request',r=>{if(r.method()!=='POST')return;const p=new URL(r.url()).pathname;if(p==='/ui/login')report.post_counts.login++;if(p==='/api/builder-actions/build'){report.browser_posts++;report.post_counts.build_alias++;}if(p==='/api/builder-hub/preflight')report.post_counts.preflight++;});
 page.on('response',r=>{const p=new URL(r.url()).pathname;if(p.startsWith('/api/builder'))report.responses.push({path:p,status:r.status()});});
 await page.goto(new URL('ui/login?next=/ui/builder',url).href);
 await page.locator('select[name="user"]').selectOption('alice');await page.locator('input[name="password"]').fill(password);await page.locator('form button').click();await page.waitForURL('**/ui/builder');
 return page;
}
async function query(page){
 const received=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/builder-hub/capabilities',{timeout:15000});
 await page.locator('#support').click();const response=await received;assert.equal(response.status(),200);const body=await response.json();assert.match(JSON.stringify(body),/intent/);await save('original-query.json',{status:200,body});
 report.checks.push('original H01 real capabilities200');
}
async function completed(page){
 await page.waitForFunction(()=>{try{const x=JSON.parse(document.getElementById('build-detail').textContent);return x.operation?.status==='completed'&&x.evidence?.status==='verified'&&x.receipt?.candidate_ref;}catch{return false;}},null,{timeout:15000});
 const detail=JSON.parse(await page.locator('#build-detail').textContent());assert.equal(detail.alias,alias);assert.match(detail.command_id,/^[A-Za-z0-9][A-Za-z0-9._-]{0,119}$/);
 assert.deepEqual(detail.receipt.task_ref,detail.operation.task_ref);assert.deepEqual(detail.receipt.candidate_ref,detail.operation.candidate_ref);assert.deepEqual(detail.operation.candidate_ref,detail.evidence.candidate_ref);assert.deepEqual(detail.operation.evidence_ref,detail.evidence.evidence_ref);
 assert.equal(detail.evidence.kind,'engineering');assert.equal(detail.operation.quality,'not_run');assert(['complete','failed','pending','not_run','unknown'].includes(detail.operation.cleanup));
 await page.locator('#builder-build details').evaluate(e=>e.open=true);assert(await page.locator('#build-detail').isVisible());
 assert.match(await page.locator('#build-status').textContent(),/\u6784\u5efa\u4efb\u52a1\u5df2\u5b8c\u6210.*\u72ec\u7acb\u5de5\u7a0b\u9a8c\u6536/);
 await save('completed-detail.json',detail);await page.screenshot({path:resolve(output,'completed-expanded.png'),fullPage:true});report.command_id=detail.command_id;report.checks.push('completed receipt/candidate/evidence expanded with bound refs; quality and cleanup explicit');return detail;
}
async function assertCleared(page,old){
 await page.waitForFunction(()=>document.getElementById('build-detail').textContent===''&&document.getElementById('build-history').children.length===0&&document.getElementById('build-submit').disabled,null,{timeout:15000});
 const visible=await page.locator('body').innerText();assert(!visible.includes(old.command_id)&&!visible.includes(old.receipt.candidate_ref.sha256));assert.match(await page.locator('#build-status').textContent(),/\u6743\u9650/);
}
async function component(){
 const source=await readFile(resolve(root,'agent-festival/go/cmd/credit-demo/builder_build_ui.go'),'utf8');const html=source.slice(source.indexOf('`')+1,source.lastIndexOf('`')).replaceAll('{{.Base}}','');
 for(const scenario of ['late-command','late-builds','late-json','late-error']){
  const page=await browser.newPage();page.setDefaultTimeout(10000);
  await page.setContent(`<script>
  window.scenario=${JSON.stringify(scenario)};window.commandCalls=0;window.buildCalls=0;window.release=null;
  const row={alias:'finite-fixture',command_id:'h04.component'};
  const detail={alias:row.alias,command_id:row.command_id,receipt:{candidate_ref:{id:'SYNTHETIC-PRIVATE',sha256:'a'.repeat(64)}},operation:{status:'completed',cleanup:'unknown'},evidence:{status:'verified'}};
  const listing=()=>new Response(JSON.stringify({aliases:[row.alias],intents:[row]}),{status:200});
  const success=()=>new Response(JSON.stringify(detail),{status:200});
  window.fetch=async path=>{
   if(path.endsWith('/builds')){window.buildCalls++;if(scenario==='late-builds'&&window.buildCalls===2)return await new Promise(r=>window.release=()=>r(listing()));return listing();}
   if(path.includes('/commands/')){window.commandCalls++;if(window.commandCalls===1){
    if(scenario==='late-builds')return success();
    if(scenario==='late-json')return {ok:true,status:200,json:()=>new Promise(r=>window.release=()=>r(detail))};
    return await new Promise((r,reject)=>window.release=()=>scenario==='late-error'?reject(new Error('old failure')):r(success()));}
    return new Response('{"error":{"code":"builder-build-unavailable"}}',{status:403});}
   throw new Error('unexpected component request');};
  </script>${html}`);
  await page.waitForFunction(()=>window.commandCalls===1&&document.getElementById('build-history').children.length===1);
  if(scenario==='late-builds'){await page.locator('#build-refresh').click();await page.waitForFunction(()=>window.buildCalls===2);}
  await page.locator('#build-history button').click();await page.waitForFunction(()=>document.getElementById('build-status').textContent.includes('\u6743\u9650'));
  assert.equal(await page.locator('#build-detail').textContent(),'');await page.evaluate(()=>window.release());
  await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
  assert.equal(await page.locator('#build-detail').textContent(),'');assert.equal(await page.locator('#build-history li').count(),0);assert.equal(await page.locator('#build-alias option').count(),0);assert(await page.locator('#build-submit').isDisabled());assert(await page.locator('#build-refresh').isDisabled());assert.match(await page.locator('#build-status').textContent(),/\u91cd\u65b0\u52a0\u8f7d/);
  report.checks.push('COMPONENT ONLY '+scenario+' cannot refill after current403');await page.close();
 }
 // A fresh page can receive current permission and the same retained identity again.
 const fresh=await browser.newPage();await fresh.setContent(`<script>window.fetch=async p=>new Response(JSON.stringify(p.endsWith('/builds')?{aliases:['finite-fixture'],intents:[{alias:'finite-fixture',command_id:'h04.component'}]}:{command_id:'h04.component',operation:{status:'completed',cleanup:'unknown'},evidence:{status:'verified'}}),{status:200});</script>${html}`);
 await fresh.waitForFunction(()=>document.getElementById('build-detail').textContent.includes('h04.component'),null,{timeout:10000});assert(!(await fresh.locator('#build-refresh').isDisabled()));await fresh.close();report.checks.push('COMPONENT ONLY new page rechecks restored permission without old-page unlock');
}
try{
 browser=await chromium.launch({headless:true,...(process.env.BROWSER_EXECUTABLE?{executablePath:process.env.BROWSER_EXECUTABLE}:{})});
 if(!native)await component();else{
  const context=await browser.newContext(),page=await login(context);
  if(mode==='disabled'){
   assert.equal(await page.locator('#builder-build').count(),0);await query(page);
   const response=await context.request.post(new URL('api/builder-actions/build',url).href,{data:{alias},headers:{Origin:url.origin}});report.post_counts.build_off_probe++;assert.equal(response.status(),404);await save('query-and-off.json',{H01_status:200,build_POST_status:404,build_panel_count:0});await page.screenshot({path:resolve(output,'build-off-original-query.png'),fullPage:true});report.checks.push('build switch off removes panel and rejects API404');
  }else if(mode==='completed'||mode==='revoked'){
   await query(page);const old=await completed(page);
   if(mode==='revoked'){
    // Two-stage operator coordination only. This script never runs a shell, changes grants, or accepts credentials through HTTP.
    await save('revocation-ready.json',{schema:'builder-build-ui-revocation/v1',command_id:old.command_id,alias,requires:'trusted original Host operator revokes current read grant, then writes revocation-applied.json with read_revoked:true'});
    const deadline=Date.now()+60000;let applied=false;
    while(Date.now()<deadline){if(existsSync(resolve(output,'revocation-applied.json'))){assert.deepEqual(JSON.parse(await readFile(resolve(output,'revocation-applied.json'),'utf8')),{read_revoked:true});applied=true;break;}await pause(200);}assert(applied,'trusted operator revocation checkpoint not received');
    const response=page.waitForResponse(r=>new URL(r.url()).pathname.startsWith('/api/builder-actions/')&&r.status()===403,{timeout:15000});await page.locator('#build-refresh').click();const denied=await response;await assertCleared(page,old);
    // Chromium may leave an unconsumed fetch-error body unavailable to response.text; independently GET the same original path.
    const probe=await context.request.get(denied.url());assert.equal(probe.status(),403);const body=await probe.text();assert(!body.includes(old.command_id)&&!body.includes(old.receipt.candidate_ref.sha256));
    await save('real-403.json',{page_status:403,body_probe_status:403,body_source:'additional same-cookie same-path original GET',body,detail:'',history_count:0,button_disabled:true});await page.screenshot({path:resolve(output,'real-403-hidden.png'),fullPage:true});report.checks.push('real current Host grant revocation: real403 clears already rendered detail/history; safe same-path403 body');
   }
  }else if(mode==='denied'){
   await page.locator('#build-alias').selectOption(alias);await page.locator('#build-submit').click();await page.waitForFunction(()=>document.getElementById('build-status')?.textContent.includes('\u6743\u9650'),null,{timeout:15000});assert.equal(await page.locator('#build-detail').textContent(),'');report.checks.push('current denied response has no receipt/evidence');
  }else{
   await page.waitForFunction(()=>!document.getElementById('build-submit')?.disabled,null,{timeout:15000});await page.locator('#build-alias').selectOption(alias);await page.locator('#build-submit').click();await page.waitForFunction(()=>document.getElementById('build-history')?.children.length>0,null,{timeout:15000});
   const firstResponse=await context.request.get(new URL('api/builder-actions/builds',url).href);assert.equal(firstResponse.status(),200);const first=(await firstResponse.json()).intents.find(x=>x.alias===alias);assert(first?.command_id);await page.screenshot({path:resolve(output,'first-browser.png'),fullPage:true});await context.close();
   const second=await browser.newContext(),recovered=await login(second);await recovered.waitForFunction(()=>document.getElementById('build-history')?.children.length>0,null,{timeout:15000});
   const next=await second.request.get(new URL('api/builder-actions/builds',url).href);assert.equal((await next.json()).intents.find(x=>x.alias===alias)?.command_id,first.command_id);await recovered.locator('#build-alias').selectOption(alias);await recovered.locator('#build-submit').click();await recovered.waitForFunction(()=>!document.getElementById('build-submit')?.disabled,null,{timeout:15000});
   const final=await second.request.get(new URL('api/builder-actions/builds',url).href);assert.equal((await final.json()).intents.filter(x=>x.alias===alias).length,1);await recovered.screenshot({path:resolve(output,'new-browser-recovery.png'),fullPage:true});report.command_id=first.command_id;report.checks.push('original login and finite alias submission','new browser recovers same command','repeat submission retains one durable association');await second.close();
  }
  if(mode!=='build')await context.close();
 }
 report.host_send_count='independent original Host counters required; browser POST count is not Host sends';report.result='PASS';
}catch(error){report.error=String(error);throw error;}
finally{await save('browser-result.json',report);if(browser)await Promise.race([browser.close(),pause(10000).then(()=>{throw new Error('owned browser close timeout');})]);}
