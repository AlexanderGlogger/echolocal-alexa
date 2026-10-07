//go:build board_checkers || board_cronos || board_rook

package settings

import (
	"log/slog"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/shell"
)

func systemPage() *shell.Page { return basicPages().System() }

func keyboardPage() *shell.Page { return basicPages().Keyboard() }

func SetKeyboard(k config.KeyboardSize) {
	if err := config.Set().Screen().Keyboard(k); err != nil {
		slog.Error("the keyboard size could not be saved", "err", err)
	}
	shell.Get().Redraw()
}

func dateTimePage() *shell.Page { return basicPages().DateTime() }
