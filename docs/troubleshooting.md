# Troubleshooting

Start with these commands:

```bash
javahome doctor
javahome list
javahome current
```

They show whether `JAVA_HOME` is set, whether it points to a valid JDK, whether `java` and `javac` exist, and how many installations were discovered.

## No Java installations found

Install a full JDK or move it into a common location scanned by `javahome`.

Common locations include:

- Linux: `/usr/lib/jvm`, `/usr/java`, `/opt/java`, `/opt/jdk*`
- macOS: `/Library/Java/JavaVirtualMachines`
- Windows: `C:\Program Files\Java`, `C:\Program Files\Eclipse Adoptium`, `%USERPROFILE%\.jdks`
- SDKMAN: `~/.sdkman/candidates/java`
- asdf: `~/.asdf/installs/java`
- mise: `~/.local/share/mise/installs/java`

## JAVA_HOME is not set

Temporary fix for Bash:

```bash
eval "$(javahome use 17 --shell bash)"
```

Permanent Bash default:

```bash
javahome use 17 --global --shell bash
source ~/.bashrc
```

Use the correct shell name for your environment: `bash`, `zsh`, `fish`, or `powershell`.

## javac not found

This usually means you selected a JRE instead of a full JDK.

Run:

```bash
javahome current
javahome list
```

Then select a full JDK:

```bash
javahome use 17 --global --shell bash
```

## java -version still shows the old version

Reload your profile or open a new terminal.

For Bash:

```bash
source ~/.bashrc
```

For Zsh:

```zsh
source ~/.zshrc
```

For Fish:

```fish
source ~/.config/fish/config.fish
```

For PowerShell, open a new window or reload your profile.

## Other Java version managers

SDKMAN, jEnv, asdf, and mise can also modify `JAVA_HOME` and `PATH`.

Recommended approach:

- use only one tool to modify Java in a given shell profile
- put `javahome` after other Java tools if you want `javahome` to win
- use `javahome print <version>` in scripts if you only need a path

## Preview profile changes

Always preview global changes first when in doubt:

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
