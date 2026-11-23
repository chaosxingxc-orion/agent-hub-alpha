package com.openjiuwen.demo.professional;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import java.util.*;
/** Frozen accepted fixture vectors; no Main, Worker service or real network. */
public final class H03EngineeringProbe {

    static final JsonNode FIXTURE = H03BoundNodeCodec.parse("{\"sum\":{\"contract\":{\"schema_version\":\"hub-node-contract/v1\",\"input_fields\":[{\"name\":\"component_a\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"},{\"name\":\"component_b\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"},{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"}],\"output_fields\":[{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"},{\"name\":\"total\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"}]},\"input\":{\"schema_version\":\"hub-bound-input/v1\",\"execution_binding_sha256\":\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\",\"node_contract_sha256\":\"sha256:1529f647e38c976923e21f0362cb94a0796b543318ad64e131940349250219e6\",\"plan_hash\":\"sha256:ad6922eb0624c6a6e3e13fe80c7b7fc8ffff24d67f7579ce60c39df78f5852ec\",\"run_id\":\"R1\",\"node_execution_id\":\"sha256:592757eb3061ad13c9b3515dbe55103f9ee2b5cb747a9d0fe4cd263e5a677675\",\"node_path\":\"sum\",\"attempt\":1,\"fields\":[{\"name\":\"component_a\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"40\"},{\"name\":\"component_b\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"60\"},{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"95\"}],\"dependencies\":[{\"kind\":\"material\",\"id\":\"engineering-material-a\",\"sha256\":\"sha256:d24deb3bc00530e2d9530f27100eff56b725b1ca7d6e76250d9199cea3532060\"}]},\"output\":{\"schema_version\":\"hub-node-output/v1\",\"execution_binding_sha256\":\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\",\"node_contract_sha256\":\"sha256:1529f647e38c976923e21f0362cb94a0796b543318ad64e131940349250219e6\",\"input_digest\":\"sha256:5a44234942a2cdeeaf5e94f4be13b71801d393c8f4855a5be86ae874df673221\",\"fields\":[{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"95\"},{\"name\":\"total\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"100\"}]},\"preimage\":{\"input\":\"{\\\"attempt\\\":1,\\\"dependencies\\\":[{\\\"id\\\":\\\"engineering-material-a\\\",\\\"kind\\\":\\\"material\\\",\\\"sha256\\\":\\\"sha256:d24deb3bc00530e2d9530f27100eff56b725b1ca7d6e76250d9199cea3532060\\\"}],\\\"execution_binding_sha256\\\":\\\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\\\",\\\"fields\\\":[{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"component_a\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"40\\\"},{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"component_b\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"60\\\"},{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"reference\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"95\\\"}],\\\"node_contract_sha256\\\":\\\"sha256:1529f647e38c976923e21f0362cb94a0796b543318ad64e131940349250219e6\\\",\\\"node_execution_id\\\":\\\"sha256:592757eb3061ad13c9b3515dbe55103f9ee2b5cb747a9d0fe4cd263e5a677675\\\",\\\"node_path\\\":\\\"sum\\\",\\\"plan_hash\\\":\\\"sha256:ad6922eb0624c6a6e3e13fe80c7b7fc8ffff24d67f7579ce60c39df78f5852ec\\\",\\\"run_id\\\":\\\"R1\\\",\\\"schema_version\\\":\\\"hub-bound-input/v1\\\"}\",\"goal\":\"engineering arithmetic only。范围：engineering-entity，consolidated口径，FY2025 年；任务重点：reconciliation\",\"binding\":{\"agent_id\":\"h03-engineering-sum\",\"version\":\"0.1.0\",\"card_digest\":\"sha256:8a8a431f00ff8f9b54afb8dd4d2b864255dfea9852941f539bd7cc2fe1b67ea1\",\"contract_digest\":\"sha256:30e3204060ed57bd6583c057838f540ad2ec0474974d1890359844e927f2a7ec\"},\"model\":\"engineering-no-model\"},\"artifact_digest\":\"sha256:5033b744de6cd2f37388620f8d1aa48e30508d4dd8bbdb8dbf8dcb8f360e4173\"},\"reconcile\":{\"contract\":{\"schema_version\":\"hub-node-contract/v1\",\"input_fields\":[{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"},{\"name\":\"total\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"}],\"output_fields\":[{\"name\":\"difference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\"}]},\"input\":{\"schema_version\":\"hub-bound-input/v1\",\"execution_binding_sha256\":\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\",\"node_contract_sha256\":\"sha256:ed98d4b8c49fb6ad8784664c461752ae70e2ce17f6ac3af120679c8a10ab283d\",\"plan_hash\":\"sha256:ad6922eb0624c6a6e3e13fe80c7b7fc8ffff24d67f7579ce60c39df78f5852ec\",\"run_id\":\"R1\",\"node_execution_id\":\"sha256:3700f123898c4e63a719ce0fa125e0bc20aff132bff69b5abd20af483a4ffc88\",\"node_path\":\"reconcile\",\"attempt\":1,\"fields\":[{\"name\":\"reference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"95\"},{\"name\":\"total\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"100\"}],\"dependencies\":[{\"kind\":\"step\",\"id\":\"sum\",\"node_execution_id\":\"sha256:592757eb3061ad13c9b3515dbe55103f9ee2b5cb747a9d0fe4cd263e5a677675\",\"artifact_digest\":\"sha256:5033b744de6cd2f37388620f8d1aa48e30508d4dd8bbdb8dbf8dcb8f360e4173\"}]},\"output\":{\"schema_version\":\"hub-node-output/v1\",\"execution_binding_sha256\":\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\",\"node_contract_sha256\":\"sha256:ed98d4b8c49fb6ad8784664c461752ae70e2ce17f6ac3af120679c8a10ab283d\",\"input_digest\":\"sha256:4fa1e38110ce8902cf6921f9aca98c2cd6dca79c8ca1011a58685171df499eff\",\"fields\":[{\"name\":\"difference\",\"type\":\"decimal\",\"unit\":\"USD-million\",\"period\":\"FY2025\",\"basis\":\"consolidated\",\"value\":\"5\"}]},\"preimage\":{\"input\":\"{\\\"attempt\\\":1,\\\"dependencies\\\":[{\\\"artifact_digest\\\":\\\"sha256:5033b744de6cd2f37388620f8d1aa48e30508d4dd8bbdb8dbf8dcb8f360e4173\\\",\\\"id\\\":\\\"sum\\\",\\\"kind\\\":\\\"step\\\",\\\"node_execution_id\\\":\\\"sha256:592757eb3061ad13c9b3515dbe55103f9ee2b5cb747a9d0fe4cd263e5a677675\\\"}],\\\"execution_binding_sha256\\\":\\\"sha256:82bdf449d7b225e3faedc45b38e062fe31950af7d7a03ea535d48dacccae09ee\\\",\\\"fields\\\":[{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"reference\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"95\\\"},{\\\"basis\\\":\\\"consolidated\\\",\\\"name\\\":\\\"total\\\",\\\"period\\\":\\\"FY2025\\\",\\\"type\\\":\\\"decimal\\\",\\\"unit\\\":\\\"USD-million\\\",\\\"value\\\":\\\"100\\\"}],\\\"node_contract_sha256\\\":\\\"sha256:ed98d4b8c49fb6ad8784664c461752ae70e2ce17f6ac3af120679c8a10ab283d\\\",\\\"node_execution_id\\\":\\\"sha256:3700f123898c4e63a719ce0fa125e0bc20aff132bff69b5abd20af483a4ffc88\\\",\\\"node_path\\\":\\\"reconcile\\\",\\\"plan_hash\\\":\\\"sha256:ad6922eb0624c6a6e3e13fe80c7b7fc8ffff24d67f7579ce60c39df78f5852ec\\\",\\\"run_id\\\":\\\"R1\\\",\\\"schema_version\\\":\\\"hub-bound-input/v1\\\"}\",\"goal\":\"engineering arithmetic only。范围：engineering-entity，consolidated口径，FY2025 年；任务重点：reconciliation\",\"binding\":{\"agent_id\":\"h03-engineering-reconcile\",\"version\":\"0.1.0\",\"card_digest\":\"sha256:5c8c2312be1a661f97cbfcf71dd778aa6986d753e5aa3e1ad77333671b350be3\",\"contract_digest\":\"sha256:9b998fedda163fe87f94b11491f54d75219e1426cc93ba16e32b69024f9ed57f\"},\"model\":\"engineering-no-model\"},\"artifact_digest\":\"sha256:9f50df1213b6646601aa65e1646051bd63a5681638dcf2a5746451d2d925eaf3\"}}");
    static int assertions;
    static void check(boolean value) { assertions++; if (!value) throw new AssertionError("H03 probe assertion " + assertions); }
    static void reject(Runnable body) { assertions++; try { body.run(); } catch (IllegalArgumentException | IllegalStateException expected) { return; } throw new AssertionError("Expected rejection " + assertions); }
    static JsonNode data(String role) { return FIXTURE.path(role); }
    static ObjectNode call(String role) {
        JsonNode f = data(role), i = f.path("input");
        ObjectNode c = H03BoundNodeCodec.JSON.createObjectNode();
        c.put("node_execution_id", H03BoundNodeCodec.text(i,"node_execution_id")); c.put("graph_revision","P1-v1");
        c.put("node_path",role); c.put("attempt",1); c.set("binding", f.path("preimage").path("binding").deepCopy());
        c.put("input_digest",H03BoundNodeCodec.text(f.path("output"),"input_digest")); c.put("input",H03BoundNodeCodec.canonical(i));
        c.put("goal",H03BoundNodeCodec.text(f.path("preimage"),"goal")); c.put("request_id","fixture-request-"+role); c.put("context_id","fixture-context-"+role);
        return c;
    }
    static void vectors() {
        for (String role : List.of("sum","reconcile")) {
            JsonNode f=data(role), c=call(role), contract=f.path("contract");
            JsonNode i=H03BoundNodeCodec.input(c,contract,"R1",role);
            check(H03BoundNodeCodec.sha(H03BoundNodeCodec.canonical(f.path("preimage"))).equals(H03BoundNodeCodec.text(c,"input_digest")));
            check(H03BoundNodeCodec.sha(H03BoundNodeCodec.canonical(f.path("output"))).equals(H03BoundNodeCodec.text(f,"artifact_digest")));
            check(H03BoundNodeCodec.rawHash(contract).equals(H03BoundNodeCodec.text(i,"node_contract_sha256")));
            check(H03BoundNodeCodec.output(H03BoundNodeCodec.canonical(f.path("output")),c,i,contract).equals(H03BoundNodeCodec.canonical(f.path("output"))));
            reject(()->H03BoundNodeCodec.input(c,contract,"OTHER",role));
            ObjectNode wrong=c.deepCopy(); wrong.put("node_execution_id","sha256:"+"0".repeat(64)); reject(()->H03BoundNodeCodec.input(wrong,contract,"R1",role));
            ObjectNode bad=i.deepCopy(); bad.put("context_id","forbidden"); ObjectNode contextInput=c.deepCopy(); contextInput.put("input",H03BoundNodeCodec.canonical(bad));
            reject(()->H03BoundNodeCodec.input(contextInput,contract,"R1",role));
            ObjectNode out=f.path("output").deepCopy(); out.put("input_digest","sha256:"+"0".repeat(64)); reject(()->H03BoundNodeCodec.output(H03BoundNodeCodec.canonical(out),c,i,contract));
            ObjectNode dup=i.deepCopy(); ((com.fasterxml.jackson.databind.node.ArrayNode)dup.path("fields")).add(dup.path("fields").get(0).deepCopy());
            reject(()->H03BoundNodeCodec.fields(dup.path("fields"),contract.path("input_fields")));
        }
        reject(()->H03BoundNodeCodec.parse("{\"x\":1,\"x\":2}")); reject(()->H03BoundNodeCodec.parse("{} {}")); reject(()->H03BoundNodeCodec.parse("{\"x\":1.1}"));
        for (String n : List.of("-0","0.0","01","1.0000001","10000000000","1e2")) reject(()->H03BoundNodeCodec.decimal(n));
        check(H03BoundNodeCodec.decimal("-0.000001").toPlainString().equals("-0.000001")); check(H03BoundNodeCodec.decimal("9999999999.999999").signum()>0);
    }

    static ObjectNode safe(String role,ObjectNode call) {
        ObjectNode n=H03BoundNodeCodec.JSON.createObjectNode(); n.put("expert","eng."+role); n.set("tools",H03BoundNodeCodec.JSON.createArrayNode());
        n.set("node_contract",data(role).path("contract").deepCopy());n.set("call",call);return n;
    }
    static Map<String,String> env(String role) { return Map.of("H03_ENGINEERING_ROLE",role,"AGENT_SERVICE_ID","h03-engineering-"+role,"OPENJIUWEN_SERVICE_VERSION","0.1.0","CREDIT_URL","http://unreachable.invalid"); }
    static String execute(String role,ObjectNode c) {
        H03EngineeringHandler h=new H03EngineeringHandler(env(role),auth->{check(auth.keySet().equals(Set.of("run_id","step","token","service_id","service_version","context_id"))); return safe(role,c);});
        return h.perform(H03BoundNodeCodec.canonical(Map.of("run_id","R1","step",role,"token","synthetic-scoped-token")),"fixture-context-"+role);
    }
    static String value(String artifact,String field) {
        JsonNode out=H03BoundNodeCodec.parse(H03BoundNodeCodec.text(H03BoundNodeCodec.parse(artifact),"output"));
        return H03BoundNodeCodec.text(H03BoundNodeCodec.specs(out.path("fields"),true).get(field),"value");
    }
    static ObjectNode withValue(String role,String name,String value) {
        ObjectNode c=call(role),i=H03BoundNodeCodec.parse(H03BoundNodeCodec.text(c,"input")).deepCopy();
        for(JsonNode f:i.path("fields"))if(H03BoundNodeCodec.text(f,"name").equals(name))((ObjectNode)f).put("value",value);
        c.put("input",H03BoundNodeCodec.canonical(i));ObjectNode p=data(role).path("preimage").deepCopy();p.put("input",H03BoundNodeCodec.canonical(i));c.put("input_digest",H03BoundNodeCodec.sha(H03BoundNodeCodec.canonical(p)));return c;
    }
    static void arithmetic() {
        check(value(execute("sum",call("sum")),"total").equals("100")); check(value(execute("reconcile",call("reconcile")),"difference").equals("5"));
        check(value(execute("sum",withValue("sum","component_a","-40")),"total").equals("20"));
        check(value(execute("sum",withValue("sum","component_a","0.000001")),"total").equals("60.000001"));
        check(value(execute("reconcile",withValue("reconcile","total","94")),"difference").equals("-1"));
        ObjectNode c=call("sum"),input=H03BoundNodeCodec.parse(H03BoundNodeCodec.text(c,"input")).deepCopy();
        ((ObjectNode)input.path("fields").get(0)).put("value","41"); c.put("input",H03BoundNodeCodec.canonical(input));
        ObjectNode pre=data("sum").path("preimage").deepCopy();pre.put("input",H03BoundNodeCodec.canonical(input)); c.put("input_digest",H03BoundNodeCodec.sha(H03BoundNodeCodec.canonical(pre)));
        String changed=execute("sum",c);check(value(changed,"total").equals("101"));
        ObjectNode next=call("reconcile"),ni=H03BoundNodeCodec.parse(H03BoundNodeCodec.text(next,"input")).deepCopy();
        ((ObjectNode)ni.path("fields").get(1)).put("value","101");
        ((ObjectNode)ni.path("dependencies").get(0)).put("artifact_digest",H03BoundNodeCodec.sha(H03BoundNodeCodec.text(H03BoundNodeCodec.parse(changed),"output")));
        next.put("input",H03BoundNodeCodec.canonical(ni)); ObjectNode np=data("reconcile").path("preimage").deepCopy();np.put("input",H03BoundNodeCodec.canonical(ni));next.put("input_digest",H03BoundNodeCodec.sha(H03BoundNodeCodec.canonical(np)));
        check(value(execute("reconcile",next),"difference").equals("6"));
        ObjectNode wrong=call("sum");wrong.put("context_id","different");reject(()->execute("sum",wrong));
        for(String key:List.of("unit","period","basis","type")) {
            ObjectNode bad=call("sum"),i=H03BoundNodeCodec.parse(H03BoundNodeCodec.text(bad,"input")).deepCopy();
            ((ObjectNode)i.path("fields").get(0)).put(key,"wrong");bad.put("input",H03BoundNodeCodec.canonical(i));reject(()->execute("sum",bad));
        }
        ObjectNode overflow=call("sum"),oi=H03BoundNodeCodec.parse(H03BoundNodeCodec.text(overflow,"input")).deepCopy();
        ((ObjectNode)oi.path("fields").get(0)).put("value","9999999999");overflow.put("input",H03BoundNodeCodec.canonical(oi));reject(()->execute("sum",overflow));
        ObjectNode missing=call("sum"),mi=H03BoundNodeCodec.parse(H03BoundNodeCodec.text(missing,"input")).deepCopy();((com.fasterxml.jackson.databind.node.ArrayNode)mi.path("fields")).remove(0);missing.put("input",H03BoundNodeCodec.canonical(mi));reject(()->execute("sum",missing));
        ObjectNode roleMismatch=call("reconcile");reject(()->execute("sum",roleMismatch));
        Map<String,String> unsafe=new HashMap<>(env("sum"));unsafe.put("H03_RUNTIME_CREDENTIAL","fixture");reject(()->new H03EngineeringHandler(unsafe));
        unsafe.remove("H03_RUNTIME_CREDENTIAL");unsafe.put("SWARM_API_KEY","");reject(()->new H03EngineeringHandler(unsafe));
    }
    public static void main(String[] args) { vectors(); arithmetic();System.out.println("H03_PROBE_PASS assertions="+assertions+" model_instances=0 native=NOT_RUN"); }

}