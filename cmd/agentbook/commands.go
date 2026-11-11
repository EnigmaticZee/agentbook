package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/EnigmaticZee/agentbook/internal/ui"
	"github.com/EnigmaticZee/agentbook/internal/ws"
)

// ---- Home ----

func cmdHome() int {
	if !ui.IsTTY() {
		return cmdHelp(nil)
	}
	w, err := ws.Find(".")
	var items []ui.MenuItem
	card := ""
	if err == nil {
		card = overviewCard(w, true)
		items = []ui.MenuItem{
			mi("status", "◎", "Status", "Where the project stands and what's next"),
			mi("find", "⌕", "Find", "Search every decision, note and session log"),
			mi("check", "✓", "Check", "Is everything saved, pushed, logged and tested?"),
			mi("prompt", "❝", "Setup prompt", "Copy the prompt that lets any agent set things up"),
			mi("doctor", "✚", "Doctor", "Look for problems on this computer"),
			mi("help", "?", "Help", "Every command"),
		}
	} else {
		body := ui.Bold.Render("No agentbook here yet.") + "\n" +
			ui.Dim.Render("Create a workspace for a new or existing project, or open one you already have.")
		card = ui.Card("", body, ui.Line, ui.Width())
		items = []ui.MenuItem{
			mi("init", "✦", "New workspace", "Start a project, or adopt the one in this folder"),
			mi("prompt", "❝", "Setup prompt", "Copy the prompt that lets any agent set it up"),
			mi("doctor", "✚", "Doctor", "Check git, bash and GitHub CLI on this computer"),
			mi("help", "?", "Help", "Every command"),
		}
	}
	choice, err := ui.Home(header(w), card, items)
	if err != nil || choice == "" {
		return 0
	}
	if choice == "prompt" {
		return cmdPrompt([]string{"--copy"})
	}
	for _, c := range commands {
		if c.name == choice {
			return c.run(nil)
		}
	}
	return 0
}

// ---- Status ----

func relDate(d string) string {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return d
	}
	days := int(time.Since(t).Hours() / 24)
	switch {
	case d == time.Now().Format("2006-01-02"):
		return "today"
	case days <= 1:
		return "yesterday"
	case days < 60:
		return fmt.Sprintf("%d days ago", days)
	}
	return d
}

func firstPara(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "\n\n"); i >= 0 {
		s = s[:i]
	}
	return strings.Join(strings.Fields(s), " ")
}

func overviewCard(w *ws.Workspace, compact bool) string {
	width := ui.Width()
	stage, layout := w.Get("STAGE", "local"), w.Get("LAYOUT", "single")
	stageColor := map[string]lipgloss.TerminalColor{"local": ui.Muted, "private": ui.Brand, "public": ui.Green}[stage]
	if stageColor == nil {
		stageColor = ui.Muted
	}
	updated := w.LastUpdated()
	upd := ui.Dim.Render("updated " + relDate(updated))
	if updated != time.Now().Format("2006-01-02") {
		upd = ui.Warm.Render("updated " + relDate(updated))
	}
	pills := ui.Pill(stage, stageColor) + " " + ui.Pill(layout+" layout", ui.Muted) + "  " + upd

	status := w.Read("STATUS.md")
	next := firstPara(ws.Section(status, "Next step"))
	if next == "" || strings.HasPrefix(next, "The exact next thing") {
		next = "Not written yet. Ask your agent to fill in STATUS.md."
	}
	dec, ses := w.Decisions(), w.Sessions()
	stat := func(n int, what string) string {
		if n == 1 {
			what = strings.TrimSuffix(what, "s")
		}
		return ui.Bold.Render(fmt.Sprint(n)) + " " + ui.Dim.Render(what)
	}
	stats := strings.Join([]string{stat(len(dec), "decisions"), stat(len(ses), "sessions"),
		stat(w.Questions(), "questions"), stat(w.CodeCommits(), "code commits")}, ui.Fainter.Render("  ·  "))

	var b strings.Builder
	b.WriteString(pills + "\n\n")
	b.WriteString(ui.Warm.Render("NEXT STEP") + "\n" + ui.Text.Render(next) + "\n\n")
	b.WriteString(stats)
	if !compact {
		if len(dec) > 0 {
			b.WriteString("\n\n" + ui.Warm.Render("LATEST DECISIONS"))
			for i := len(dec) - 1; i >= 0 && i >= len(dec)-4; i-- {
				parts := strings.SplitN(dec[i], ": ", 2)
				line := ui.Fainter.Render(parts[0])
				if len(parts) == 2 {
					line += "  " + ui.Text.Render(ui.Truncate(parts[1], width-26))
				}
				b.WriteString("\n" + line)
			}
		}
		if len(ses) > 0 {
			last := ses[len(ses)-1]
			sum := ws.Section(w.Read(filepath.Join("history", "sessions", last)), "Summary")
			b.WriteString("\n\n" + ui.Warm.Render("LAST SESSION") + "  " + ui.Fainter.Render(strings.TrimSuffix(last, ".md")))
			n := 0
			for _, l := range strings.Split(sum, "\n") {
				l = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "- "))
				if l == "" || n >= 4 {
					continue
				}
				b.WriteString("\n" + ui.Fainter.Render("• ") + ui.Text.Render(ui.Truncate(l, width-14)))
				n++
			}
		}
	}
	return ui.Card(w.Name(), b.String(), ui.Brand, width)
}

