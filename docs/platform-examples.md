# Platform examples

This document shows common `javahome` workflows on Linux, macOS, and Windows.

## General pattern

Use `javahome list` to see discovered JDKs, `javahome doctor` to diagnose the current setup, and `javahome print <version>` when scripts only need the JDK path.

## Linux

Bash current session:

```bash
javahome list
eval "$(javahome use 17 --shell bash)"
java -version
echo "$JAVA_HOME"
```

Bash permanent default:

```bash
javahome use 17 --global --shell bash
source ~/.bashrc
```

Zsh permanent default:

```zsh
javahome use 21 --global --shell zsh
source ~/.zshrc
```

Fish current session:

```fish
javahome use 17 --shell fish | source
java -version
```

CI or server script:

```bash
JAVA_HOME="$(javahome print 17)"
PATH="$JAVA_HOME/bin:$PATH"
export JAVA_HOME PATH
java -version
```

## macOS

macOS usually uses Zsh by default:

```zsh
javahome list
eval "$(javahome use 17 --shell zsh)"
java -version
```

Permanent Zsh default:

```zsh
javahome use 17 --global --shell zsh
source ~/.zshrc
```

Project requiring Java 21:

```zsh
cd my-project
javahome use 21 --project
cat .javahome.toml
eval "$(javahome use 21 --shell zsh)"
```

## Windows

PowerShell current session:

```powershell
javahome list
javahome use 17 --shell powershell | Invoke-Expression
java -version
$env:JAVA_HOME
```

PowerShell permanent default:

```powershell
javahome use 17 --global --shell powershell
. $PROFILE
```

CMD current session:

```cmd
javahome use 17 --shell cmd
```

Then paste the printed `set JAVA_HOME=...` and `set PATH=...` lines into the same CMD window.

## Vendor filtering

If multiple JDKs with the same major version are installed:

```bash
javahome list
javahome print 17 --vendor temurin
eval "$(javahome use 17 --vendor temurin --shell bash)"
```

## Safe changes

Preview profile changes before writing anything:

```bash
javahome use 17 --global --shell bash --dry-run
```
