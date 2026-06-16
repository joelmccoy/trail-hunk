package tui

type Screen string

const (
	ScreenStartup     Screen = "startup"
	ScreenOverview    Screen = "overview"
	ScreenWalkthrough Screen = "walkthrough"
	ScreenComments    Screen = "comments"
	ScreenSubmit      Screen = "submit"
)
