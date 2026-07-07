# javahome

`javahome` is a small, dependency-free Java home switcher written in Go.

It discovers installed JDKs, shows the active `JAVA_HOME`, emits shell-specific activation snippets, updates shell profile files on request, and provides a simple diagnostic command.

## Why this exists

Many small `set_java_home` scripts are fragile because they assume one fixed Java path layout. `javahome` is designed to be more universal and safer:

- works as a single binary
- supports Linux, macOS, and Windows path conventions
- detects Java installations from common locations
- validates `JAVA_HOME/bin/java`
- removes stale Java `bin` entries from `PATH`
- does not silently edit shell files unless `--global` is used
- has `--dry-run` and `doctor` commands
- uses only the Go standard library

## Install from source

```bash
go install github.com/mrAibo/javahome/cmd/javahome@latest
```

Or build locally:

```bash
git clone https://github.com/mrAibo/javahome.git
cd javahome
go build ./cmd/javahome
```

## Usage

List discovered Java installations:

```bash
javahome list
```

Show the active Java home:

```bash
javahome current
```

Print the Java home for a major version:

```bash
javahome print 17
```

Activate Java 17 in the current Bash shell:

```bash
eval "$(javahome use 17 --shell bash)"
```

Activate Java 17 in the current Zsh shell:

```bash
eval "$(javahome use 17 --shell zsh)"
```

Activate Java 17 in Fish:

```fish
javahome use 17 --shell fish | source
```

Activate Java 17 in PowerShell:

```powershell
javahome use 17 --shell powershell | Invoke-Expression
```

Make Java 17 permanent for Bash:

```bash
javahome use 17 --global --shell bash
```

Create a project-local config file:

```bash
javahome use 17 --project
```

Run diagnostics:

```bash
javahome doctor
```

## Shell helper

Install a small helper function named `jhome`:

```bash
javahome init bash --global
```

Then use:

```bash
jhome 17
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

## Development

```bash
go test ./...
go build ./cmd/javahome
```

## License

MIT
