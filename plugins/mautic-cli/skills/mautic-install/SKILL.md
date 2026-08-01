---
name: mautic-install
description: Install or update the `mautic` command line client for the Mautic REST API, from an existing binary on PATH, a GitHub release, or source. Use when `mautic` is not found, when `mautic llm` is not a known subcommand, or before running the mautic-usage skill on a machine that has never had it.
license: MIT
compatibility: Needs curl and tar; the source route additionally needs a Go 1.24+ toolchain. Installs into the first writable directory already on PATH and never runs sudo.
allowed-tools: Bash(command:*) Bash(curl:*) Bash(tar:*) Bash(uname:*) Bash(mkdir:*) Bash(mv:*) Bash(chmod:*) Bash(go:*) Bash(mautic:*) Read
---

# Installing mautic

Three routes, in this order. Stop at the first that works.

## Route 1 — a binary already on PATH

```bash
command -v mautic && mautic --version
```

If it resolves, **use it and stop**. Do not check for a newer release; that is
an API call nobody asked for. Two things to confirm first:

- **Is it the right tool?** `mautic` is a short name and Mautic's own PHP
  console is sometimes installed under it. `mautic llm | head -1` must print
  `# mautic — guide`. If something else owns the name, say so and use an
  explicit path rather than shadowing it.
- **Is it recent enough?** If `mautic llm` reports an unknown subcommand, the
  binary predates the embedded reference. Continue to Route 2 to upgrade.

## Route 2 — the latest GitHub release

```bash
VERSION=$(curl -fsSL https://api.github.com/repos/ideamans/mautic-cli/releases/latest \
  | grep '"tag_name"' | head -1 | cut -d'"' -f4)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64
[ "$ARCH" = "aarch64" ] && ARCH=arm64

curl -fsSL -o /tmp/mautic.tar.gz \
  "https://github.com/ideamans/mautic-cli/releases/download/${VERSION}/mautic-cli_${VERSION#v}_${OS}_${ARCH}.tar.gz"
tar -xzf /tmp/mautic.tar.gz -C /tmp
```

Three things that are only knowable from a real release:

- The archive is named after goreleaser's `project_name`, which is
  **`mautic-cli`**, and the version inside the filename has **no leading `v`**.
- `uname -m` says `x86_64` and `aarch64`; goreleaser says `amd64` and `arm64`.
  Map both.
- **The binary inside the archive is `mautic`**, not `mautic-cli`.

If a download 404s, list what the release actually has rather than guessing
variations:

```bash
curl -fsSL https://api.github.com/repos/ideamans/mautic-cli/releases/latest \
  | grep '"name"' | grep -E 'tar\.gz|zip'
```

Then install into the first writable directory already on PATH:

```bash
mkdir -p "$HOME/.local/bin" && mv /tmp/mautic "$HOME/.local/bin/mautic" && chmod +x "$HOME/.local/bin/mautic"
```

Two things not to decide alone:

- **Nothing on PATH is writable** → leave the binary in `/tmp`, print the exact
  `sudo mv` line, and let the user run it. **Do not run `sudo`.**
- **`~/.local/bin` is not on PATH** → give the user the line for their shell
  profile. **Do not edit the profile.**

## Route 3 — build from source

Last, because it needs a toolchain and a compile. It is the escape hatch when
the release matrix has no asset for the platform.

```bash
go install github.com/ideamans/mautic-cli/cmd/mautic@latest
```

This installs a command called `mautic` into `$(go env GOPATH)/bin`, matching
the name in the release archive.

## Verify

```bash
command -v mautic && mautic --version && mautic llm | head -1
```

Report the version and the location. Then run `mautic auth status` — it works
with nothing configured and prints what credentials are still needed, which is
the natural next step into the **mautic-usage** skill.

## Failure modes

| Symptom | Fix |
| --- | --- |
| `curl: (22) ... 404` on the archive | the asset name is wrong — list the release assets with the command above and match exactly |
| `mautic: command not found` right after installing | the install directory is not on PATH; print the `export PATH=…` line for the user to add |
| `cannot execute binary file` | the OS/arch pair was wrong — re-check the `uname` mapping |
| `permission denied` moving into `/usr/local/bin` | do not escalate; print the `sudo mv` line and let the user run it |
| `mautic llm` reports an unknown subcommand | the binary is older than the embedded reference — upgrade via Route 2 |
| `go: module ... not found` | the repository is public; check network access and that the module path is `github.com/ideamans/mautic-cli` |
