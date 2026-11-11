// Command agentbook keeps a project's handbook so any coding agent, on any computer, can pick up
// exactly where the last session left off.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	_ "github.com/EnigmaticZee/agentbook/internal/early" // must load before the terminal libraries
	"github.com/EnigmaticZee/agentbook/internal/ui"
	"github.com/EnigmaticZee/agentbook/internal/ws"
)

type command struct {
	name, args, short string
	run               func(args []string) int
	group             string
}

var commands []command

func init() {
	commands = []command{
		{"init", "[folder]", "Create a workspace for a new or existing project", cmdInit, "Start"},
		{"status", "", "Where the project stands and what's next", cmdStatus, "Every session"},
		{"find", "[words]", "Search everything recorded (no words: live)", cmdFind, "Every session"},
		{"check", "[--no-tests]", "Saved, pushed, logged and tested?", cmdCheck, "Every session"},
		{"doctor", "", "Look for problems on this computer", cmdDoctor, "Tools"},
		{"upgrade", "", "Update this workspace's kit/", cmdUpgrade, "Tools"},
		{"prompt", "[--copy]", "The setup prompt for any coding agent", cmdPrompt, "Tools"},
		{"version", "", "Show the version", cmdVersion, "Tools"},
		{"help", "", "Show this help", cmdHelp, "Tools"},
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		os.Exit(cmdHome())
	}
	name := args[0]
	switch name {
	case "-h", "--help":
		name = "help"
	case "-v", "--version":
		name = "version"
	}
	for _, c := range commands {
		if c.name == name {
			os.Exit(c.run(args[1:]))
		}
	}
	fmt.Fprintln(os.Stderr, ui.Indent+ui.BadStyle.Render(ui.IconBad)+" "+ui.Bold.Render("Unknown command “"+name+"”")+ui.Dim.Render(". Run ")+ui.Code.Render("agentbook help")+ui.Dim.Render("."))
	os.Exit(2)
}

// workspace finds the workspace from the current folder or explains how to get one.
func workspace() (*ws.Workspace, bool) {
	w, err := ws.Find(".")
	if err == nil {
		return w, true
	}
	if errors.Is(err, ws.ErrNoWorkspace) {
		body := ui.Bold.Render("No agentbook workspace here.") + "\n" +
			ui.Dim.Render("Run this inside a workspace (or its code/ folder), or create one with ") + ui.Code.Render("agentbook init") + ui.Dim.Render(".")
		fmt.Println("\n" + ui.Card("", body, ui.Amber, ui.Width()))
	} else {
		fmt.Println(ui.Indent + ui.BadStyle.Render(ui.IconBad) + " " + err.Error())
	}
	return nil, false
}

func header(w *ws.Workspace) string {
	ctx := ""
	if w != nil {
		ctx = w.Name()
	}
	return ui.Header(ws.Version(), ctx)
}

func cmdVersion(_ []string) int {
	fmt.Println("agentbook " + ws.Version())
	return 0
}

func cmdHelp(_ []string) int {
	w := ui.Width()
	var b strings.Builder
	b.WriteString("\n" + ui.Header(ws.Version(), "") + "\n")
	b.WriteString(ui.Indent + ui.Italic.Render("Your project remembers everything. Any agent, any computer.") + "\n")
	group := ""
	for _, c := range commands {
		if c.group != group {
			group = c.group
			b.WriteString("\n" + ui.Indent + ui.Warm.Render(strings.ToUpper(group)) + "\n")
		}
		left := ui.Accent.Render(c.name)
		if c.args != "" {
			left += " " + ui.Fainter.Render(c.args)
		}
		pad := 26 - lipgloss.Width(left)
		if pad < 2 {
			pad = 2
		}
		b.WriteString(ui.Indent + "  " + left + strings.Repeat(" ", pad) + ui.Dim.Render(ui.Truncate(c.short, w-32)) + "\n")
	}
	b.WriteString("\n" + ui.Indent + ui.Dim.Render("Run ") + ui.Code.Render("agentbook") + ui.Dim.Render(" on its own for the interactive menu.") + "\n")
	b.WriteString(ui.Indent + ui.Fainter.Render("Docs: https://enigmaticzee.github.io/agentbook") + "\n")
	fmt.Println(b.String())
	return 0
}

func cmdPrompt(args []string) int {
	p := ws.Prompt()
	if i := strings.Index(p, "\n---\n"); i >= 0 {
		p = strings.TrimSpace(p[i+5:])
	}
	if len(args) > 0 && args[0] == "--copy" {
		if err := copyToClipboard(p); err != nil {
			fmt.Println(ui.Indent + ui.BadStyle.Render(ui.IconBad) + " couldn't copy: " + err.Error())
			return 1
		}
		fmt.Println(ui.Indent + ui.OKStyle.Render(ui.IconOK) + " " + ui.Text.Render("Setup prompt copied. Paste it into your coding agent."))
		return 0
	}
	fmt.Println(p)
	return 0
}
