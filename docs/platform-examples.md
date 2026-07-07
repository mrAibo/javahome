# Platform examples

This document shows common `javahome` workflows on Linux, macOS, and Windows.

The key rule is simple:

- use `javahome use <version> --shell <shell>` for the current terminal session
- use `javahome use <version> --global --shell <shell>` for future terminal sessions
- use `javahome print <version>` in scripts when you only need the JDK path

## Linux

### Bash: current terminal only

```bash
javahome list
eval "$(javahome use 17 --shell bash)"
echo "$JAVA_HOME"
java -version
```

This changes only the current Bash process.

### Bash: permanent default

```bash
javahome use 17 --global --shell bash
source ~/.bashrc
java -version
```

`javahome` writes a marked block to `~/.bashrc`. You can preview it first:

```bash
javahome use 17 --global --shell bash --dry-run
```

### Zsh: permanent default

```zsh
javahome use 21 --global --shell zsh
source ~/.zshrc
java -version
```

### Fish: current terminal only

```fish
javahome use 17 --shell fish | source
echo $JAVA_HOME
java -version
```

### Fish: permanent default

```fish
javahome use 17 --global --shell fish
source ~/.config/fish/config.fish
```

### Linux CI job

```bash
JAVA_HOME="$(javahome print 17)"
PATH="$JAVA_HOME/bin:$PATH"
export JAVA_HOME PATH
java -version
mvn test
```

This avoids profile files and is usually better for CI/CD.

## macOS

### Zsh current session

```zsh
javahome list
eval "$(javahome use 17 --shell zsh)"
java -version
```

### Zsh permanent default

```zsh
javahome use 17 --global --shell zsh
source ~/.zshrc
```

### Bash on macOS

```bash
javahome use 17 --global --shell bash
source ~/.bashrc
```

### Project requiring Java 21

```zsh
cd my-project
javahome use 21 --project
cat .javahome.toml
eval "$(javahome use 21 --shell zsh)"
```

## Windows

### PowerShell current session

```powershell
javahome list
javahome use 17 --shell powershell | Invoke-Expression
$env:JAVA_HOME
java -version
```

### PowerShell permanent default

```powershell
javahome use 17 --global --shell powershell
. $PROFILE
```

If your PowerShell profile is not loaded because of execution policy restrictions, open a new PowerShell window or adjust your execution policy according to your organization's rules.

### CMD current session

CMD does not have a safe built-in equivalent to Bash `eval` or PowerShell `Invoke-Expression` for this use case.

Print the commands:

```cmd
javahome use 17 --shell cmd
```

Then paste the generated lines into the same CMD window:

```cmd
set JAVA_HOME=C:\Path\To\JDK
set PATH=C:\Path\To\JDK\bin;...
```

### Windows automation

In PowerShell scripts, prefer:

```powershell
$jdk = javahome print 17
$env:JAVA_HOME = $jdk
$env:Path = "$jdk\bin;$env:Path"
java -version
```

## Vendor filtering

If multiple JDKs with the same major version are installed, filter by vendor text:

```bash
javahome list
javahome print 17 --vendor temurin
eval "$(javahome use 17 --vendor temurin --shell bash)"
```

The filter is text-based and case-insensitive.

## Safe changes

Before writing anything to a profile file, preview the change:

```bash
javahome use 17 --global --shell bash --dry-run
```

The global profile update uses a marked block:

```text
# >>> javahome >>>
...
# <<< javahome <<<
```

That makes later updates safer because `javahome` can replace its own block instead of blindly appending new lines.
