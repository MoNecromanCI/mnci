package tui

import (
	"io"

	"github.com/MoNecromanCI/mnci/apps/mnci/doctor"
	"github.com/MoNecromanCI/mnci/apps/mnci/release"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
)

// Sources load what the shell's views show; tests replace them.
type Sources struct {
	Doctor  func() (doctor.Report, error)
	Release func() releaseLoaded
}

// WorkspaceSources open the workspace in dir for each view, as the CLI
// commands do; plugin output is dropped so it cannot tear the screen.
func WorkspaceSources(dir, version string) Sources {
	open := func() (*session.Session, error) {
		return session.Open(session.Options{Dir: dir, Version: version, Log: io.Discard})
	}
	return Sources{
		Doctor: func() (doctor.Report, error) {
			s, err := open()
			if err != nil {
				return doctor.Report{}, err
			}
			defer s.Close()
			return doctor.Run(s), nil
		},
		Release: func() releaseLoaded {
			s, err := open()
			if err != nil {
				return releaseLoaded{err: err}
			}
			defer s.Close()
			var out releaseLoaded
			for _, f := range release.CheckLineage(s, false) {
				if f.Blocking {
					out.blockers = append(out.blockers, f.Message)
				} else {
					out.warnings = append(out.warnings, f.Message)
				}
			}
			out.plan, out.err = release.MakePlan(s.Git, s.Graph)
			return out
		},
	}
}
