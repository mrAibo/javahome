package shellenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanPathForJavaPrependsAndDedupes(t *testing.T) {
	tmp := t.TempDir()
	javaHome := filepath.Join(tmp, "jdk-17")
	javaBin := filepath.Join(javaHome, "bin")
	if err := os.MkdirAll(javaBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(javaHome, "release"), []byte(`JAVA_VERSION="17.0.10"`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(javaBin, "java"), []byte(""), 0o755); err != nil {
		t.Fatal(err)
	}

	sep := PathSeparator()
	input := strings.Join([]string{"/usr/bin", javaBin, "/usr/local/bin", "/usr/bin"}, sep)
	got := CleanPathForJava(input, javaHome)
	parts := strings.Split(got, sep)

	if parts[0] != javaBin {
		t.Fatalf("first PATH element = %q, want %q", parts[0], javaBin)
	}
	count := 0
	for _, part := range parts {
		if part == "/usr/bin" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected /usr/bin once, got %d in %q", count, got)
	}
}
