package all

// Amazon's native Alexa detector, on a biscuit installed with the Pryon companion. It is cheap to link
// and does nothing until the installer has put the companion there, so every build carries it; it is
// in a file of its own rather than in the list in all.go only to keep that list as upstream has it.
import (
	_ "github.com/ygelfand/echolocal/internal/feature/pryon"
)
