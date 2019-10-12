package components

import "github.com/jroimartin/gocui"

const ProgressBarName = "progressbar"

type ProgressBar struct {
	component
	playing       bool
	song          string
	artist        string
	album         string
	updatePending bool
}

func NewProgressBar() *ProgressBar {
	p := &ProgressBar{}
	p.name = ProgressBarName
	p.Title = "Status"
	p.Editable = false
	p.Frame = true
	p.Scaling = scalingMax
	p.SizeMin = Point{X: 30, Y: 2}
	p.SizeMax = Point{X: 60, Y: 3}
	p.initialized = true
	p.updatePending = true
	p.updateFunc = p.update
	return p
}

func (p *ProgressBar) AssignKeyBindings(gui *gocui.Gui) error {
	if err := gui.SetKeybinding("", gocui.KeySpace, gocui.ModNone, p.onSpace); err != nil {
		return err
	}
	return nil
}

func (p *ProgressBar) onSpace(g *gocui.Gui, v *gocui.View) error {
	p.playing = !p.playing
	p.updatePending = true
	p.update()
	return nil
}

func (p *ProgressBar) update() {
	p.view.Clear()

	status := ""
	if p.playing {
		status += "⏮ ⏸ ⏯ " + p.song
	} else {
		status += "⏮ ▶ ⏯ " + p.song
	}
	p.view.Write([]byte(status))
	p.updatePending = false
}
