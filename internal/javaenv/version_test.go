package javaenv

import "testing"

func TestMajorVersion(t *testing.T) {
	tests := map[string]int{
		"1.8.0_392":       8,
		"8":               8,
		"11.0.22":         11,
		"17.0.10+7":       17,
		"21-ea":           21,
		`"23.0.1"`:        23,
		"openjdk 17.0.10": 17,
		"":                0,
	}

	for input, want := range tests {
		got := MajorVersion(input)
		if got != want {
			t.Fatalf("MajorVersion(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestVersionMatches(t *testing.T) {
	if !VersionMatches("1.8.0_392", "8") {
		t.Fatal("expected Java 8 match")
	}
	if !VersionMatches("17.0.10", "17") {
		t.Fatal("expected Java 17 match")
	}
	if VersionMatches("17.0.10", "21") {
		t.Fatal("did not expect Java 21 match")
	}
}
