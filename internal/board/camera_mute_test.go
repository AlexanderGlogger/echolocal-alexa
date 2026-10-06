package board

import "testing"

func TestCameraMuteIsAnExplicitBoardCapability(t *testing.T) {
	if !Checkers.Has(MicMutesCamera) || !Cronos.Has(MicMutesCamera) {
		t.Fatal("Show mute coupling missing")
	}
	if Biscuit.Has(MicMutesCamera) || Rook.Has(MicMutesCamera) {
		t.Fatal("mute coupling applied to unrelated boards")
	}
}
