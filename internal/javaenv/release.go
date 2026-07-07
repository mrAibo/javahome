package javaenv

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Installation struct {
	Path      string `json:"path"`
	Version   string `json:"version"`
	Major     int    `json:"major"`
	Vendor    string `json:"vendor"`
	Source    string `json:"source"`
	IsCurrent bool   `json:"is_current"`
}

func ReadReleaseFile(home string) map[string]string {
	values := map[string]string{}
	f, err := os.Open(filepath.Join(home, "release"))
	if err != nil {
		return values
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"`)
		values[key] = val
	}
	return values
}

func InstallationFromHome(home string, source string, currentHome string) (Installation, bool) {
	home = filepath.Clean(home)
	if !IsJavaHome(home) {
		return Installation{}, false
	}

	rel := ReadReleaseFile(home)
	version := rel["JAVA_VERSION"]
	vendor := firstNonEmpty(rel["IMPLEMENTOR"], rel["JAVA_VENDOR"], rel["IMPLEMENTOR_VERSION"], "Unknown")

	inst := Installation{
		Path:      home,
		Version:   version,
		Major:     MajorVersion(version),
		Vendor:    vendor,
		Source:    source,
		IsCurrent: currentHome != "" && samePath(home, currentHome),
	}
	return inst, true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func samePath(a string, b string) bool {
	ac, errA := filepath.Abs(filepath.Clean(a))
	bc, errB := filepath.Abs(filepath.Clean(b))
	if errA == nil {
		a = ac
	}
	if errB == nil {
		b = bc
	}
	if filepath.Separator == '\\' {
		return strings.EqualFold(a, b)
	}
	return a == b
}
