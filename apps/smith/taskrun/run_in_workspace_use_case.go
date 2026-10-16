package taskrun

import (
	"context"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/affected"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

// Request is what to run in an open workspace.
type Request struct {
	Tasks []string
	// Projects are selection patterns; empty means every project.
	Projects []string
	// Affected narrows the selection to projects affected since Base.
	Affected bool
	Base     string
	Parallel int
	Bail     bool
	NoCache  bool
	Out      io.Writer
}

// RunInWorkspace selects projects, builds the task graph and runs it.
func RunInWorkspace(ctx context.Context, s *session.Session, req Request) (Summary, error) {
	names, err := s.Graph.Select(req.Projects)
	if err != nil {
		return Summary{}, err
	}
	if req.Affected {
		res := affected.Compute(s.Git, s.Graph, req.Base)
		fmt.Fprintf(req.Out, "affected: %s → %d project(s)\n", res.Reason, len(res.Projects))
		keep := map[string]bool{}
		for _, n := range res.Projects {
			keep[n] = true
		}
		var narrowed []string
		for _, n := range names {
			if keep[n] {
				narrowed = append(narrowed, n)
			}
		}
		names = narrowed
	}
	if len(names) == 0 {
		fmt.Fprintln(req.Out, "nothing to run")
		return Summary{}, nil
	}
	var withTask []string
	for _, n := range names {
		for _, t := range req.Tasks {
			if _, ok := s.Graph.Get(n).Tasks[t]; ok {
				withTask = append(withTask, n)
				break
			}
		}
	}
	if len(withTask) == 0 {
		fmt.Fprintf(req.Out, "no selected project has a %v task\n", req.Tasks)
		return Summary{}, nil
	}
	g, err := taskgraph.Build(s.Graph, withTask, req.Tasks)
	if err != nil {
		return Summary{}, err
	}
	store, err := s.Storage()
	if err != nil {
		return Summary{}, err
	}
	return Run(ctx, g, Options{
		Root: s.Root, Files: s.Files, Projects: s.Graph, Storage: store,
		Parallel: req.Parallel, Bail: req.Bail, NoCache: req.NoCache, Out: req.Out, ToolVersion: s.Version,
	}), nil
}
