# {{PROJECT}}: architecture

## Repos and stages

| Stage | Layout | Repos |
| --- | --- | --- |
| local | single | one repo on this computer: handbook + `code/` |
| private | single | one private GitHub repo `<project>`: handbook + `code/` |
| private (two repos) | split | private `<project>` (code) + private `<project>-handbook` (workspace) |
| public | split | public `<project>` (code) + private `<project>-handbook` (workspace) |

The workspace is what you clone and open. Current stage, layout and remotes: `agentbook.conf` and
`STATUS.md`.

### New computer
Clone the workspace repo and open it in an agent. Its first step (`kit/sync.sh`, per AGENTS.md)
installs dependencies and connects hooks, links and identity; in the split layout it also clones
`CODE_REMOTE` into `code/`.

### Moving between stages
- **local → private (single):** create one private repo `<project>`; add it as `origin`; push; set
  `STAGE="private"` and `HANDBOOK_REMOTE`; log it in DECISIONS.
- **single → public:** `kit/go-public.sh --check` shows exactly what would be published (only commits
  that changed `code/`, rooted at `code/`) and audits it. Go through every finding with the owner.
  With their yes: rename the private repo to `<project>-handbook`, create an empty public
  `<project>`, run `kit/go-public.sh <its URL>`. The workspace switches to split, `code/` becomes a
  clone of the public repo (local-only files stay), and `STAGE`, `LAYOUT`, `CODE_REMOTE` update. Log it.
- **single → separate private code repo:** same, with `--stage private`.
- **split, private code → public:** run `kit/public-audit.sh`; fix findings with the owner (history
  rewrites need their approval); flip the code repo to public; set `STAGE="public"`; log it.
- **public → private:** flip the code repo's visibility and set `STAGE`. Anything already published
  may have been copied; note that in DECISIONS.

## Files and folders (paths inside `code/`)
| Path | What it is |
| --- | --- |

## How it works
Data flow, main components, APIs and messages.

## Storage
Storage keys, schemas, file locations, where settings live.

## External services and accounts
What the project talks to, and where credentials are kept (never the credentials themselves).

## Build, release and tests
Setup, build, test and release commands, and what the tests cover and don't.
