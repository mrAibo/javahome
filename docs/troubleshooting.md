# Troubleshooting

Run these commands first:

```bash
javahome doctor
javahome list
javahome current
```

They tell you whether `JAVA_HOME` is set, whether it points to a valid JDK, whether `java` and `javac` exist, and how many installations were discovered.

## `No Java installations found`

`javahome` could not find a JDK in known locations.

Check whether Java exists at all:

```bash
java -version
javac -version
```

Then check common locations:

Linux:

```bash
ls -la /usr/lib/jvm /usr/java /opt/java /opt 2>/dev/null
```

macOS:

```bash
ls -la /Library/Java/JavaVirtualMachines 2>/dev/null
```

Windows PowerShell:

```powershell
Get-ChildItem 'C:\Program Files\Java' -ErrorAction SilentlyContinue
Get-ChildItem 'C:\Program Files\Eclipse Adoptium' -ErrorAction SilentlyContinue
Get-ChildItem "$env:USERPROFILE\.jdks" -ErrorAction SilentlyContinue
```

## `JAVA_HOME is not set`

Temporary fix for the current shell:

```bash
eval "$(javahome use 17 --shell bash)"
```

Permanent fix for future Bash shells:

```bash
javahome use 17 --global --shell bash
source ~/.bashrc
```

Use the correct shell name for your environment: `bash`, `zsh`, `fish`, or `powershell`.

## `javac not found`

This usually means you selected a JRE instead of a full JDK.

Check:

```bash
javahome current
javac -version
```

Install or select a full JDK, then run:

```bash
javahome list
javahome use 17 --global --shell bash
```

## `java -version` still shows the old version

The current terminal probably did not load the generated environment yet.

For current shell activation:

```bash
eval "$(javahome use 17 --shell bash)"
```

For permanent activation:

```bash
source ~/.bashrc
```

or open a new terminal window.

Also check that no other version manager is rewriting `PATH` after `javahome`, such as SDKMAN, jEnv, asdf, or mise.

## I use SDKMAN, jEnv, asdf, or mise

`javahome` can still discover installations from common SDKMAN/asdf/mise locations, but it does not try to replace those tools.

Recommended approach:

- use SDKMAN/jEnv/asdf/mise when you want their full ecosystem
- use `javahome` when you want a small, dependency-free binary that prints or activates a JDK path
- avoid having multiple tools modify `JAVA_HOME` in the same shell profile unless you know the order

## PowerShell profile did not reload

Run:

```powershell
$PROFILE
Test-Path $PROFILE
. $PROFILE
```

If execution policy blocks profile loading, open a new PowerShell window or follow your organization's policy for PowerShell profile execution.

## CMD limitations

CMD cannot safely evaluate multi-line command output like Bash or PowerShell.

Use:

```cmd
javahome use 17 --shell cmd
```

Then paste the printed `set JAVA_HOME=...` and `set PATH=...` lines into the same CMD window.

For scripts, prefer PowerShell.

## Disable or force color output

Colors are automatic in supported terminals and disabled for redirected output.

Disable colors:

```bash
NO_COLOR=1 javahome doctor
JAVAHOME_COLOR=never javahome list
```

Force colors, useful for demos or screenshots:

```bash
JAVAHOME_COLOR=always javahome doctor
```

Shell activation output is never colorized when generated with `--shell`, because it must remain safe for `eval`, `source`, or `Invoke-Expression`.

## Profile file looks wrong

Preview before writing:

```bash
javahome use 17 --global --shell bash --dry-run
```

`javahome` writes a marked block:

```text
# >>> javahome >>>
...
# <<< javahome <<<
```

You can remove that block manually if needed.
