// Package ws finds agentbook workspaces, reads their settings and runs their kit scripts.
package ws

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	agentbook "github.com/EnigmaticZee/agentbook"
)

// ConfKeys are the settings read from agentbook.conf.
var ConfKeys = []string{"PROJECT_NAME", "STAGE", "LAYOUT", "CODE_DIR", "CODE_REMOTE", "HANDBOOK_REMOTE",
	"GIT_NAME", "GIT_EMAIL", "HIDE_AI_TRACES", "AGENT_FILES", "TEST_CMD", "INSTALL_CMD"}

// Workspace is an agentbook workspace on disk.
type Workspace struct {
	Root string
	Conf map[string]string
}

// ErrNoWorkspace means no agentbook.conf was found in the folder or any folder above it.
var ErrNoWorkspace = errors.New("not inside an agentbook workspace")

// Find walks up from dir until it finds a folder holding agentbook.conf and kit/.
func Find(dir string) (*Workspace, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	for {
		if isFile(filepath.Join(dir, "agentbook.conf")) && isDir(filepath.Join(dir, "kit")) {
			w := &Workspace{Root: dir}
			w.Conf, err = ReadConf(filepath.Join(dir, "agentbook.conf"))
			return w, err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNoWorkspace
		}
		dir = parent
	}
}

// ReadConf sources the file with bash (it is shell syntax) and returns the known keys.
func ReadConf(path string) (map[string]string, error) {
	script := `. "$1" >/dev/null 2>&1; shift; for v in "$@"; do printf '%s\0%s\0' "$v" "${!v}"; done`
	args := append([]string{"-c", script, "_", path}, ConfKeys...)
	out, err := exec.Command("bash", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	parts := strings.Split(string(out), "\x00")
	conf := map[string]string{}
	for i := 0; i+1 < len(parts); i += 2 {
		conf[parts[i]] = parts[i+1]
	}
	return conf, nil
}

// Get returns a setting with a default.
func (w *Workspace) Get(key, def string) string {
	if v := w.Conf[key]; v != "" {
		return v
	}
	return def
}

// Name is the project's display name.
func (w *Workspace) Name() string { return w.Get("PROJECT_NAME", filepath.Base(w.Root)) }

// KitVersion is the kit version recorded in the workspace.
func (w *Workspace) KitVersion() string {
	b, err := os.ReadFile(filepath.Join(w.Root, "kit", "VERSION"))
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(b))
}

// Script returns a command for kit/<name>.sh in this workspace.
func (w *Workspace) Script(name string, args ...string) *exec.Cmd {
	cmd := exec.Command("bash", append([]string{filepath.Join(w.Root, "kit", name+".sh")}, args...)...)
	cmd.Dir = w.Root
	return cmd
}

// Stream runs cmd and sends each output line (stdout and stderr together) to lines, then closes it.
func Stream(cmd *exec.Cmd, lines chan<- string) error {
	defer close(lines)
	r, wr := io.Pipe()
	cmd.Stdout, cmd.Stderr = wr, wr
	cmd.Stdin = nil
	if err := cmd.Start(); err != nil {
		wr.Close()
		return err
	}
	done := make(chan struct{})
	go func() {
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(done)
	}()
	err := cmd.Wait()
	wr.Close()
	<-done
	return err
}

// ExtractKit writes the embedded kit into a new temporary folder and returns its path.
func ExtractKit() (string, error) {
	dir, err := os.MkdirTemp("", "agentbook-kit-")
	if err != nil {
		return "", err
	}
	err = fs.WalkDir(agentbook.Files, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, p)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		b, err := agentbook.Files.ReadFile(p)
		if err != nil {
			return err
		}
		mode := fs.FileMode(0o644)
		if strings.HasSuffix(p, ".sh") || strings.HasPrefix(p, "kit/hooks/") {
			mode = 0o755
		}
		return os.WriteFile(dst, b, mode)
	})
	if err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

// Version is the kit version bundled in this binary.
func Version() string {
	b, _ := agentbook.Files.ReadFile("VERSION")
	return strings.TrimSpace(string(b))
}

// Prompt is the bundled PROMPT.md.
func Prompt() string {
	b, _ := agentbook.Files.ReadFile("PROMPT.md")
	return string(b)
}

// ---- Reading the handbook ----

// Section returns the body of a "## Title" section in a markdown file (without the heading).
func Section(md, title string) string {
	var out []string
	in := false
	for _, l := range strings.Split(md, "\n") {
		if strings.HasPrefix(l, "## ") {
			if in {
				break
			}
			in = strings.EqualFold(strings.TrimSpace(l[3:]), title)
			continue
		}
		if in {
			out = append(out, l)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// Read returns a workspace file's text, or "" when missing.
func (w *Workspace) Read(rel string) string {
	b, _ := os.ReadFile(filepath.Join(w.Root, rel))
	return string(b)
}

var lastUpdated = regexp.MustCompile(`Last updated: ([0-9-]+)`)

// LastUpdated is STATUS.md's "Last updated" date.
func (w *Workspace) LastUpdated() string {
	if m := lastUpdated.FindStringSubmatch(w.Read("STATUS.md")); m != nil {
		return m[1]
	}
	return ""
}

// Decisions returns DECISIONS.md entry headings ("YYYY-MM-DD: title"), oldest first.
func (w *Workspace) Decisions() []string {
	var out []string
	for _, l := range strings.Split(w.Read("DECISIONS.md"), "\n") {
		if strings.HasPrefix(l, "## ") && len(l) > 13 && l[3] >= '0' && l[3] <= '9' {
			out = append(out, strings.TrimSpace(l[3:]))
		}
	}
	return out
}

// Sessions returns session log file names (YYYY-MM-DD.md), oldest first.
func (w *Workspace) Sessions() []string {
	entries, _ := os.ReadDir(filepath.Join(w.Root, "history", "sessions"))
	var out []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// Questions counts "### Q:" entries across all session logs.
func (w *Workspace) Questions() int {
	n := 0
	for _, s := range w.Sessions() {
		n += bytes.Count([]byte(w.Read(filepath.Join("history", "sessions", s))), []byte("### Q:"))
	}
	return n
}

// CodeCommits counts entries in history/commits.md.
func (w *Workspace) CodeCommits() int {
	n := 0
	for _, l := range strings.Split(w.Read(filepath.Join("history", "commits.md")), "\n") {
		if strings.HasPrefix(l, "- ") {
			n++
		}
	}
	return n
}

func isFile(p string) bool { st, err := os.Stat(p); return err == nil && !st.IsDir() }
func isDir(p string) bool  { st, err := os.Stat(p); return err == nil && st.IsDir() }
