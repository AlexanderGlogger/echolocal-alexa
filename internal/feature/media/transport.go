package media

func (p *Player) Transport() Controller {
	if c := p.Holder(); c != nil {
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

func (p *Player) ResetSession() { p.sourceOwner().ResetSession() }