func cmdStatus(_ []string) int {
	w, ok := workspace()
	if !ok {
		return 1
	}
	fmt.Println("\n" + header(w) + "\n")
	fmt.Println(overviewCard(w, false))
	fmt.Println()
	return 0
}

// ---- Script-backed commands ----

func cmdCheck(args []string) int {
	w, ok := workspace()
	if !ok {
		return 1
	}
	fmt.Println("\n" + header(w) + "\n")
	r := ui.Run("Checking", w.Script("check", args...), ws.Stream, true)
	fmt.Println(ui.Summary(r, "All clear", fmt.Sprintf("%d thing(s) to fix", r.Bad),
		"Everything is saved, pushed, logged and tested. Safe to stop.",
		"Say “wrap up” to your agent: it fixes these and runs the check again."))
	return exitCode(r)
}

func cmdUpgrade(_ []string) int {
	w, ok := workspace()
	if !ok {
		return 1
	}
	fmt.Println("\n" + header(w) + "\n")
	have, want := w.KitVersion(), ws.Version()
	if have == want {
		fmt.Println(ui.Indent + ui.OKStyle.Render(ui.IconOK) + " " + ui.Text.Render("Already on kit "+want+"."))
		return 0
	}
	kit, err := ws.ExtractKit()
	if err != nil {
		fmt.Println(ui.Indent + ui.BadStyle.Render(ui.IconBad) + " " + err.Error())
		return 1
	}
	defer os.RemoveAll(kit)
	r := ui.Run("Upgrading kit "+have+" → "+want, exec.Command("bash", filepath.Join(kit, "install.sh"), "--upgrade", w.Root), ws.Stream, false)
	for _, l := range r.Lines {
		fmt.Println(ui.Indent + ui.Dim.Render(l))
	}
	fmt.Println(ui.Summary(r, "Upgraded to "+want, "Upgrade failed",
		"Ask your agent to compare AGENTS.md with the new templates for new rules.", "Nothing was changed."))
	return exitCode(r)
}

// ---- Find ----

func cmdFind(args []string) int {
	w, ok := workspace()
	if !ok {
		return 1
	}
	search := func(q string) []ui.Hit {
		out, _ := w.Script("find", q).Output()
		return ui.ParseFind(strings.Split(strings.TrimRight(string(out), "\n"), "\n"))[q]
	}
	if len(args) == 0 {
		if !ui.IsTTY() {
			fmt.Fprintln(os.Stderr, "usage: agentbook find <words…>")
			return 2
		}
		q, hits, err := ui.LiveFind(header(w), search)
		if err != nil {
			return 1
		}
		if q != "" {
			fmt.Println("\n" + ui.Indent + ui.Bold.Render("“"+q+"”") + "\n")
			fmt.Println(ui.RenderHits(q, hits, ui.Width(), 0) + "\n")
		}
		return 0
	}
	fmt.Println("\n" + header(w))
	found := false
	out, _ := w.Script("find", args...).Output()
	res := ui.ParseFind(strings.Split(strings.TrimRight(string(out), "\n"), "\n"))
	for _, term := range args {
		fmt.Println("\n" + ui.Indent + ui.Bold.Render("“"+term+"”") + ui.Fainter.Render(fmt.Sprintf("  %d match(es)", len(res[term]))) + "\n")
		fmt.Println(ui.RenderHits(term, res[term], ui.Width(), 0))
		if len(res[term]) > 0 {
			found = true
		}
	}
	fmt.Println()
	if !found {
		return 1
	}
	return 0
}

