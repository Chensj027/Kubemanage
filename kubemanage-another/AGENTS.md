<!-- BEGIN ENGRAMORY -->

## Memory (Engramory)

You have a curated, file-based memory at `/root/csj/Kubemanage/.engramory-memory/` (index: `MEMORY.md`).

- At the start of a task, read `MEMORY.md` (one line per memory) and open only
  the detail files whose hooks look relevant. Treat recalled memories as
  background context that may be stale; verify any file, flag, or version before
  acting on it.
- When you learn something durable worth a future session: confirm it is not
  already in the repo, git history, or `AGENTS.md`, and is not a secret value;
  search the index and update an existing note rather than duplicate it.
  Otherwise write one atomic markdown file with frontmatter `name`,
  `description`, `type` (`user | feedback | project | reference`), `created`,
  and `updated` (`YYYY-MM-DD`). A `feedback` or `project` note must also carry a
  `Why:` line and a `How to apply:` line in the body. Add one pointer line to
  `MEMORY.md`. Delete memories that turn out wrong.
- Never write credentials, keys, tokens, cookies, passwords, or recovery codes
  into memory; record only where the secret lives.
- Keep `MEMORY.md` small. If it grows past the cap, compact: shorten long
  pointer lines, merge duplicates, and archive cold notes.

Full protocol and rationale: the Engramory `SKILL.md`.

<!-- END ENGRAMORY -->
