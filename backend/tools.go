//go:build tools
// +build tools

// Package tools imports CLIs used by the project so they appear in go.mod
package tools

import (
	_ "github.com/swaggo/swag/cmd/swag"
)
