#!/usr/bin/env python3
"""Check persisted RDC/Gateway UI evidence through a real synthetic credit run.

Uses the already configured model. Creates one completed test run and one
unapproved plan; performs no bank operation and no component restart.
"""
import json
import time
from smoke_bus import CACHE, Planner, TEAM, environment_record, normal, redact, request


def main():
    evidence = {"scenario": "bus-ui", "started_at": time.time(), "snapshots": [], "environment": environment_record()}
    target = CACHE / "evidence" / "agent-bus-ui" / f"integration-{int(time.time())}.json"
    target.parent.mkdir(parents=True, exist_ok=True)
    try:
        normal(evidence)
        graph, _ = request('/api/graph?run=' + evidence['run_id'])
        assert graph['bus'] == {'planned_nodes': 5, 'services': 3, 'linked_tasks': 5, 'completed': 5, 'unresolved': 0, 'failed': 0, 'reused': 0}
        for node in graph['nodes']:
            if node['execution']['kind'] != 'agent_bus':
                assert 'remote' not in node
                continue
            r = node['remote']
            assert r['discovery']['checked_at'] and any(c['eligible'] and c['instance_id'] for c in r['discovery']['candidates'])
            gw = r['gateway']
            assert gw['source'] == 'java_host' and gw['path_mode'] == 'DIRECT'
            assert gw['send_count'] == 1 and gw['first_send']['rpc_id'] == r['request_id']
            assert gw['first_receipt']['outcome'] == 'task_received' and gw['first_receipt']['http_status'] == 200
            assert gw['last_exchange']['outcome'] == 'task_received'
            assert r['task_id'] and r['artifact_id'] and r['committed_at'] and r['agent_started_at']
        text = json.dumps(graph).lower()
        for private in ('capability_token', 'authorization', 'routehandle', 'gateway_exchanges'):
            assert private not in text, 'Private call data appeared in UI graph'
        evidence['graph'] = graph
        planner = Planner()
        plan = planner.tool('propose_plan', {'case_id': 'C006', 'goal': 'Agent Bus 页面检查：仅预览候选计划，等待人工批准', 'agents': TEAM})
        preview, _ = request(f'/api/graph?plan={plan["plan_id"]}&v={plan["version"]}')
        assert 'run' not in preview and preview['bus']['linked_tasks'] == 0
        assert all('remote' not in n for n in preview['nodes'])
        evidence['preview'] = preview
        evidence['checks'].update(discovery_snapshot=True, gateway_receipts=True, planned_vs_actual=True, no_private_projection=True)
        evidence['passed'] = True
        print(json.dumps({'run': evidence['run_id'], 'preview': plan['plan_id'], 'checks': evidence['checks']}, ensure_ascii=False))
    except Exception as error:
        evidence['passed'] = False
        evidence['error'] = str(error)
        raise
    finally:
        evidence['finished_at'] = time.time()
        target.write_text(json.dumps(redact(evidence), ensure_ascii=False, indent=2))
        print('Evidence:', target)


if __name__ == '__main__':
    main()
