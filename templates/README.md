# {{PROJECT}} (private workspace)

<sub>Kept by [Agentbook](https://github.com/EnigmaticZee/agentbook).</sub>

This is the repo you clone to work on {{PROJECT}}. It holds the project's handbook (status, decisions,
history, how it works) and the code, in `code/` (if the code has its own public repo, it's fetched
into `code/` for you).

## How you work

1. **Clone this repo** (any way you like: `git clone`, GitHub Desktop, your editor).
2. **Open the folder in your coding agent** and say what you want to do. Ask anything about the
   project's past ("did we look into…?", "why did we drop…?"); it answers from the records.
3. **Ask it to commit and push** when you're happy.
4. **Say "wrap up"** when you stop for the day.

That's all. The agent reads `AGENTS.md` by itself. The first thing it does in every session is fetch
or update the code (into `code/`), install what it can, and connect the hooks. Every commit is logged
here, pushes are checked, and the handbook is pushed along with the code.

On a brand-new computer the agent may tell you about a one-time manual step, such as installing an
app or signing in to GitHub.

## Good to know

- **The code is in `code/`.** Open that folder in other apps (Xcode, an editor) if you want.
- **Want to go public later?** Say "Get this ready to go public." Your notes stay private.
- **Same steps on every computer.** Whatever you did yesterday on another computer is here, as long
  as that session ended with "wrap up" (or a push).
- **Nothing private reaches the code repo.** Even if the code is public, this workspace stays private.

## If something goes wrong

| You see | Say to the agent |
| --- | --- |
| It says the sync failed | "Sort out the sync." Nothing was lost. |
| A push is blocked | Usually nothing: it updates the handbook or fixes the tests. To push anyway: "push without the checks this time". |
| It doesn't seem to know the project | "Read AGENTS.md and follow it." |
| You want to double-check everything is saved | "Run the check." |
