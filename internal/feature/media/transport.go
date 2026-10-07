package media

// Transport is whoever the buttons reach: whoever holds the card, or this player when nothing does.
func Transport() Controller {
	if c := Get().Holder(); c != nil {
		return c
	}
	return queue{}
}

type queue struct{}

func (queue) Play()     { Get().stream.Unpause() }
func (queue) Pause()    { Get().stream.Pause() }
func (queue) Stop()     { Get().Stop() }
func (queue) Next()     { Get().Next() }
func (queue) Previous() { Get().Previous() }
