# UX roadmap

Ideas that would make `javahome` even more user-friendly.

## Implemented in 0.3.0

- `javahome select` numbered interactive selection
- `javahome activate` for `.javahome.toml`
- Bash/Zsh/Fish/PowerShell completion generation
- native discovery hooks for macOS, Linux, and Windows
- GitHub release workflow for tagged releases
- SDKMAN/jEnv/asdf/mise conflict hints in `doctor`

## Remaining high-impact improvements

1. Arrow-key interactive selection

   Upgrade `javahome select` from numbered prompts to true arrow-key navigation while keeping the numbered fallback.

2. Better native discovery

   Add deeper integration with platform-native mechanisms:

   - macOS `/usr/libexec/java_home`
   - Linux `update-alternatives` and `update-java-alternatives`
   - Windows Registry entries for JDK vendors

3. Safer profile backup

   Before `--global` modifies a profile, create a timestamped backup such as `.bashrc.javahome-backup-YYYYMMDD-HHMMSS`.

4. Better conflict remediation

   Add exact suggestions when SDKMAN, jEnv, asdf, or mise appear after `javahome` in profile files.

## Nice-to-have improvements

- `javahome uninstall` to remove marked profile blocks
- `javahome env` to print detailed environment diagnostics
- `javahome where java` to explain which `java` binary wins in `PATH`
- optional compact output mode for small terminals
- optional `--plain` flag in addition to `NO_COLOR=1`
- support for custom scan paths via a config file
