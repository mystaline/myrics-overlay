package overlay

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Window manages the overlay display
type Window struct {
	fyneWindow fyne.Window
	label      *widget.Label
}

// NewWindow creates a new overlay window
func NewWindow(app fyne.App) *Window {
	// TODO: Step 8 - Create overlay window with Fyne

	w := app.NewWindow("Myrics Overlay")

	label := widget.NewLabel("Waiting for music...")
	content := container.NewCenter(label)

	w.SetContent(content)
	w.Resize(fyne.NewSize(600, 100))

	return &Window{
		fyneWindow: w,
		label:      label,
	}
}

// UpdateLyrics updates the displayed lyrics text
func (w *Window) UpdateLyrics(text string) {
	w.label.SetText(text)
}

// Show displays the window
func (w *Window) Show() {
	w.fyneWindow.Show()
}

// ShowAndRun displays the window and starts the event loop
func (w *Window) ShowAndRun() {
	w.fyneWindow.ShowAndRun()
}
