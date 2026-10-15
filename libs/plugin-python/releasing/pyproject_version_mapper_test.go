package releasing

import "testing"

func TestSetVersion(t *testing.T) {
	src := "[tool.x]\r\nversion = \"9\"\r\n\r\n[project]\r\nname = \"a\"\r\nversion   =  '0.0.0'  # managed by mnci\r\n\r\n[tool.y]\r\nversion = \"8\"\r\n"
	got, err := SetVersion(src, "1.4.0")
	if err != nil {
		t.Fatal(err)
	}
	want := "[tool.x]\r\nversion = \"9\"\r\n\r\n[project]\r\nname = \"a\"\r\nversion   =  '1.4.0'  # managed by mnci\r\n\r\n[tool.y]\r\nversion = \"8\"\r\n"
	if got != want {
		t.Fatalf("got %q", got)
	}
	if _, err := SetVersion("[tool.x]\nversion = \"1\"\n", "2"); err == nil {
		t.Fatal("version outside [project] must not count")
	}
}
