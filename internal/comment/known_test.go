// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

import (
	"reflect"
	"testing"
)

// wantLanguage describes the expected shape of a named Language: whether
// each form is present, its wrap output for a representative line, and
// which form Default resolves to.
type wantLanguage struct {
	hasLine, hasBlock   bool
	lineWant, blockWant []string
	defaultIsLine       bool
}

func checkLanguage(t *testing.T, name string, l Language, w wantLanguage) {
	t.Helper()

	if w.hasLine {
		if l.Line == nil {
			t.Errorf("%s: Line is nil, want a Style", name)
		} else if got := Wrap([]string{"hello"}, l.Line); !reflect.DeepEqual(got, w.lineWant) {
			t.Errorf("%s: Line wrap = %v, want %v", name, got, w.lineWant)
		}
	} else if l.Line != nil {
		t.Errorf("%s: Line = %v, want nil", name, l.Line)
	}

	if w.hasBlock {
		if l.Block == nil {
			t.Errorf("%s: Block is nil, want a Style", name)
		} else if got := Wrap([]string{"hello"}, l.Block); !reflect.DeepEqual(got, w.blockWant) {
			t.Errorf("%s: Block wrap = %v, want %v", name, got, w.blockWant)
		}
	} else if l.Block != nil {
		t.Errorf("%s: Block = %v, want nil", name, l.Block)
	}

	if l.Default == nil {
		t.Fatalf("%s: Default is nil, want a Style", name)
	}
	if w.defaultIsLine {
		if l.Default != l.Line {
			t.Errorf("%s: Default != Line, want Default to be Line", name)
		}
	} else {
		if l.Default != l.Block {
			t.Errorf("%s: Default != Block, want Default to be Block", name)
		}
	}
}

func TestKnownStyles(t *testing.T) {
	tests := []struct {
		name string
		lang Language
		want wantLanguage
	}{
		{
			name: "CStyle",
			lang: CStyle,
			want: wantLanguage{
				hasLine: true, lineWant: []string{"// hello"},
				hasBlock: true, blockWant: []string{"/*", "hello", "*/"},
				defaultIsLine: true,
			},
		},
		{
			name: "ShellStyle",
			lang: ShellStyle,
			want: wantLanguage{
				hasLine: true, lineWant: []string{"# hello"},
				defaultIsLine: true,
			},
		},
		{
			name: "CSSStyle",
			lang: CSSStyle,
			want: wantLanguage{
				hasBlock: true, blockWant: []string{"/*", " * hello", " */"},
				defaultIsLine: false,
			},
		},
		{
			name: "SassStyle",
			lang: SassStyle,
			want: wantLanguage{
				hasLine: true, lineWant: []string{"// hello"},
				hasBlock: true, blockWant: []string{"/*", " * hello", " */"},
				defaultIsLine: false,
			},
		},
		{
			name: "HTMLStyle",
			lang: HTMLStyle,
			want: wantLanguage{
				hasBlock: true, blockWant: []string{"<!--", "hello", "-->"},
				defaultIsLine: false,
			},
		},
		{
			name: "JavaStyle",
			lang: JavaStyle,
			want: wantLanguage{
				hasLine: true, lineWant: []string{"// hello"},
				hasBlock: true, blockWant: []string{"/*", " * hello", " */"},
				defaultIsLine: false,
			},
		},
		{
			name: "DashStyle",
			lang: DashStyle,
			want: wantLanguage{
				hasLine: true, lineWant: []string{"-- hello"},
				defaultIsLine: true,
			},
		},
		{
			name: "LispStyle",
			lang: LispStyle,
			want: wantLanguage{
				hasLine: true, lineWant: []string{";; hello"},
				defaultIsLine: true,
			},
		},
		{
			name: "ErlangStyle",
			lang: ErlangStyle,
			want: wantLanguage{
				hasLine: true, lineWant: []string{"% hello"},
				defaultIsLine: true,
			},
		},
		{
			name: "VimStyle",
			lang: VimStyle,
			want: wantLanguage{
				hasLine: true, lineWant: []string{`" hello`},
				defaultIsLine: true,
			},
		},
		{
			name: "JinjaStyle",
			lang: JinjaStyle,
			want: wantLanguage{
				hasBlock: true, blockWant: []string{"{#", "hello", "#}"},
				defaultIsLine: false,
			},
		},
		{
			name: "OCamlStyle",
			lang: OCamlStyle,
			want: wantLanguage{
				hasBlock: true, blockWant: []string{"(**", "   hello", "*)"},
				defaultIsLine: false,
			},
		},
		{
			name: "PowerShellStyle",
			lang: PowerShellStyle,
			want: wantLanguage{
				hasLine: true, lineWant: []string{"# hello"},
				hasBlock: true, blockWant: []string{"<#", " hello", "#>"},
				defaultIsLine: false,
			},
		},
		{
			name: "HandlebarsStyle",
			lang: HandlebarsStyle,
			want: wantLanguage{
				hasBlock: true, blockWant: []string{"{{!--", "  hello", "--}}"},
				defaultIsLine: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkLanguage(t, tt.name, tt.lang, tt.want)
		})
	}
}
