package main

import (
	"github.com/ygelfand/libcountertop/pkg/runtime/payload"

	"github.com/ygelfand/echolocal/internal/cli/echod"
)

func main() {
	payload.Enter()
	echod.Execute()
}
