# UX roadmap

Ideas that would make `javahome` even more user-friendly.

## High-impact improvements

1. Interactive selection

   Add `javahome select` to show discovered JDKs and let the user choose one with arrow keys. This would be useful for humans, while the existing commands remain best for scripts.

2. Shell completions

   Generate completions for Bash, Zsh, Fish, and PowerShell. This would make commands and flags easier to discover.

3. Release workflow

   Add a GitHub release workflow that builds binaries for Linux, macOS, and Windows whenever a tag such as `v0.2.0` is pushed.

4. Better native discovery

   Add deeper integration with platform-native mechanisms:

   - macOS `/usr/libexec/java_home`
   - Linux `update-alternatives` and `update-java-alternatives`
   - Windows Registry entries for JDK vendors

5. Safer profile backup

   Before `--global` modifies a profile, create a timestamped backup such as `.bashrc.javahome-backup-YYYYMMDD-HHMMSS`.

6. Apply project config

   Add `javahome activate` to read `.javahome.toml` and emit the correct shell activation snippet.

7. Better conflict detection

   Make `javahome doctor` detect SDKMAN, jEnv, asdf, and mise when they appear to override `JAVA_HOME` or `PATH` after `javahome`.

## Nice-to-have improvements

- `javahome uninstall` to remove marked profile blocks
- `javahome env` to print detailed environment diagnostics
- `javahome where java` to explain which `java` binary wins in `PATH`
- optional compact output mode for small terminals
- optional `--plain` flag in addition to `NO_COLOR=1`
- support for custom scan paths via a config file
