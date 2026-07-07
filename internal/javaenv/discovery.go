package javaenv

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func IsJavaHome(home string) bool {
	if strings.TrimSpace(home) == "" {
		return false
	}
	bin := filepath.Join(home, "bin", javaExecutableName())
	stat, err := os.Stat(bin)
	return err == nil && !stat.IsDir()
}

func javaExecutableName() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

func JavaHomeFromBinary(binary string) string {
	if binary == "" {
		return ""
	}
	resolved := binary
	if abs, err := filepath.Abs(binary); err == nil {
		resolved = abs
	}
	if eval, err := filepath.EvalSymlinks(resolved); err == nil {
		resolved = eval
	}
	return filepath.Clean(filepath.Dir(filepath.Dir(resolved)))
}

func Discover() []Installation {
	currentHome := os.Getenv("JAVA_HOME")
	candidates := map[string]string{}

	add := func(path string, source string) {
		if path == "" {
			return
		}
		path = expandPlatformEnv(path)
		path = filepath.Clean(path)
		if existing, ok := candidates[path]; ok {
			if !strings.Contains(existing, source) {
				candidates[path] = existing + "," + source
			}
			return
		}
		candidates[path] = source
	}

	add(currentHome, "JAVA_HOME")

	if javaPath, err := exec.LookPath(javaExecutableName()); err == nil {
		add(JavaHomeFromBinary(javaPath), "PATH")
	}

	for _, pattern := range discoveryPatterns() {
		matches, _ := filepath.Glob(expandPlatformEnv(pattern))
		for _, match := range matches {
			add(normalizeCandidate(match), "scan")
		}
	}

	installs := make([]Installation, 0, len(candidates))
	seen := map[string]bool{}
	for path, source := range candidates {
		inst, ok := InstallationFromHome(path, source, currentHome)
		if !ok {
			continue
		}
		key := filepath.Clean(inst.Path)
		if filepath.Separator == '\\' {
			key = strings.ToLower(key)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		installs = append(installs, inst)
	}

	sort.SliceStable(installs, func(i, j int) bool {
		if installs[i].Major != installs[j].Major {
			return installs[i].Major < installs[j].Major
		}
		if installs[i].Vendor != installs[j].Vendor {
			return installs[i].Vendor < installs[j].Vendor
		}
		return installs[i].Path < installs[j].Path
	})
	return installs
}

func normalizeCandidate(path string) string {
	path = filepath.Clean(path)
	if strings.HasSuffix(path, filepath.Join("Contents", "Home")) {
		return path
	}
	if strings.HasSuffix(path, ".jdk") && runtime.GOOS == "darwin" {
		return filepath.Join(path, "Contents", "Home")
	}
	return path
}

func expandPlatformEnv(value string) string {
	value = os.ExpandEnv(value)
	if runtime.GOOS != "windows" {
		return value
	}

	replacements := map[string]string{
		"%ProgramFiles%":      os.Getenv("ProgramFiles"),
		"%ProgramFiles(x86)%": os.Getenv("ProgramFiles(x86)"),
		"%USERPROFILE%":       os.Getenv("USERPROFILE"),
	}
	for marker, replacement := range replacements {
		if replacement != "" {
			value = strings.ReplaceAll(value, marker, replacement)
		}
	}
	return value
}

func discoveryPatterns() []string {
	home, _ := os.UserHomeDir()
	patterns := []string{}

	switch runtime.GOOS {
	case "darwin":
		patterns = append(patterns,
			"/Library/Java/JavaVirtualMachines/*.jdk/Contents/Home",
			"/System/Library/Java/JavaVirtualMachines/*.jdk/Contents/Home",
			"/opt/homebrew/opt/openjdk*/libexec/openjdk.jdk/Contents/Home",
			"/usr/local/opt/openjdk*/libexec/openjdk.jdk/Contents/Home",
			"/opt/homebrew/Cellar/openjdk*/*/libexec/openjdk.jdk/Contents/Home",
			"/usr/local/Cellar/openjdk*/*/libexec/openjdk.jdk/Contents/Home",
		)
	case "windows":
		envPatterns := []string{
			`%ProgramFiles%\Java\*`,
			`%ProgramFiles%\Eclipse Adoptium\*`,
			`%ProgramFiles%\Microsoft\jdk-*`,
			`%ProgramFiles%\Amazon Corretto\*`,
			`%ProgramFiles(x86)%\Java\*`,
			`%USERPROFILE%\.jdks\*`,
		}
		patterns = append(patterns, envPatterns...)
	default:
		patterns = append(patterns,
			"/usr/lib/jvm/*",
			"/usr/java/*",
			"/opt/java/*",
			"/opt/jdk*",
			"/opt/*jdk*",
			"/opt/*java*",
		)
	}

	if home != "" {
		patterns = append(patterns,
			filepath.Join(home, ".sdkman", "candidates", "java", "*"),
			filepath.Join(home, ".asdf", "installs", "java", "*"),
			filepath.Join(home, ".local", "share", "mise", "installs", "java", "*"),
		)
	}

	return patterns
}

func Select(installs []Installation, version string, vendor string) (Installation, bool) {
	version = strings.TrimSpace(version)
	vendor = strings.TrimSpace(strings.ToLower(vendor))

	var candidates []Installation
	for _, inst := range installs {
		if version != "" && !VersionMatches(inst.Version, version) {
			continue
		}
		if vendor != "" && !strings.Contains(strings.ToLower(inst.Vendor), vendor) && !strings.Contains(strings.ToLower(inst.Path), vendor) {
			continue
		}
		candidates = append(candidates, inst)
	}

	if len(candidates) == 0 {
		return Installation{}, false
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].IsCurrent != candidates[j].IsCurrent {
			return candidates[i].IsCurrent
		}
		return candidates[i].Path < candidates[j].Path
	})
	return candidates[0], true
}
