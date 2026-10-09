//go:build !with_screen

package control

import harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

func showing(*harness.Tree) {}