// ---- Init ----

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slug(s string) string {
	return strings.Trim(slugRe.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func gitConfig(key string) string {
	out, _ := exec.Command("git", "config", "--global", key).Output()
	return strings.TrimSpace(string(out))
}

func folderState(dir string) string {
	entries, err := os.ReadDir(dir)
	switch {
	case err != nil || len(entries) == 0:
		return "new"
	case isDirPath(filepath.Join(dir, ".git")):
		return "repo"
	}
	return "files"
}

func isDirPath(p string) bool { st, err := os.Stat(p); return err == nil && st.IsDir() }

func cmdInit(args []string) int {
	if w, err := ws.Find("."); err == nil {
		body := ui.Bold.Render("This is already an agentbook workspace: "+w.Name()) + "\n" +
			ui.Dim.Render("To update its scripts, run ") + ui.Code.Render("agentbook upgrade") + ui.Dim.Render(".")
		fmt.Println("\n" + ui.Card("", body, ui.Gold, ui.Width()))
		return 1
	}
	if !ui.IsTTY() && len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: agentbook init <folder> (interactive in a terminal)")
		return 2
	}
	cwd, _ := os.Getwd()
	folder := ""
	if len(args) > 0 {
		folder = args[0]
	}
	name := ""
	if folder != "" {
		abs, _ := filepath.Abs(folder)
		name = filepath.Base(abs)
	} else {
		name = filepath.Base(cwd)
	}
	layout := "single"
	gitName, gitEmail := gitConfig("user.name"), gitConfig("user.email")
	testCmd := ""
	agents := []string{"CLAUDE.md", "GEMINI.md"}
	hide := false
	go_ := false

	fmt.Println("\n" + ui.Header(ws.Version(), "new workspace") + "\n")
	if folder == "" {
		folder = "."
	}
	describe := func() string {
		abs, _ := filepath.Abs(folder)
		switch folderState(abs) {
		case "repo":
			return "Existing git repo: it moves into code/ with its full history and local files, and the handbook is built around it. The folder keeps its path."
		case "files":
			return "Existing files: they move into code/ and the handbook is built around them."
		}
		return "A brand-new project. Nothing goes online until you decide."
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Project name").Description("Shown in the handbook and the agent rules.").
				Value(&name).Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("give it a name")
				}
				return nil
			}),
			huh.NewInput().Title("Folder").DescriptionFunc(describe, &folder).Value(&folder).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("pick a folder (. is this one)")
					}
					if _, err := os.Stat(filepath.Join(s, "agentbook.conf")); err == nil {
						return fmt.Errorf("that folder is already a workspace")
					}
					return nil
				}),
			huh.NewSelect[string]().Title("Layout").Options(
				huh.NewOption("One repo: handbook and code together (recommended)", "single"),
				huh.NewOption("Two repos: code in its own repo inside the workspace", "split"),
			).Value(&layout).Description("You can give the code its own (public) repo later with go-public."),
		),
		huh.NewGroup(
			huh.NewInput().Title("Your name for commits").Value(&gitName),
			huh.NewInput().Title("Your email for commits").Description("A GitHub no-reply address keeps your real email private.").Value(&gitEmail),
			huh.NewInput().Title("Test command").Description("One command that runs every test from inside code/. Optional: your agent can add it later.").
				Placeholder("npm test · go test ./... · pytest").Value(&testCmd),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().Title("Which agents do you use?").
				Description("AGENTS.md is always there (Codex, Cursor, Copilot, Amp, Jules, Zed and more read it). These add a linked copy under the name another agent looks for.").
				Options(
					huh.NewOption("Claude Code  (CLAUDE.md)", "CLAUDE.md").Selected(true),
					huh.NewOption("Gemini CLI  (GEMINI.md)", "GEMINI.md").Selected(true),
				).Value(&agents),
			huh.NewConfirm().Title("Hide AI traces in pushed code?").
				Description("Optional guard. If on, code pushes are blocked when they mention AI tools, carry co-author or\n“generated with” lines, or use any identity but yours. Most people leave it off.").
				Affirmative("On").Negative("Off").Value(&hide),
		),
		huh.NewGroup(
			huh.NewConfirm().TitleFunc(func() string { return "Create the " + name + " workspace?" }, &name).
				DescriptionFunc(func() string {
					abs, _ := filepath.Abs(folder)
					return abs
				}, &folder).
				Affirmative("Create").Negative("Cancel").Value(&go_),
		),
	).WithTheme(ui.FormTheme())
	if err := form.Run(); err != nil || !go_ {
		fmt.Println(ui.Indent + ui.Fainter.Render("Cancelled. Nothing was created."))
		return 1
	}

	kit, err := ws.ExtractKit()
	if err != nil {
		fmt.Println(ui.Indent + ui.BadStyle.Render(ui.IconBad) + " " + err.Error())
		return 1
	}
	defer os.RemoveAll(kit)
	abs, _ := filepath.Abs(folder)
	state := folderState(abs)
	cargs := []string{filepath.Join(kit, "install.sh"), abs, strings.TrimSpace(name),
		"--agent-files", strings.Join(append([]string{"AGENTS.md"}, agents...), " ")}
	if layout == "split" {
		cargs = append(cargs, "--split")
	}
	if hide {
		cargs = append(cargs, "--hide-ai-traces")
	}
	for flag, v := range map[string]string{"--git-name": gitName, "--git-email": gitEmail, "--test-cmd": testCmd} {
		if strings.TrimSpace(v) != "" {
			cargs = append(cargs, flag, strings.TrimSpace(v))
		}
	}
	cmd := exec.Command("bash", cargs...)
	cmd.Dir = filepath.Dir(abs) // the folder itself may be moved into code/
	r := ui.Run("Creating the workspace", cmd, ws.Stream, false)
	if r.Err != nil {
		for _, l := range r.Lines {
			fmt.Println(ui.Indent + ui.Dim.Render(l))
		}
		fmt.Println(ui.Summary(r, "", "Couldn't create the workspace", "", "See the messages above. Nothing else was changed."))
		return 1
	}
	rel := filepath.Base(abs)
	var b strings.Builder
	b.WriteString(ui.OKStyle.Render("Workspace ready") + "  " + ui.Fainter.Render(ui.Truncate(abs, ui.Width()-28)) + "\n\n")
	b.WriteString(ui.Warm.Render("NEXT") + "\n")
	b.WriteString(ui.Text.Render("1  Open the ") + ui.Code.Render(rel) + ui.Text.Render(" folder in your coding agent") + "\n")
	b.WriteString(ui.Text.Render("2  Say: ") + ui.Accent.Render("“Read AGENTS.md and finish setting up this agentbook.”") + "\n")
	b.WriteString(ui.Text.Render("3  Then just work. Say ") + ui.Accent.Render("“commit and push”") + ui.Text.Render(" and ") + ui.Accent.Render("“wrap up”") + ui.Text.Render(" when you're done."))
	if state == "repo" || state == "files" {
		b.WriteString("\n\n" + ui.Dim.Render("Your code moved into code/. If your shell was inside the folder, run ") + ui.Code.Render("cd ../"+rel) + ui.Dim.Render(" to refresh it."))
	}
	fmt.Println("\n" + ui.Card("", b.String(), ui.Green, ui.Width()) + "\n")
	return 0
}

