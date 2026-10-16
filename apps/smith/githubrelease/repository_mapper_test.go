package githubrelease

import "testing"

func TestParseRepository(t *testing.T) {
	for _, in := range []string{
		"https://github.com/MoNecromanCI/mnci.git",
		"https://github.com/MoNecromanCI/mnci",
		"https://x-access-token:abc@github.com/MoNecromanCI/mnci.git",
		"git@github.com:MoNecromanCI/mnci.git",
		"ssh://git@github.com/MoNecromanCI/mnci.git",
		"MoNecromanCI/mnci",
	} {
		r, err := ParseRepository(in)
		if err != nil || r.String() != "MoNecromanCI/mnci" {
			t.Errorf("%s: %v %v", in, r, err)
		}
	}
	for _, in := range []string{"", "https://github.com/onlyowner", "C:/local/bare.git/x/y"} {
		if _, err := ParseRepository(in); err == nil {
			t.Errorf("%q must be rejected", in)
		}
	}
}
