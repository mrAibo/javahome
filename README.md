# javahome

`javahome` is a small, dependency-free Java home switcher written in Go.

It helps you answer three everyday questions:

1. Which JDKs are installed on this machine?
2. Which Java is currently active?
3. How do I switch to Java 8, 11, 17, or 21 without fragile shell hacks?

`javahome` discovers installed JDKs, validates `JAVA_HOME`, emits shell-specific activation commands, can update your shell profile on request, and includes a `doctor` command for quick diagnostics.

## Features

- single Go binary
- no third-party runtime dependencies
- Linux, macOS, and Windows path support
- Bash, Zsh, Fish, PowerShell, and CMD snippets
- current-shell activation without editing files
- optional permanent profile update via `--global`
- optional project-local `.javahome.toml` via `--project`
- stale Java `bin` entries removed from generated `PATH`
- JSON output for scripts and automation
- safe preview mode with `--dry-run`

## Quick start

Build or install the tool:

```bash
go install github.com/mrAibo/javahome/cmd/javahome@latest
```

Then run:

```bash
javahome list
javahome doctor
```

Switch Java for the current terminal session:

```bash
eval "$(javahome use 17 --shell bash)"
```

Make Java 17 permanent for Bash:

```bash
javahome use 17 --global --shell bash
```

Need another shell? Use the matching command from the examples below.

## Build locally

```bash
git clone https://github.com/mrAibo/javahome.git
cd javahome
make test
make build
./bin/javahome doctor
```

Without `make`:

```bash
go test ./...
go build -o bin/javahome ./cmd/javahome
```

## Choose your scenario

| Scenario | Command |
|---|---|
| List found JDKs | `javahome list` |
| Show active Java | `javahome current` |
| Print path for Java 17 | `javahome print 17` |
| Current Bash session | `eval "$(javahome use 17 --shell bash)"` |
| Current Zsh session | `eval "$(javahome use 17 --shell zsh)"` |
| Current Fish session | `javahome use 17 --shell fish \| source` |
| Current PowerShell session | `javahome use 17 --shell powershell \| Invoke-Expression` |
| Permanent Bash default | `javahome use 17 --global --shell bash` |
| Permanent Zsh default | `javahome use 17 --global --shell zsh` |
| Permanent Fish default | `javahome use 17 --global --shell fish` |
| Permanent PowerShell default | `javahome use 17 --global --shell powershell` |
| Project-local preference | `javahome use 17 --project` |
| Preview before writing | `javahome use 17 --global --shell bash --dry-run` |
| Machine-readable output | `javahome list --json` |
| Diagnose setup | `javahome doctor` |

## Linux examples

### Temporary switch in Bash

```bash
javahome list
eval "$(javahome use 17 --shell bash)"
java -version
echo "$JAVA_HOME"
```

### Permanent Bash default

```bash
javahome use 21 --global --shell bash
source ~/.bashrc
java -version
```

### Zsh user

```bash
javahome use 17 --global --shell zsh
source ~/.zshrc
```

### Fish user

```fish
javahome use 17 --shell fish | source
java -version
```

Permanent Fish default:

```fish
javahome use 17 --global --shell fish
source ~/.config/fish/config.fish
```

### Server or CI script

```bash
JAVA_HOME="$(javahome print 17)"
PATH="$JAVA_HOME/bin:$PATH"
export JAVA_HOME PATH
java -version
```

## macOS examples

macOS usually uses Zsh by default:

```zsh
javahome list
eval "$(javahome use 17 --shell zsh)"
java -version
```

Make Java 17 the default for future Zsh sessions:

```zsh
javahome use 17 --global --shell zsh
source ~/.zshrc
```

If you use Bash on macOS:

```bash
javahome use 17 --global --shell bash
source ~/.bashrc
```

## Windows examples

### PowerShell current session

```powershell
javahome list
javahome use 17 --shell powershell | Invoke-Expression
java -version
$env:JAVA_HOME
```

### PowerShell permanent default

```powershell
javahome use 17 --global --shell powershell
. $PROFILE
```

If PowerShell script execution is restricted, open a new PowerShell window after running the `--global` command.

### CMD current session

CMD cannot safely evaluate multi-line command output the same way Bash or PowerShell can. Print the commands first:

```cmd
javahome use 17 --shell cmd
```

Then paste the printed `set JAVA_HOME=...` and `set PATH=...` lines into the same CMD window.

## Project-local example

Inside a project that requires Java 17:

```bash
cd my-java-project
javahome use 17 --project
cat .javahome.toml
```

This creates a small config file that documents the intended Java version for the project. You can commit it if the whole team should use the same Java major version.

## Shell helper: `jhome`

Install a small helper function into your profile:

```bash
javahome init bash --global
```

Then switch with a shorter command:

```bash
jhome 17
jhome 21
```

For Zsh, Fish, or PowerShell:

```bash
javahome init zsh --global
javahome init fish --global
javahome init powershell --global
```

## Commands

```text
javahome list [--json]
javahome current [--json]
javahome print [version] [--vendor text] [--json]
javahome use <version> [--vendor text] [--shell bash|zsh|fish|powershell|cmd]
javahome use <version> --global [--shell bash|zsh|fish|powershell]
javahome use <version> --project
javahome doctor [--json]
javahome init [bash|zsh|fish|powershell] [--global]
javahome version
```

## Discovery locations

Linux examples:

- `/usr/lib/jvm/*`
- `/usr/java/*`
- `/opt/java/*`
- `/opt/jdk*`
- `~/.sdkman/candidates/java/*`
- `~/.asdf/installs/java/*`
- `~/.local/share/mise/installs/java/*`

macOS examples:

- `/Library/Java/JavaVirtualMachines/*.jdk/Contents/Home`
- Homebrew OpenJDK locations
- SDKMAN/asdf/mise Java locations

Windows examples:

- `%ProgramFiles%\Java\*`
- `%ProgramFiles%\Eclipse Adoptium\*`
- `%ProgramFiles%\Microsoft\jdk-*`
- `%ProgramFiles%\Amazon Corretto\*`
- `%USERPROFILE%\.jdks\*`

## Important shell note

No external program can directly change the environment of an already-running parent shell. That is why current-shell activation uses `eval`, `source`, or `Invoke-Expression`.

For permanent activation, use `--global`; it writes a marked `javahome` block to the relevant profile file.

## Troubleshooting

Start with:

```bash
javahome doctor
javahome list
javahome current
```

Common problems:

| Problem | Fix |
|---|---|
| `No Java installations found` | Install a JDK or put it in a known location such as `/usr/lib/jvm`, `/Library/Java/JavaVirtualMachines`, `%ProgramFiles%\Java`, SDKMAN, asdf, or mise. |
| `JAVA_HOME is not set` | Run `javahome use <version> --global --shell <shell>` or use a current-session activation command. |
| `javac not found` | You probably selected a JRE instead of a JDK. Install/select a full JDK. |
| `java -version` still shows the old version | Reload the profile, open a new terminal, or ensure the generated `PATH` is active in the current shell. |
| `permission denied` while writing profile | Check the ownership/permissions of your shell profile file. |

More examples are in [`docs/platform-examples.md`](docs/platform-examples.md). Troubleshooting notes are in [`docs/troubleshooting.md`](docs/troubleshooting.md).

## Development

```bash
make test
make build
make build-all
```

## License

MIT