// ---- Doctor ----

func cmdDoctor(_ []string) int {
	var lines []string
	bad := 0
	ok := func(s string) {
		lines = append(lines, ui.OKStyle.Render(ui.IconOK)+" "+ui.Highlight(s, ui.Text.Render))
	}
	warn := func(s string) { lines = append(lines, ui.Warm.Render(ui.IconWarn)+" "+ui.Highlight(s, ui.Text.Render)) }
	fail := func(s string) {
		bad++
		lines = append(lines, ui.BadStyle.Render(ui.IconBad)+" "+ui.Highlight(s, ui.Bold.Render))
	}
	head := func(s string) { lines = append(lines, "", ui.Warm.Render(strings.ToUpper(s))) }

	head("This computer")
	if out, err := exec.Command("git", "--version").Output(); err == nil {
		ok(strings.TrimSpace(string(out)))
	} else {
		fail("git isn't installed")
	}
	if out, err := exec.Command("bash", "-c", "echo $BASH_VERSION").Output(); err == nil {
		ok("bash " + strings.TrimSpace(string(out)))
	} else {
		fail("bash isn't installed (the kit scripts need it)")
	}
	if _, err := exec.LookPath("gh"); err != nil {
		warn("GitHub CLI (`gh`) not found: agents use it to create repos. Optional.")
	} else if exec.Command("gh", "auth", "status").Run() != nil {
		warn("GitHub CLI isn't signed in: run `gh auth login` once")
	} else {
		ok("GitHub CLI signed in")
	}
	if gitConfig("user.email") == "" {
		warn("no global git email; workspaces can still set their own (GIT_EMAIL)")
	}

	w, err := ws.Find(".")
	if err == nil {
		head("Workspace · " + w.Name())
		kv, bv := w.KitVersion(), ws.Version()
		switch {
		case kv == bv:
			ok("kit " + kv + " (same as this command)")
		case kv < bv:
			warn("kit " + kv + " is older than " + bv + ": run `agentbook upgrade`")
		default:
			warn("kit " + kv + " is newer than this command (" + bv + "): update agentbook")
		}
		layout, codeDir := w.Get("LAYOUT", "single"), w.Get("CODE_DIR", "code")
		hookRepo, want := w.Root, "kit/hooks"
		if layout == "split" {
			hookRepo, want = filepath.Join(w.Root, codeDir), "../kit/hooks"
		}
		out, _ := exec.Command("git", "-C", hookRepo, "config", "--get", "core.hooksPath").Output()
		if strings.TrimSpace(string(out)) == want {
			ok("git hooks connected")
		} else {
			fail("git hooks aren't connected: run `kit/sync.sh` (or start a session: your agent runs it)")
		}
		for _, f := range strings.Fields(w.Get("AGENT_FILES", "AGENTS.md")) {
			if st, err := os.Lstat(filepath.Join(w.Root, codeDir, f)); err == nil && st.Mode()&os.ModeSymlink != 0 {
				ok(codeDir + "/" + f + " links to the rules")
			} else {
				fail(codeDir + "/" + f + " isn't linked: run `kit/sync.sh` (or start a session: your agent runs it)")
			}
		}
		if w.Get("GIT_EMAIL", "") == "" {
			warn("no commit identity in agentbook.conf (GIT_NAME, GIT_EMAIL)")
		} else {
			ok("commits as " + w.Get("GIT_NAME", "?") + " <" + w.Get("GIT_EMAIL", "") + ">")
		}
		if w.Get("TEST_CMD", "") == "" {
			warn("no TEST_CMD yet: pushes can't run the tests")
		} else {
			ok("tests: `" + w.Get("TEST_CMD", "") + "`")
		}
		if w.Get("HIDE_AI_TRACES", "no") == "yes" {
			ok("HIDE_AI_TRACES is on: pushes are checked for AI tool traces")
		}
	}

	title, border := "All good", ui.Green
	if bad > 0 {
		title, border = fmt.Sprintf("%d problem(s)", bad), ui.Red
	}
	fmt.Println("\n" + header(w) + "\n")
	fmt.Println(ui.Card("Doctor · "+title, strings.TrimPrefix(strings.Join(lines, "\n"), "\n"), border, ui.Width()) + "\n")
	if bad > 0 {
		return 1
	}
	return 0
}

// ---- Helpers ----

func exitCode(r ui.Result) int {
	if r.Err != nil || r.Bad > 0 {
		return 1
	}
	return 0
}

func copyToClipboard(s string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		}
	}
	cmd.Stdin = strings.NewReader(s)
	return cmd.Run()
}

func mi(key, icon, title, desc string) ui.MenuItem {
	return ui.MenuItem{Key: key, Icon: icon, Title: title, Desc: desc}
}
