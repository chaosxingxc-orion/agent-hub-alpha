package main

const builderBuildBody = `<section id="builder-build"><h2>构建场景 Agent</h2>
<p>Intent GATE · 选择已登记的参考任务，构建后独立核验产物。专业业务验收和生产启用需另行确认。</p>
<label>参考任务 <select id="build-alias" aria-label="参考任务"></select></label>
<button type="button" id="build-submit" disabled>提交构建</button>
<button type="button" id="build-refresh">查看原任务</button>
<p id="build-status" role="status">正在读取可用任务</p>
<ul id="build-history"></ul>
<details><summary>查看任务与候选标识</summary><pre id="build-detail">尚无记录</pre></details>
</section><script>(()=>{
const base='{{.Base}}/api/builder-actions/', select=document.getElementById('build-alias');
const status=document.getElementById('build-status'), detail=document.getElementById('build-detail'), history=document.getElementById('build-history');
const submit=document.getElementById('build-submit'), refreshButton=document.getElementById('build-refresh');
const permissionMessage='当前没有查看或构建权限；权限恢复后请重新加载页面。';
let original=new Map(), permissionLost=false, generation=0;
function deny(){permissionLost=true;generation++;original.clear();detail.textContent='';history.replaceChildren();select.replaceChildren();select.disabled=true;submit.disabled=true;refreshButton.disabled=true;status.textContent=permissionMessage;}
function current(value){if(permissionLost||value!==generation)throw new Error(permissionMessage);}
function failed(error){status.textContent=permissionLost?permissionMessage:error.message;}
async function call(path,body){
 const active=generation;current(active);
 const r=await fetch(base+path,{method:body?'POST':'GET',credentials:'same-origin',cache:'no-store',...(body?{headers:{'Content-Type':'application/json'},body:JSON.stringify(body)}:{})});
 if(r.status===403){deny();throw new Error(permissionMessage);}
 current(active);if(!r.ok)throw new Error('暂时无法核对原任务，请稍后查看');const value=await r.json();current(active);return value;
}
async function show(id){
 const active=generation, value=await call('commands/'+encodeURIComponent(id));current(active);
 detail.textContent=JSON.stringify(value,null,2);
 if(value.operation){const states={running:'正在构建',waiting:'等待补充信息',failed:'构建失败',cancelled:'已取消',stopped:'已停止',completed:'构建任务已完成',unknown:'执行情况待核对'};
  status.textContent=(states[value.operation.status]||'执行情况待核对')+(value.evidence?.status==='verified'?'；已有独立工程验收证据':'；独立验收尚未确认')+'。资源清理：'+({complete:'已确认',failed:'失败',pending:'进行中',not_run:'未运行',unknown:'待核对'}[value.operation.cleanup]||'待核对');
 }else status.textContent='已保存原任务关联，执行情况待核对；查看会继续核对同一个任务。';
}
async function refresh(){
 const active=generation, value=await call('builds');current(active);original=new Map(value.intents.map(x=>[x.alias,x.command_id]));
 const selected=select.value;select.replaceChildren();for(const alias of value.aliases){const option=document.createElement('option');option.value=alias;option.textContent=alias;select.append(option);}if(value.aliases.includes(selected))select.value=selected;
 history.replaceChildren();for(const row of value.intents){const li=document.createElement('li'),button=document.createElement('button');button.type='button';button.textContent=row.alias+' · 查看原任务';button.onclick=()=>show(row.command_id).catch(failed);li.append(button);history.append(li);}
 submit.disabled=!select.value;submit.textContent=original.has(select.value)?'核对原构建':'提交构建';
 if(original.has(select.value))await show(original.get(select.value));else status.textContent=select.value?'可以提交这个已登记的参考任务':'当前没有可用的参考任务';
}
select.onchange=()=>{if(permissionLost)return;submit.textContent=original.has(select.value)?'核对原构建':'提交构建';};
submit.onclick=async()=>{if(permissionLost)return;submit.disabled=true;try{await call('build',{alias:select.value});await refresh();}catch(e){failed(e);if(!permissionLost){try{await refresh();}catch(recovery){failed(recovery);}}}finally{submit.disabled=permissionLost||!select.value;}};
refreshButton.onclick=()=>refresh().catch(failed);
refresh().catch(failed);
})();</script>`
