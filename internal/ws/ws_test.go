package ws

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSection(t *testing.T) {
	md := "# T\n\n## Next step\nShip it.\n\nThen rest.\n\n## Waiting\nNothing.\n"
	if got := Section(md, "next step"); got != "Ship it.\n\nThen rest." {
		t.Fatalf("got %q", got)
	}
	if got := Section(md, "missing"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestFindAndConf(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "kit"), 0o755)
	os.MkdirAll(filepath.Join(root, "code", "src"), 0o755)
	conf := "PROJECT_NAME=\"My & App\"\nTEST_CMD='go test ./... | tee \"out\"'\nLAYOUT=\"single\"\n"
	os.WriteFile(filepath.Join(root, "agentbook.conf"), []byte(conf), 0o644)
	w, err := Find(filepath.Join(root, "code", "src"))
	if err != nil {
		t.Fatal(err)
	}
	if w.Name() != "My & App" || w.Get("TEST_CMD", "") != `go test ./... | tee "out"` || w.Get("STAGE", "local") != "local" {
		t.Fatalf("conf read wrong: %+v", w.Conf)
	}
	if _, err := Find(t.TempDir()); err != ErrNoWorkspace {
		t.Fatalf("want ErrNoWorkspace, got %v", err)
	}
}

func TestEmbeddedKit(t *testing.T) {
	dir, err := ExtractKit()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	for _, f := range []string{"install.sh", "kit/sync.sh", "kit/hooks/pre-push", "templates/AGENTS.md"} {
		st, err := os.Stat(filepath.Join(dir, f))
		if err != nil {
			t.Fatalf("%s missing: %v", f, err)
		}
		if (f == "install.sh" || f == "kit/sync.sh" || f == "kit/hooks/pre-push") && st.Mode()&0o100 == 0 {
			t.Fatalf("%s isn't executable", f)
		}
	}
	if Version() == "" {
		t.Fatal("no version")
	}
}
