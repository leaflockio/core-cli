// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package lang

import "testing"

// --- Language.String ---

func TestString_knownLanguage(t *testing.T) {
	if got := Go.String(); got != "go" {
		t.Errorf("expected %q, got %q", "go", got)
	}
}

func TestString_unknownLanguage(t *testing.T) {
	if got := Unknown.String(); got != "unknown" {
		t.Errorf("expected %q, got %q", "unknown", got)
	}
}

func TestString_outOfRangeLanguage(t *testing.T) {
	if got := Language(9999).String(); got != "unknown" {
		t.Errorf("expected %q for out-of-range language, got %q", "unknown", got)
	}
}

// --- FromExtension ---

func TestFromExtension_empty(t *testing.T) {
	if got := FromExtension(""); got != Unknown {
		t.Errorf("expected Unknown for empty extension, got %v", got)
	}
}

func TestFromExtension_withLeadingDot(t *testing.T) {
	if got := FromExtension(".go"); got != Go {
		t.Errorf("expected Go for .go, got %v", got)
	}
}

func TestFromExtension_withoutLeadingDot(t *testing.T) {
	if got := FromExtension("go"); got != Go {
		t.Errorf("expected Go for go, got %v", got)
	}
}

func TestFromExtension_caseInsensitive(t *testing.T) {
	if got := FromExtension(".GO"); got != Go {
		t.Errorf("expected Go for .GO, got %v", got)
	}
}

func TestFromExtension_unknownExtension(t *testing.T) {
	if got := FromExtension(".xyz"); got != Unknown {
		t.Errorf("expected Unknown for .xyz, got %v", got)
	}
}

func TestFromExtension_typescript(t *testing.T) {
	if got := FromExtension(".ts"); got != TypeScript {
		t.Errorf("expected TypeScript for .ts, got %v", got)
	}
}

func TestFromExtension_tsx(t *testing.T) {
	if got := FromExtension(".tsx"); got != TypeScript {
		t.Errorf("expected TypeScript for .tsx, got %v", got)
	}
}

func TestFromExtension_javascript(t *testing.T) {
	if got := FromExtension(".js"); got != JavaScript {
		t.Errorf("expected JavaScript for .js, got %v", got)
	}
}

func TestFromExtension_python(t *testing.T) {
	if got := FromExtension(".py"); got != Python {
		t.Errorf("expected Python for .py, got %v", got)
	}
}

func TestFromExtension_shell(t *testing.T) {
	if got := FromExtension(".sh"); got != Shell {
		t.Errorf("expected Shell for .sh, got %v", got)
	}
}

func TestFromExtension_rust(t *testing.T) {
	if got := FromExtension(".rs"); got != Rust {
		t.Errorf("expected Rust for .rs, got %v", got)
	}
}

func TestFromExtension_java(t *testing.T) {
	if got := FromExtension(".java"); got != Java {
		t.Errorf("expected Java for .java, got %v", got)
	}
}

func TestFromExtension_kotlin(t *testing.T) {
	if got := FromExtension(".kt"); got != Kotlin {
		t.Errorf("expected Kotlin for .kt, got %v", got)
	}
}

func TestFromExtension_swift(t *testing.T) {
	if got := FromExtension(".swift"); got != Swift {
		t.Errorf("expected Swift for .swift, got %v", got)
	}
}

func TestFromExtension_ruby(t *testing.T) {
	if got := FromExtension(".rb"); got != Ruby {
		t.Errorf("expected Ruby for .rb, got %v", got)
	}
}

func TestFromExtension_css(t *testing.T) {
	if got := FromExtension(".css"); got != CSS {
		t.Errorf("expected CSS for .css, got %v", got)
	}
}

func TestFromExtension_html(t *testing.T) {
	if got := FromExtension(".html"); got != HTML {
		t.Errorf("expected HTML for .html, got %v", got)
	}
}

func TestFromExtension_proto(t *testing.T) {
	if got := FromExtension(".proto"); got != Proto {
		t.Errorf("expected Proto for .proto, got %v", got)
	}
}

func TestFromExtension_terraform(t *testing.T) {
	if got := FromExtension(".tf"); got != Terraform {
		t.Errorf("expected Terraform for .tf, got %v", got)
	}
}

func TestFromExtension_cpp(t *testing.T) {
	if got := FromExtension(".cpp"); got != CPP {
		t.Errorf("expected CPP for .cpp, got %v", got)
	}
}

func TestFromExtension_c(t *testing.T) {
	if got := FromExtension(".c"); got != C {
		t.Errorf("expected C for .c, got %v", got)
	}
}

func TestFromExtension_csharp(t *testing.T) {
	if got := FromExtension(".cs"); got != CSharp {
		t.Errorf("expected CSharp for .cs, got %v", got)
	}
}

func TestFromExtension_vue(t *testing.T) {
	if got := FromExtension(".vue"); got != Vue {
		t.Errorf("expected Vue for .vue, got %v", got)
	}
}

func TestFromExtension_svelte(t *testing.T) {
	if got := FromExtension(".svelte"); got != Svelte {
		t.Errorf("expected Svelte for .svelte, got %v", got)
	}
}

func TestFromExtension_scala(t *testing.T) {
	if got := FromExtension(".scala"); got != Scala {
		t.Errorf("expected Scala for .scala, got %v", got)
	}
}

func TestFromExtension_elixir(t *testing.T) {
	if got := FromExtension(".ex"); got != Elixir {
		t.Errorf("expected Elixir for .ex, got %v", got)
	}
}

func TestFromExtension_haskell(t *testing.T) {
	if got := FromExtension(".hs"); got != Haskell {
		t.Errorf("expected Haskell for .hs, got %v", got)
	}
}

func TestFromExtension_lua(t *testing.T) {
	if got := FromExtension(".lua"); got != Lua {
		t.Errorf("expected Lua for .lua, got %v", got)
	}
}

func TestFromExtension_dart(t *testing.T) {
	if got := FromExtension(".dart"); got != Dart {
		t.Errorf("expected Dart for .dart, got %v", got)
	}
}
