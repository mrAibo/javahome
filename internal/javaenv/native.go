package javaenv

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

func nativeCandidates() map[string]string {
	out := map[string]string{}
	add := func(home, source string) {
		home = strings.TrimSpace(home)
		if home == "" {
			return
		}
		out[home] = source
	}

	switch runtime.GOOS {
	case "darwin":
		for _, home := range macOSJavaHomes() {
			add(home, "macos-java_home")
		}
	case "linux":
		for _, home := range linuxAlternativesJavaHomes() {
			add(home, "linux-alternatives")
		}
	case "windows":
		for _, home := range windowsRegistryJavaHomes() {
			add(home, "windows-registry")
		}
	}
	return out
}

func macOSJavaHomes() []string {
	if runtime.GOOS != "darwin" {
		return nil
	}
	cmd := exec.Command("/usr/libexec/java_home", "-V")
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run()

	combined := stdout.String() + "\n" + stderr.String()
	lines := strings.Split(combined, "\n")
	homes := []string{}
	seen := map[string]bool{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		idx := strings.Index(line, "/")
		if idx < 0 {
			continue
		}
		path := strings.TrimSpace(line[idx:])
		if !strings.Contains(path, ".jdk/Contents/Home") && !strings.HasSuffix(path, "/Contents/Home") {
			continue
		}
		if !seen[path] {
			seen[path] = true
			homes = append(homes, path)
		}
	}
	return homes
}

func linuxAlternativesJavaHomes() []string {
	if runtime.GOOS != "linux" {
		return nil
	}
	seen := map[string]bool{}
	homes := []string{}
	addBinary := func(binary string) {
		binary = strings.TrimSpace(binary)
		if binary == "" || seen[binary] {
			return
		}
		seen[binary] = true
		homes = append(homes, JavaHomeFromBinary(binary))
	}
	addHome := func(home string) {
		home = strings.TrimSpace(home)
		if home == "" || seen[home] {
			return
		}
		seen[home] = true
		homes = append(homes, home)
	}

	if out, err := exec.Command("update-alternatives", "--query", "java").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Value:") {
				addBinary(strings.TrimSpace(strings.TrimPrefix(line, "Value:")))
			}
			if strings.HasPrefix(line, "Alternative:") {
				addBinary(strings.TrimSpace(strings.TrimPrefix(line, "Alternative:")))
			}
		}
	}

	if out, err := exec.Command("update-java-alternatives", "--list").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				addHome(fields[2])
			}
		}
	}

	return homes
}

func windowsRegistryJavaHomes() []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	roots := []string{
		`HKLM\SOFTWARE\JavaSoft\JDK`,
		`HKLM\SOFTWARE\JavaSoft\Java Development Kit`,
		`HKLM\SOFTWARE\Eclipse Adoptium`,
		`HKLM\SOFTWARE\Microsoft\JDK`,
		`HKLM\SOFTWARE\Amazon Corretto`,
		`HKLM\SOFTWARE\WOW6432Node\JavaSoft\JDK`,
	}
	seen := map[string]bool{}
	homes := []string{}
	for _, root := range roots {
		out, err := exec.Command("reg", "query", root, "/s").CombinedOutput()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if !strings.Contains(strings.ToLower(line), "javahome") || !strings.Contains(line, "REG_") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			home := strings.Join(fields[2:], " ")
			if home != "" && !seen[strings.ToLower(home)] {
				seen[strings.ToLower(home)] = true
				homes = append(homes, home)
			}
		}
	}
	return homes
}
