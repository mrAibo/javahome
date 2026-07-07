package shellenv

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func PathSeparator() string {
	if runtime.GOOS == "windows" {
		return ";"
	}
	return ":"
}

func CleanPathForJava(currentPath string, javaHome string) string {
	sep := PathSeparator()
	parts := strings.Split(currentPath, sep)
	javaBin := filepath.Clean(filepath.Join(javaHome, "bin"))

	out := []string{javaBin}
	seen := map[string]bool{pathKey(javaBin): true}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		clean := filepath.Clean(part)
		key := pathKey(clean)
		if seen[key] {
			continue
		}
		if isLikelyJavaBin(clean) {
			continue
		}
		seen[key] = true
		out = append(out, clean)
	}
	return strings.Join(out, sep)
}

func isLikelyJavaBin(path string) bool {
	javaName := "java"
	if runtime.GOOS == "windows" {
		javaName = "java.exe"
	}
	if _, err := os.Stat(filepath.Join(path, javaName)); err != nil {
		return false
	}
	home := filepath.Dir(path)
	if _, err := os.Stat(filepath.Join(home, "release")); err == nil {
		return true
	}
	parent := filepath.Base(path)
	return strings.EqualFold(parent, "bin") && strings.Contains(strings.ToLower(home), "java")
}

func pathKey(path string) string {
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}
