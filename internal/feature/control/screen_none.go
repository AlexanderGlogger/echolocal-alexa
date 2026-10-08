//go:build !(board_checkers || board_cronos || board_rook)

package control

import harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

func showing(*harness.Tree) {}
