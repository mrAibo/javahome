package javaenv

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type ManagerDetection struct {
	Name    string `json:"name"`
	Found   bool   `json:"found"`
	Active  bool   `json:"active"`
	Message string `json:"message"`
}

func DetectManagers(pathValue string) []ManagerDetection {
	home, _ := os.UserHomeDir()
	sep := ":"
	if runtime.GOOS == "windows" {
		sep = ";"
	}
	parts := strings.Split(pathValue, sep)
	contains := func(fragment string) bool {
		fragment = normalizeManagerPath(fragment)
		for _, part := range parts {
			if strings.Contains(normalizeManagerPath(part), fragment) {
				return true
			}
		}
		return false
	}
	exists := func(path string) bool {
		if path == "" {
			return false
		}
		_, err := os.Stat(path)
		return err == nil
	}
	inHome := func(parts ...string) string {
		if home == "" {
			return ""
		}
		return filepath.Join(append([]string{home}, parts...)...)
	}

	items := []ManagerDetection{}

	sdkmanDir := firstEnv("SDKMAN_DIR", inHome(".sdkman"))
	sdkmanFound := os.Getenv("SDKMAN_DIR") != "" || exists(sdkmanDir)
	sdkmanActive := contains(filepath.Join(".sdkman", "candidates", "java"))
	items = append(items, ManagerDetection{"SDKMAN", sdkmanFound, sdkmanActive, managerMessage("SDKMAN", sdkmanFound, sdkmanActive)})

	jenvDir := firstEnv("JENV_ROOT", inHome(".jenv"))
	jenvFound := os.Getenv("JENV_ROOT") != "" || exists(jenvDir)
	jenvActive := contains(filepath.Join(".jenv", "shims")) || contains(filepath.Join(".jenv", "bin"))
	items = append(items, ManagerDetection{"jEnv", jenvFound, jenvActive, managerMessage("jEnv", jenvFound, jenvActive)})

	asdfDir := firstEnv("ASDF_DIR", inHome(".asdf"))
	asdfFound := os.Getenv("ASDF_DIR") != "" || exists(asdfDir)
	asdfActive := contains(filepath.Join(".asdf", "shims")) || contains(filepath.Join("asdf", "shims"))
	items = append(items, ManagerDetection{"asdf", asdfFound, asdfActive, managerMessage("asdf", asdfFound, asdfActive)})

	miseDir := firstEnv("MISE_DATA_DIR", inHome(".local", "share", "mise"))
	miseFound := os.Getenv("MISE_DATA_DIR") != "" || exists(miseDir)
	miseActive := contains(filepath.Join("mise", "shims")) || contains(filepath.Join(".local", "share", "mise"))
	items = append(items, ManagerDetection{"mise", miseFound, miseActive, managerMessage("mise", miseFound, miseActive)})

	return items
}

func managerMessage(name string, found bool, active bool) string {
	if active {
		return name + " appears in PATH and may override JAVA_HOME/PATH order"
	}
	if found {
		return name + " found, but not active in PATH"
	}
	return name + " not detected"
}

func firstEnv(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func normalizeManagerPath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	return strings.ToLower(path)
}
