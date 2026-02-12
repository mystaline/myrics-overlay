package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// TODO: Step 1 - Create a basic Fyne application

	myApp := app.New()
	myWindow := myApp.NewWindow("Myrics Overlay")

	myWindow.SetFixedSize(true)
	bg := canvas.NewRectangle(color.RGBA{0, 0, 0, 200})

	label := widget.NewLabel("This will become a lyrics overlay.")
	label.Alignment = fyne.TextAlignCenter
	label.TextStyle = fyne.TextStyle{
		Bold: true,
	}

	content := container.NewStack(
		bg, container.NewCenter(label),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 100))
	myWindow.CenterOnScreen()

	myWindow.ShowAndRun()
}
