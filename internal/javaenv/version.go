package javaenv

import (
	"regexp"
	"strconv"
	"strings"
)

var versionNumberRE = regexp.MustCompile(`\d+(?:\.\d+)*(?:[_+\-][A-Za-z0-9.]+)?`)

// MajorVersion extracts the Java major version from common Java version strings.
// Examples: 1.8.0_392 -> 8, 8 -> 8, 17.0.10 -> 17, 21-ea -> 21.
func MajorVersion(version string) int {
	version = strings.TrimSpace(strings.Trim(version, `"'`))
	if version == "" {
		return 0
	}

	match := versionNumberRE.FindString(version)
	if match == "" {
		return 0
	}

	match = strings.ReplaceAll(match, "_", ".")
	match = strings.Split(match, "+")[0]
	match = strings.Split(match, "-")[0]
	parts := strings.Split(match, ".")
	if len(parts) == 0 {
		return 0
	}

	if parts[0] == "1" && len(parts) > 1 {
		major, err := strconv.Atoi(parts[1])
		if err == nil {
			return major
		}
		return 0
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}
	return major
}

func VersionMatches(version string, major string) bool {
	major = strings.TrimSpace(major)
	if major == "" {
		return true
	}
	parsedMajor := MajorVersion(version)
	if parsedMajor == 0 {
		return false
	}
	requested, err := strconv.Atoi(major)
	if err != nil {
		return strings.EqualFold(version, major)
	}
	return parsedMajor == requested
}
