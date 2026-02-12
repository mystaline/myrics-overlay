package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// TODO: Step 1 - Create a basic Fyne application

	myApp := app.New()
	myWindow := myApp.NewWindow("Myrics Overlay")

	hello := widget.NewLabel("This will become a lyrics overlay.")
	myWindow.SetContent(hello)

	myWindow.ShowAndRun()
}
