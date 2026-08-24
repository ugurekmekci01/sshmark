# sshmark

Project bookmarks for SSH. One command to open the right machine, folder, and tunnels.

```bash
sshmark open api
```

## Requirements

- OpenSSH client on your PATH: `ssh`, `ssh-copy-id`, `ssh-keygen`
- macOS or Linux (amd64 or arm64)

## Install

### Install script (recommended)

```bash
curl -fsSL https://github.com/ugurekmekci01/sshmark/releases/latest/download/install.sh | bash
```

Installs to `~/.local/bin/sshmark`. If `sshmark` is not on your PATH yet, add this to `~/.zshrc` or `~/.bashrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Go install

```bash
go install github.com/ugurekmekci01/sshmark/cmd/sshmark@latest
```

Ensure `$(go env GOPATH)/bin` is on your PATH.

### Manual download

1. Download the archive for your OS/arch from [GitHub Releases](https://github.com/ugurekmekci01/sshmark/releases).
2. Verify the SHA-256 checksum from `checksums.txt`.
3. Extract and move `sshmark` to a directory on your PATH (e.g. `~/.local/bin`).

### Build from source

```bash
git clone https://github.com/ugurekmekci01/sshmark.git
cd sshmark
go build -o bin/sshmark ./cmd/sshmark
./bin/sshmark version
```

For local development, run `./bin/sshmark` directly or add `bin/` to your PATH.

## Quick start

```bash
sshmark add api      # set up once (SSH keys if machine is new)
sshmark open api     # every day
```

### What `add` does

1. Project name
2. Machine (IP or nickname; default nickname `{project}-box`)
3. SSH user (default: your local `$USER`)
4. Remote directory (default: `/home/<user>/dev`; created on remote if missing)
5. Tunnels (optional)

On a **new machine**, sshmark also runs `ssh-copy-id` and writes `~/.ssh/config`. Same machine again → skips that.

### What `open` does

1. Load bookmark
2. Check local ports
3. Ensure remote directory exists
4. SSH + tunnels + `cd` into folder
5. Exit shell → tunnels close

Run your app manually (`npm run dev`, etc.).

## Example

| Project | Machine | Directory | Tunnels |
|---------|---------|-----------|---------|
| `api` | buildbox | `/home/user/dev` | 3000, 5432 |
| `train` | gpu | `/home/user/ml` | 8888 |

```bash
sshmark open api     # → buildbox
sshmark open train   # → gpu
```

## Power users

Edit `~/.config/sshmark/projects/api.toml`:

```toml
name = "api"
host = "buildbox"
user = "user"
directory = "/home/user/dev"

[[tunnels]]
local_port = 3000
remote_port = 3000
```

Config directory: `~/.config/sshmark/projects/` (override with `SSHMARK_CONFIG`).

## Commands

| Command | Purpose |
|---------|---------|
| `add` | Create bookmark |
| `open` | Connect |
| `list` / `show` | View bookmarks |
| `check` | Validate without opening a shell |
| `edit` / `remove` | Manage bookmarks |
| `version` | Print version |

## Development

```bash
make setup    # pre-commit hook → make ci
make test
make ci
```

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
