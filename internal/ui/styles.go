// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package ui

import "github.com/charmbracelet/lipgloss"

// Chrome colors — structural UI elements such as command names, headers, and
// descriptive text. These carry no pass/fail meaning.
var (
	colorPrimary     = lipgloss.AdaptiveColor{Light: "#00695C", Dark: "#26A69A"}
	colorSecondary   = lipgloss.AdaptiveColor{Light: "#4527A0", Dark: "#7E57C2"}
	colorDescription = lipgloss.AdaptiveColor{Light: "#455A64", Dark: "#90A4AE"}
	colorFlag        = lipgloss.AdaptiveColor{Light: "#00838F", Dark: "#26C6DA"}
)

// Semantic colors — operation results only. Each color carries a specific
// meaning and must not be repurposed for structural elements.
var (
	colorSuccess = lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#66BB6A"}
	colorError   = lipgloss.AdaptiveColor{Light: "#C62828", Dark: "#EF5350"}
	colorWarning = lipgloss.AdaptiveColor{Light: "#E65100", Dark: "#FFA726"}
	colorInfo    = lipgloss.AdaptiveColor{Light: "#1565C0", Dark: "#42A5F5"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#616161", Dark: "#9E9E9E"}
)

// Chrome styles for structural UI elements.
var (
	StylePrimary     = lipgloss.NewStyle().Foreground(colorPrimary)
	StyleSecondary   = lipgloss.NewStyle().Foreground(colorSecondary)
	StyleDescription = lipgloss.NewStyle().Foreground(colorDescription)
	StyleFlag        = lipgloss.NewStyle().Foreground(colorFlag)
)

// Semantic styles for operation results.
var (
	StyleSuccess = lipgloss.NewStyle().Foreground(colorSuccess)
	StyleError   = lipgloss.NewStyle().Foreground(colorError)
	StyleWarning = lipgloss.NewStyle().Foreground(colorWarning)
	StyleInfo    = lipgloss.NewStyle().Foreground(colorInfo)
	StyleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
)
