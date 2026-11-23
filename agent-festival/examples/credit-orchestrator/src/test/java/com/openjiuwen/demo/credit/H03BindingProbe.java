package com.openjiuwen.demo.credit;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import java.util.*;
/** Frozen accepted fixture vectors; no Main, Worker service or real network. */
public final class H03BindingProbe {

    static final JsonNode FIXTURE = H03NodeCodec.parse("{\"sum\":{\"contract\":{\"schema_version\":\"hub-node-contract/v1\",\"input_fields\":[{\"name\":\"component_a\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"},{\"name\":\"component_b\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"},{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"}],\"output_fields\":[{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"},{\"name\":\"total\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"}]},\"input\":{\"schema_version\":\"hub-bound-input/v1\",\"execution_binding_sha256\":\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\",\"node_contract_sha256\":\"sha256:1529f647e38c976923e21f0362cb94a0796b543318ad64e131940349250219e6\",\"plan_hash\":\"sha256:ad6922eb0624c6a6e3e13fe80c7b7fc8ffff24d67f7579ce60c39df78f5852ec\",\"run_id\":\"R1\",\"node_execution_id\":\"sha256:592757eb3061ad13c9b3515dbe55103f9ee2b5cb747a9d0fe4cd263e5a677675\",\"node_path\":\"sum\",\"attempt\":1,\"fields\":[{\"name\":\"component_a\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"40\"},{\"name\":\"component_b\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"60\"},{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"95\"}],\"dependencies\":[{\"kind\":\"material\",\"id\":\"engineering-material-a\",\"sha256\":\"sha256:d24deb3bc00530e2d9530f27100eff56b725b1ca7d6e76250d9199cea3532060\"}]},\"output\":{\"schema_version\":\"hub-node-output/v1\",\"execution_binding_sha256\":\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\",\"node_contract_sha256\":\"sha256:1529f647e38c976923e21f0362cb94a0796b543318ad64e131940349250219e6\",\"input_digest\":\"sha256:5a44234942a2cdeeaf5e94f4be13b71801d393c8f4855a5be86ae874df673221\",\"fields\":[{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"95\"},{\"name\":\"total\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"100\"}]},\"preimage\":{\"input\":\"{\\\"attempt\\\":1,\\\"dependencies\\\":[{\\\"id\\\":\\\"engineering-material-a\\\",\\\"kind\\\":\\\"material\\\",\\\"sha256\\\":\\\"sha256:d24deb3bc00530e2d9530f27100eff56b725b1ca7d6e76250d9199cea3532060\\\"}],\\\"execution_binding_sha256\\\":\\\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\\\",\\\"fields\\\":[{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"component_a\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"40\\\"},{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"component_b\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"60\\\"},{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"reference\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"95\\\"}],\\\"node_contract_sha256\\\":\\\"sha256:1529f647e38c976923e21f0362cb94a0796b543318ad64e131940349250219e6\\\",\\\"node_execution_id\\\":\\\"sha256:592757eb3061ad13c9b3515dbe55103f9ee2b5cb747a9d0fe4cd263e5a677675\\\",\\\"node_path\\\":\\\"sum\\\",\\\"plan_hash\\\":\\\"sha256:ad6922eb0624c6a6e3e13fe80c7b7fc8ffff24d67f7579ce60c39df78f5852ec\\\",\\\"run_id\\\":\\\"R1\\\",\\\"schema_version\\\":\\\"hub-bound-input/v1\\\"}\",\"goal\":\"engineering arithmetic only。范围：engineering-entity，consolidated口径，FY2025 年；任务重点：reconciliation\",\"binding\":{\"agent_id\":\"h03-engineering-sum\",\"version\":\"0.1.0\",\"card_digest\":\"sha256:8a8a431f00ff8f9b54afb8dd4d2b864255dfea9852941f539bd7cc2fe1b67ea1\",\"contract_digest\":\"sha256:30e3204060ed57bd6583c057838f540ad2ec0474974d1890359844e927f2a7ec\"},\"model\":\"engineering-no-model\"},\"artifact_digest\":\"sha256:5033b744de6cd2f37388620f8d1aa48e30508d4dd8bbdb8dbf8dcb8f360e4173\"},\"reconcile\":{\"contract\":{\"schema_version\":\"hub-node-contract/v1\",\"input_fields\":[{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"},{\"name\":\"total\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"}],\"output_fields\":[{\"name\":\"difference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"}]},\"input\":{\"schema_version\":\"hub-bound-input/v1\",\"execution_binding_sha256\":\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\",\"node_contract_sha256\":\"sha256:ed98d4b8c49fb6ad8784664c461752ae70e2ce17f6ac3af120679c8a10ab283d\",\"plan_hash\":\"sha256:ad6922eb0624c6a6e3e13fe80c7b7fc8ffff24d67f7579ce60c39df78f5852ec\",\"run_id\":\"R1\",\"node_execution_id\":\"sha256:3700f123898c4e63a719ce0fa125e0bc20aff132bff69b5abd20af483a4ffc88\",\"node_path\":\"reconcile\",\"attempt\":1,\"fields\":[{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"95\"},{\"name\":\"total\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"100\"}],\"dependencies\":[{\"kind\":\"step\",\"id\":\"sum\",\"node_execution_id\":\"sha256:592757eb3061ad13c9b3515dbe55103f9ee2b5cb747a9d0fe4cd263e5a677675\",\"artifact_digest\":\"sha256:5033b744de6cd2f37388620f8d1aa48e30508d4dd8bbdb8dbf8dcb8f360e4173\"}]},\"output\":{\"schema_version\":\"hub-node-output/v1\",\"execution_binding_sha256\":\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\",\"node_contract_sha256\":\"sha256:ed98d4b8c49fb6ad8784664c461752ae70e2ce17f6ac3af120679c8a10ab283d\",\"input_digest\":\"sha256:4fa1e38110ce8902cf6921f9aca98c2cd6dca79c8ca1011a58685171df499eff\",\"fields\":[{\"name\":\"difference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"5\"}]},\"preimage\":{\"input\":\"{\\\"attempt\\\":1,\\\"dependencies\\\":[{\\\"artifact_digest\\\":\\\"sha256:5033b744de6cd2f37388620f8d1aa48e30508d4dd8bbdb8dbf8dcb8f360e4173\\\",\\\"id\\\":\\\"sum\\\",\\\"kind\\\":\\\"step\\\",\\\"node_execution_id\\\":\\\"sha256:592757eb3061ad13c9b3515dbe55103f9ee2b5cb747a9d0fe4cd263e5a677675\\\"}],\\\"execution_binding_sha256\\\":\\\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\\\",\\\"fields\\\":[{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"reference\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"95\\\"},{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"total\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"100\\\"}],\\\"node_contract_sha256\\\":\\\"sha256:ed98d4b8c49fb6ad8784664c461752ae70e2ce17f6ac3af120679c8a10ab283d\\\",\\\"node_execution_id\\\":\\\"sha256:3700f123898c4e63a719ce0fa125e0bc20aff132bff69b5abd20af483a4ffc88\\\",\\\"node_path\\\":\\\"reconcile\\\",\\\"plan_hash\\\":\\\"sha256:ad6922eb0624c6a6e3e13fe80c7b7fc8ffff24d67f7579ce60c39df78f5852ec\\\",\\\"run_id\\\":\\\"R1\\\",\\\"schema_version\\\":\\\"hub-bound-input/v1\\\"}\",\"goal\":\"engineering arithmetic only。范围：engineering-entity，consolidated口径，FY2025 年；任务重点：reconciliation\",\"binding\":{\"agent_id\":\"h03-engineering-reconcile\",\"version\":\"0.1.0\",\"card_digest\":\"sha256:5c8c2312be1a661f97cbfcf71dd778aa6986d753e5aa3e1ad77333671b350be3\",\"contract_digest\":\"sha256:9b998fedda163fe87f94b11491f54d75219e1426cc93ba16e32b69024f9ed57f\"},\"model\":\"engineering-no-model\"},\"artifact_digest\":\"sha256:9f50df1213b6646601aa65e1646051bd63a5681638dcf2a5746451d2d925eaf3\"}}");
    static int assertions;
    static void check(boolean value) { assertions++; if (!value) throw new AssertionError("H03 probe assertion " + assertions); }
    static void reject(Runnable body) { assertions++; try { body.run(); } catch (IllegalArgumentException | IllegalStateException expected) { return; } throw new AssertionError("Expected rejection " + assertions); }
    static JsonNode data(String role) { return FIXTURE.path(role); }
    static ObjectNode call(String role) {
        JsonNode f = data(role), i = f.path("input");
        ObjectNode c = H03NodeCodec.JSON.createObjectNode();
        c.put("node_execution_id", H03NodeCodec.text(i,"node_execution_id")); c.put("graph_revision","P1-v1");
        c.put("node_path",role); c.put("attempt",1); c.set("binding", f.path("preimage").path("binding").deepCopy());
        c.put("input_digest",H03NodeCodec.text(f.path("output"),"input_digest")); c.put("input",H03NodeCodec.canonical(i));
        c.put("goal",H03NodeCodec.text(f.path("preimage"),"goal")); c.put("request_id","fixture-request-"+role); c.put("context_id","fixture-context-"+role);
        return c;
    }
    static void vectors() {
        for (String role : List.of("sum","reconcile")) {
            JsonNode f=data(role), c=call(role), contract=f.path("contract");
            JsonNode i=H03NodeCodec.input(c,contract,"R1",role);
            check(H03NodeCodec.sha(H03NodeCodec.canonical(f.path("preimage"))).equals(H03NodeCodec.text(c,"input_digest")));
            check(H03NodeCodec.sha(H03NodeCodec.canonical(f.path("output"))).equals(H03NodeCodec.text(f,"artifact_digest")));
            check(H03NodeCodec.rawHash(contract).equals(H03NodeCodec.text(i,"node_contract_sha256")));
            check(H03NodeCodec.output(H03NodeCodec.canonical(f.path("output")),c,i,contract).equals(H03NodeCodec.canonical(f.path("output"))));
            reject(()->H03NodeCodec.input(c,contract,"OTHER",role));
            ObjectNode wrong=c.deepCopy(); wrong.put("node_execution_id","sha256:"+"0".repeat(64)); reject(()->H03NodeCodec.input(wrong,contract,"R1",role));
            ObjectNode bad=i.deepCopy(); bad.put("context_id","forbidden"); ObjectNode contextInput=c.deepCopy(); contextInput.put("input",H03NodeCodec.canonical(bad));
            reject(()->H03NodeCodec.input(contextInput,contract,"R1",role));
            ObjectNode out=f.path("output").deepCopy(); out.put("input_digest","sha256:"+"0".repeat(64)); reject(()->H03NodeCodec.output(H03NodeCodec.canonical(out),c,i,contract));
            ObjectNode dup=i.deepCopy(); ((com.fasterxml.jackson.databind.node.ArrayNode)dup.path("fields")).add(dup.path("fields").get(0).deepCopy());
            reject(()->H03NodeCodec.fields(dup.path("fields"),contract.path("input_fields")));
        }
        reject(()->H03NodeCodec.parse("{\"x\":1,\"x\":2}")); reject(()->H03NodeCodec.parse("{} {}")); reject(()->H03NodeCodec.parse("{\"x\":1.1}"));
        for (String n : List.of("-0","0.0","01","1.0000001","10000000000","1e2")) reject(()->H03NodeCodec.decimal(n));
        check(H03NodeCodec.decimal("-0.000001").toPlainString().equals("-0.000001")); check(H03NodeCodec.decimal("9999999999.999999").signum()>0);
    }

    static void runtime() throws Exception {
        GoClient client=new GoClient("http://unreachable.invalid",Map.of("H03_RUNTIME_CREDENTIAL","x".repeat(32)));
        for(String path:List.of("/exec/lease","/exec/heartbeat","/exec/call/begin","/exec/call/update","/exec/call/commit","/exec/done"))
            check(client.request(path,Map.of(),1).headers().firstValue("X-Hub-Runtime-Credential").orElse("").equals("x".repeat(32)));
        check(client.request("/agent/context",Map.of(),1).headers().firstValue("X-Hub-Runtime-Credential").isEmpty());
        check(new GoClient("http://unreachable.invalid",Map.of()).request("/exec/lease",Map.of(),1).headers().firstValue("X-Hub-Runtime-Credential").isEmpty());
        reject(()->new GoClient("http://unreachable.invalid",Map.of("H03_RUNTIME_CREDENTIAL","short")));
        Map<String,String> env=Map.of("CREDIT_BUS_GATEWAY","http://unreachable.invalid","CREDIT_BUS_TOKEN","synthetic-fixture-only");
        ObjectNode c=call("sum"); JsonNode contract=data("sum").path("contract");
        int[] sends={0},commits={0}; boolean[] query={false}, timeout={false};
        GoClient ledger=new GoClient("http://unreachable.invalid",Map.of(),(path,body,t)->{
            if(path.equals("/exec/call/begin")) {
                check(body.keySet().equals(Set.of("run_id","token","step")));
                ObjectNode result=H03NodeCodec.JSON.createObjectNode(); result.put("action",query[0]?"query":"send");
                ObjectNode copy=c.deepCopy(); if(query[0])copy.put("remote_task_id","fixture-task"); result.set("call",copy); return result;
            }
            if(path.equals("/exec/call/update"))return H03NodeCodec.JSON.createObjectNode();
            if(path.equals("/exec/call/commit")) { commits[0]++; check(body.get("remote_task_id").equals("fixture-task"));
                return H03NodeCodec.JSON.valueToTree(Map.of("output","[step:sum]\n"+body.get("output"))); }
            throw new AssertionError("Unexpected network route");
        });
        RemoteAgentMember.Gateway gateway=(method,id,params)->{
            if(method.equals("SendMessage")){ sends[0]++; check(!H03NodeCodec.canonical(params).contains("H03_RUNTIME")); }
            else check(method.equals("GetTask"));
            if(timeout[0]) {
                if(method.equals("SendMessage")) return H03NodeCodec.JSON.valueToTree(Map.of("id","fixture-task","contextId","fixture-context-sum","status",Map.of("state","TASK_STATE_WORKING")));
                throw new java.net.http.HttpTimeoutException("synthetic original-task query");
            }
            return H03NodeCodec.JSON.valueToTree(Map.of("id","fixture-task","contextId","fixture-context-sum","status",Map.of("state","TASK_STATE_COMPLETED"),
                "artifacts",List.of(Map.of("parts",List.of(Map.of("kind","text","text",H03NodeCodec.canonical(Map.of("node_execution_id",H03NodeCodec.text(c,"node_execution_id"),"attempt",1,"input_digest",H03NodeCodec.text(c,"input_digest"),"output",H03NodeCodec.canonical(data("sum").path("output"))))))))));
        };
        RemoteAgentMember member=new RemoteAgentMember(ledger,"R1","fixture-lease","sum",env,true,contract,gateway);
        timeout[0]=true; reject(()->member.execute("ignored-model-input","ignored-model-goal")); check(sends[0]==1 && commits[0]==0);
        query[0]=true; timeout[0]=false; check(member.execute("ignored","ignored").contains("hub-node-output/v1")); check(sends[0]==1 && commits[0]==1);
        RemoteAgentMember.Gateway duplicate=(method,id,params)->{ObjectNode n=gateway.call(method,id,params).deepCopy();((com.fasterxml.jackson.databind.node.ArrayNode)n.path("artifacts")).add(n.path("artifacts").get(0).deepCopy());return n;};
        reject(()->new RemoteAgentMember(ledger,"R1","fixture-lease","sum",env,true,contract,duplicate).execute("ignored","ignored"));check(sends[0]==1&&commits[0]==1);
        RemoteAgentMember.Gateway wrongTask=(method,id,params)->{ObjectNode n=gateway.call(method,id,params).deepCopy();n.put("id","foreign-task");return n;};
        reject(()->new RemoteAgentMember(ledger,"R1","fixture-lease","sum",env,true,contract,wrongTask).execute("ignored","ignored"));check(sends[0]==1&&commits[0]==1);
        GoClient legacy=new GoClient("http://unreachable.invalid",Map.of(),(path,body,t)->{ check(body.get("input").equals("legacy input")&&body.get("goal").equals("legacy goal")); return H03NodeCodec.JSON.valueToTree(Map.of("action","reuse","output","legacy output")); });
        check(new RemoteAgentMember(legacy,"old","lease","text",env).execute("legacy input","legacy goal").equals("legacy output"));
        var bp=Blueprints.compile("graph",List.of(Map.of("id","sum","task","sum"),Map.of("id","reconcile","task","reconcile","depends_on",List.of("sum"))));
        check(bp.members().size()==2 && Blueprints.steps(bp).size()==2 && Blueprints.digest(bp).startsWith("sha256:"));
    }
    public static void main(String[] args) throws Exception { vectors(); runtime(); System.out.println("H03_PROBE_PASS assertions="+assertions+" native=NOT_RUN"); }

}