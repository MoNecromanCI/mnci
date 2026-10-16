package mcpserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func call(t *testing.T, s Server, msg string) map[string]any {
	t.Helper()
	raw := s.Handle([]byte(msg))
	if raw == nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestServer(t *testing.T) {
	var ran []string
	s := Server{Tools: Tools, Version: "1.2.3", Run: func(args []string) Outcome {
		ran = append(ran, strings.Join(args, " "))
		return Outcome{Text: "ok"}
	}}
	init := call(t, s, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`)
	result := init["result"].(map[string]any)
	if result["protocolVersion"] != "2025-03-26" || result["serverInfo"].(map[string]any)["version"] != "1.2.3" {
		t.Fatalf("initialize %v", init)
	}
	if call(t, s, `{"jsonrpc":"2.0","method":"notifications/initialized"}`) != nil {
		t.Fatal("a notification has no answer")
	}
	listed := call(t, s, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)["result"].(map[string]any)["tools"].([]any)
	if len(listed) != len(Tools) {
		t.Fatalf("%d tools listed", len(listed))
	}
	for _, c := range []struct{ msg, want string }{
		{`{"name":"smith_projects"}`, "projects --json"},
		{`{"name":"smith_install","arguments":{"package":"zod@^3","projects":["web","apps/api"],"dev":true}}`, "install -w web -w apps/api -D zod@^3"},
		{`{"name":"smith_run","arguments":{"tasks":["test","lint"],"projects":["web"],"affected":true}}`, "run test lint -p web --affected"},
		{`{"name":"smith_add_project","arguments":{"kind":"app","name":"api","lang":"go"}}`, "add app api --lang go"},
		{`{"name":"smith_remove_project","arguments":{"name":"old","dry_run":true}}`, "remove old --dry-run"},
		{`{"name":"smith_pipeline","arguments":{"template":"deploy-pages","project":"web"}}`, "pipeline deploy-pages --project web"},
		{`{"name":"smith_graph","arguments":{"focus":["web","api"]}}`, "graph --format json --focus web,api"},
	} {
		ran = nil
		res := call(t, s, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":`+c.msg+`}`)["result"].(map[string]any)
		if res["isError"] != false || len(ran) != 1 || ran[0] != c.want {
			t.Errorf("%s: ran %v, result %v", c.msg, ran, res)
		}
	}
	ran = nil
	res := call(t, s, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"smith_install","arguments":{"package":"--registry=evil","projects":["web"]}}}`)["result"].(map[string]any)
	if res["isError"] != true || len(ran) != 0 {
		t.Fatalf("an option smuggled in as a value must be refused: %v %v", res, ran)
	}
	if e := call(t, s, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"nope"}}`)["error"]; e == nil {
		t.Fatal("unknown tool")
	}
	if e := call(t, s, `{"jsonrpc":"2.0","id":6,"method":"resources/list"}`)["error"].(map[string]any); e["code"] != float64(-32601) {
		t.Fatalf("unknown method %v", e)
	}
}
