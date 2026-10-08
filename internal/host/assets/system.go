package assets

import (
	"embed"
	"fmt"
)

//go:embed patch
var patches embed.FS

func Patch(name string) ([]byte, error) {
	b, err := patches.ReadFile("patch/" + name)
	if err != nil {
		return nil, fmt.Errorf("assets: no patch %q", name)
	}
	return b, nil
}
