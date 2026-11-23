import test from 'node:test';
import assert from 'node:assert/strict';
import { build } from 'esbuild';

const source = await build({entryPoints:['src/flow/h03-consumption.ts'],bundle:true,write:false,format:'esm',platform:'node'});
const {consumptionLabel,consumptionDetailHTML} = await import('data:text/javascript;base64,'+Buffer.from(source.outputFiles[0].contents).toString('base64'));
const d='sha256:'+ 'a'.repeat(64);
const base={status:'pending',execution_binding_sha256:d,node_contract_sha256:d,input_fields:['total','reference'],output_fields:['difference']};

test('legacy nodes and bound inputs do not claim formal consumption',()=>{
  assert.equal(consumptionDetailHTML(undefined),'');
  assert.equal(consumptionLabel(undefined),undefined);
  assert.equal(consumptionLabel(base),'字段待绑定');
  assert.equal(consumptionLabel({...base,status:'input_bound',input_digest:d}),'输入字段已绑定');
  assert.doesNotMatch(consumptionDetailHTML({...base,status:'input_bound',input_digest:d}),/字段已正式提交/);
});
test('formal records display actual identifiers and preserve business boundary',()=>{
  const html=consumptionDetailHTML({...base,status:'committed',input_digest:d,artifact_digest:d,dependencies:[{kind:'step',id:'sum',node_execution_id:d,artifact_digest:d}]});
  assert.match(html,/字段已正式提交/);assert.match(html,/已提交上游/);assert.match(html,/sum/);assert.match(html,/专业业务验收需另行确认/);
});
test('missing or unknown digest and field records remain unverified',()=>{
  for(const value of [{...base,status:'committed'}, {...base,status:'other'}, {...base,input_fields:[]}, {...base,input_fields:['total','total']}, {...base,input_fields:Array(33).fill('x')}, {...base,node_contract_sha256:'unknown'}, {...base,input_fields:['<img onerror=x>']}]){
    const html=consumptionDetailHTML(value);assert.match(html,/待核验/);assert.doesNotMatch(html,/<img|onerror|正式提交/);
  }
});
test('private values and tokens never enter the finite projection',()=>{
  const html=consumptionDetailHTML({...base,secret:'private-test',value:'100',path:'/private/path',dependencies:[{kind:'material',id:'source',sha256:'sha256:'+'b'.repeat(64),token:'hidden-test'}]});
  assert.match(html,/材料来源/);assert.doesNotMatch(html,/private-test|hidden-test|private\/path|value/);
});

test('original Go material dependency shape preserves consumption status',()=>{
  const value={...base,status:'committed',input_digest:d,artifact_digest:d,dependencies:[{kind:'material',id:'source',sha256:d}]};
  assert.equal(consumptionLabel(value),'字段已正式提交');
  assert.match(consumptionDetailHTML(value),/材料来源/);
  assert.match(consumptionDetailHTML(value),/source/);
  assert.match(consumptionDetailHTML({...value,dependencies:[{kind:'material',id:'source',sha256:'a'.repeat(64)}]}),/待核验/);
});
