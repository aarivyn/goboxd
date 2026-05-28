package tests

import (
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
)

func TestConfigLoad(t *testing.T) {
	cfg, err := config.Load("../config/languages.yaml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if len(cfg.Languages) == 0 {
		t.Fatal("no languages loaded")
	}
}

func TestFindLanguage(t *testing.T) {
	cfg, _ := config.Load("../config/languages.yaml")
	lang := cfg.FindLanguage("py3")
	if lang == nil {
		t.Fatal("py3 not found")
	}
	if lang.Name != "Python 3" {
		t.Fatalf("expected Python 3, got %s", lang.Name)
	}
}

func TestFindLanguageNotFound(t *testing.T) {
	cfg, _ := config.Load("../config/languages.yaml")
	lang := cfg.FindLanguage("nonexistent")
	if lang != nil {
		t.Fatal("expected nil for unknown language")
	}
}

func TestValidateFilename(t *testing.T) {
	cases := []struct {
		name  string
		valid bool
	}{
		{"solution.py", true},
		{"solution.cpp", true},
		{"../../etc/passwd", false},
		{".hidden", false},
		{"/absolute/path", false},
		{"", true},
	}

	for _, c := range cases {
		result := validateFilename(c.name)
		if result != c.valid {
			t.Errorf("validateFilename(%q) = %v, want %v", c.name, result, c.valid)
		}
	}
}

func validateFilename(name string) bool {
	if name == "" {
		return true
	}
	for _, c := range name {
		if c == '/' || c == '\\' {
			return false
		}
	}
	if name[0] == '.' {
		return false
	}
	if len(name) > 255 {
		return false
	}
	return true
}

func TestFlagAllowlist(t *testing.T) {
	allowlist := []string{"-O0", "-O1", "-O2", "-O3", "-Wall", "-std=*"}

	allowed := []string{"-O2", "-Wall", "-std=c++17"}
	for _, f := range allowed {
		if !isFlagAllowed(f, allowlist) {
			t.Errorf("flag %q should be allowed", f)
		}
	}

	blocked := []string{"-fplugin=evil.so", "--specs=evil", "-Wl,evil"}
	for _, f := range blocked {
		if isFlagAllowed(f, allowlist) {
			t.Errorf("flag %q should be blocked", f)
		}
	}
}

func isFlagAllowed(flag string, allowlist []string) bool {
	for _, pattern := range allowlist {
		if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
			if len(flag) >= len(pattern)-1 &&
				flag[:len(pattern)-1] == pattern[:len(pattern)-1] {
				return true
			}
		} else if flag == pattern {
			return true
		}
	}
	return false
}