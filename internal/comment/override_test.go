// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

import "testing"

// snapshotLanguages saves languages[key] and bareNames[key] for extKeys and
// fileKeys (present or not) and registers a t.Cleanup that restores exactly
// that state — so a test calling ApplyOverrides never leaks a mutation into
// any other test in this package.
func snapshotLanguages(t *testing.T, extKeys, fileKeys []string) {
	t.Helper()

	origExt := make(map[string]Language, len(extKeys))
	hadExt := make(map[string]bool, len(extKeys))
	for _, k := range extKeys {
		l, ok := languages[k]
		origExt[k] = l
		hadExt[k] = ok
	}

	origFile := make(map[string]Language, len(fileKeys))
	hadFile := make(map[string]bool, len(fileKeys))
	for _, k := range fileKeys {
		l, ok := bareNames[k]
		origFile[k] = l
		hadFile[k] = ok
	}

	t.Cleanup(func() {
		for _, k := range extKeys {
			if hadExt[k] {
				languages[k] = origExt[k]
			} else {
				delete(languages, k)
			}
		}
		for _, k := range fileKeys {
			if hadFile[k] {
				bareNames[k] = origFile[k]
			} else {
				delete(bareNames, k)
			}
		}
	})
}

func TestApplyOverrides_overwritesKnownExtension(t *testing.T) {
	snapshotLanguages(t, []string{"go"}, nil)

	ApplyOverrides(map[string]Language{"go": CSSStyle}, nil)

	got, ok := LanguageForExtension("go")
	if !ok {
		t.Fatal("LanguageForExtension(go) ok = false, want true")
	}
	if got.Default != CSSStyle.Default {
		t.Errorf("LanguageForExtension(go) = %v, want CSSStyle", got)
	}
}

func TestApplyOverrides_addsUnknownExtension(t *testing.T) {
	snapshotLanguages(t, []string{"custom"}, nil)

	if _, ok := LanguageForExtension("custom"); ok {
		t.Fatal("precondition failed: custom already known")
	}

	ApplyOverrides(map[string]Language{"custom": LispStyle}, nil)

	got, ok := LanguageForExtension("custom")
	if !ok {
		t.Fatal("LanguageForExtension(custom) ok = false, want true")
	}
	if got.Default != LispStyle.Default {
		t.Errorf("LanguageForExtension(custom) = %v, want LispStyle", got)
	}
}

func TestApplyOverrides_extensionKeyNormalization(t *testing.T) {
	snapshotLanguages(t, []string{"go"}, nil)

	ApplyOverrides(map[string]Language{".GO": CSSStyle}, nil)

	got, ok := LanguageForExtension("go")
	if !ok {
		t.Fatal("LanguageForExtension(go) ok = false, want true")
	}
	if got.Default != CSSStyle.Default {
		t.Errorf("LanguageForExtension(go) = %v, want CSSStyle", got)
	}
}

func TestApplyOverrides_overwritesKnownFile(t *testing.T) {
	snapshotLanguages(t, nil, []string{"dockerfile"})

	ApplyOverrides(nil, map[string]Language{"dockerfile": CSSStyle})

	got, ok := LanguageForFile("Dockerfile")
	if !ok {
		t.Fatal("LanguageForFile(Dockerfile) ok = false, want true")
	}
	if got.Default != CSSStyle.Default {
		t.Errorf("LanguageForFile(Dockerfile) = %v, want CSSStyle", got)
	}
}

func TestApplyOverrides_addsUnknownFile(t *testing.T) {
	snapshotLanguages(t, nil, []string{"my-custom-header-file"})

	if _, ok := LanguageForFile("my-custom-header-file"); ok {
		t.Fatal("precondition failed: my-custom-header-file already known")
	}

	ApplyOverrides(nil, map[string]Language{"my-custom-header-file": LispStyle})

	got, ok := LanguageForFile("my-custom-header-file")
	if !ok {
		t.Fatal("LanguageForFile(my-custom-header-file) ok = false, want true")
	}
	if got.Default != LispStyle.Default {
		t.Errorf("LanguageForFile(my-custom-header-file) = %v, want LispStyle", got)
	}
}

func TestApplyOverrides_fileKeyNormalization(t *testing.T) {
	snapshotLanguages(t, nil, []string{"dockerfile"})

	ApplyOverrides(nil, map[string]Language{"DOCKERFILE": CSSStyle})

	got, ok := LanguageForFile("Dockerfile")
	if !ok {
		t.Fatal("LanguageForFile(Dockerfile) ok = false, want true")
	}
	if got.Default != CSSStyle.Default {
		t.Errorf("LanguageForFile(Dockerfile) = %v, want CSSStyle", got)
	}
}

func TestApplyOverrides_doesNotTouchUnrelatedEntries(t *testing.T) {
	snapshotLanguages(t, []string{"go", "py"}, nil)

	ApplyOverrides(map[string]Language{"go": CSSStyle}, nil)

	got, ok := LanguageForExtension("py")
	if !ok {
		t.Fatal("LanguageForExtension(py) ok = false, want true")
	}
	if got.Default != ShellStyle.Default {
		t.Errorf("LanguageForExtension(py) = %v, want unchanged ShellStyle", got)
	}
}

func TestApplyOverrides_nilMapsAreNoop(t *testing.T) {
	ApplyOverrides(nil, nil)
}

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"go", "go"},
		{".go", "go"},
		{"GO", "go"},
		{".GO", "go"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := normalizeKey(tt.in); got != tt.want {
			t.Errorf("normalizeKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
