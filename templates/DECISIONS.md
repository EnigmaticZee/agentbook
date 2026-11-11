# {{PROJECT}}: decisions

Append-only, newest last. Never edit or delete an entry; a reversal or correction is a new entry that
names the one it replaces. Use this shape:

```
## YYYY-MM-DD: <short title>
- Decision: what was decided.
- Why: the reason (or "reason not recorded").
- Rejected: the alternatives and why they lost (or "none considered").
```

## {{DATE}}: Use a private handbook
- Decision: the project lives in a private workspace repo (the handbook) with the code repo in `code/`;
  the handbook is never inside the code repo.
- Why: any person or agent can resume on any computer, and the code repo can go public at any time
  without exposing notes or history.
- Rejected: notes inside the code repo (they'd be in its history forever, so going public would mean
  rewriting history).
