package internal

import (
	"context"
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"time"
)

type UI struct {
	status           *tview.TextView
	hero             *tview.TextView
	app              *tview.Application
	logCh            chan LogMessage
	roundCh          chan Round
	latestCancelFunc func()
	soundPlayer      *SoundPlayer
}

type LogMessage struct {
	Severity LogSeverity
	Message  string
}

type LogSeverity = string

const (
	LogSeverityInfo LogSeverity = "[white][INFO[][white] %s\n"
	LogSeverityFail LogSeverity = "[red][FAIL[][white] %s\n"
	LogSeverityOK   LogSeverity = "[green][ OK [][white] %s\n"
)

func NewUI(logCh chan LogMessage, roundCh chan Round, player *SoundPlayer) *UI {
	app := tview.NewApplication()
	status := tview.NewTextView().SetDynamicColors(true).
		SetWordWrap(true)
	status.SetChangedFunc(func() {
		app.Draw()
		status.ScrollToEnd()
	})
	status.SetBorder(true)
	status.SetScrollable(true)
	status.SetTitle("status")
	status.SetBorderPadding(1, 1, 1, 1)

	hero := tview.NewTextView().SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
			status.ScrollToEnd()
		})
	hero.SetBorder(true)
	hero.SetTitle("stop afk idiot")
	hero.SetBorderPadding(10, 0, 2, 0)
	hero.SetChangedFunc(func() {
		app.Draw()
	})

	return &UI{
		status:      status,
		hero:        hero,
		app:         app,
		logCh:       logCh,
		roundCh:     roundCh,
		soundPlayer: player,
	}
}

func (u *UI) Start() {
	flex := tview.NewFlex().
		AddItem(u.status, 0, 1, false).
		AddItem(u.hero, 0, 3, false)

	if err := u.app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (u *UI) ProcessChannels() {
	for {
		select {
		case log := <-u.logCh:
			u.status.Write([]byte(fmt.Sprintf(log.Severity, log.Message)))
		case round := <-u.roundCh:
			u.setHero(round)
		}
	}
}

func (u *UI) setHero(round Round) {
	if u.latestCancelFunc != nil {
		u.latestCancelFunc()
		u.latestCancelFunc = nil
	}

	switch round.Phase {
	case "freezetime":
		ctx, cancel := context.WithCancel(context.Background())
		u.latestCancelFunc = cancel
		go u.freezeTimer(ctx)
		go u.soundPlayer.PlaySound("buy")
	case "live":
		ctx, cancel := context.WithCancel(context.Background())
		u.latestCancelFunc = cancel
		go u.liveTimer(ctx)
		go u.soundPlayer.PlaySound("start")
	case "over":
		switch round.WinTeam {
		case "T":
			u.hero.SetBackgroundColor(tcell.ColorDarkOrange)
			u.setHeroFigure("T WIN")
		case "CT":
			u.hero.SetBackgroundColor(tcell.ColorBlue)
			u.setHeroFigure("CT WIN")
		default:
			u.hero.SetBackgroundColor(tcell.ColorDefault)
			u.hero.Clear()
		}
	default:
		u.hero.SetBackgroundColor(tcell.ColorDefault)
		u.hero.Clear()
	}

}

func (u *UI) freezeTimer(ctx context.Context) {
	const maxSeconds = 15

	u.hero.SetBackgroundColor(tcell.ColorDefault)
	u.setHeroFigure(fmt.Sprintf("BUY 0:%02d", maxSeconds))

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for sec := 1; sec <= maxSeconds; sec++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.setHeroFigure(fmt.Sprintf("BUY 0:%02d", maxSeconds-sec))
		}
	}
}

func (u *UI) liveTimer(ctx context.Context) {
	const maxTicks = 6

	u.setHeroFigure("ROUND LIVE")

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var colors = []tcell.Color{
		tcell.ColorDefault,
		tcell.ColorRed,
	}

	for sec := 1; sec <= maxTicks; sec++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.app.QueueUpdateDraw(func() {
				u.hero.SetBackgroundColor(colors[sec%2])
			})
		}
	}
}

func (u *UI) setHeroFigure(text string) {
	fig := figure.NewFigure(text, "", true)
	u.hero.SetText(fig.String())
}
