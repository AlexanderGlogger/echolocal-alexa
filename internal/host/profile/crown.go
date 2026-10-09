package profile

import (
	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/host/bootimg"
	"github.com/ygelfand/echolocal/internal/host/sysimg"
)

var crown = Profile{
	Board:  board.Crown,
	Boot:   bootimg.Crown,
	System: sysimg.Crown,

	Disable: checkers.Disable,
	Enable:  checkers.Enable,
}
