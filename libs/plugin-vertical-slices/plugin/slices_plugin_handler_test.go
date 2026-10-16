package plugin

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func TestContribute(t *testing.T) {
	v := New("1.0.0")
	if c := serve.CapabilitiesOf(v); !c.Contribute || c.Discover {
		t.Fatalf("capabilities %+v", c)
	}
	res, err := v.Contribute(protocol.ContributeParams{Projects: []protocol.Project{
		{Name: "tool", Language: "go", Tasks: map[string]protocol.Task{"lint": {}}},
		{Name: "kit", Language: "python"},
		{Name: "web", Language: "node"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	tool := res.Projects["tool"]
	if tool.Tasks[TaskName].Command[0] != "{smith}" || tool.Tasks[TaskName].Inputs[0] != "**/*.go" || tool.DependsOn["lint"][0] != TaskName {
		t.Fatalf("tool %+v", tool)
	}
	if kit := res.Projects["kit"]; kit.Tasks[TaskName].Inputs[0] != "**/*.py" || kit.DependsOn != nil {
		t.Fatalf("kit %+v", kit)
	}
	if _, ok := res.Projects["web"]; ok {
		t.Fatal("node projects are linted by @mnci/eslint-config")
	}
}
