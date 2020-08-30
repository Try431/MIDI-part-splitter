package main

import (
	"fmt"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
)

func main() {
	myApp := app.New()
	w := myApp.NewWindow("Choose File for MIDI Conversion")

	// gradient := canvas.NewHorizontalGradient(color.White, color.Transparent)
	// gradient := canvas.NewRadialGradient(color.White, color.Transparent)
	// w.SetContent(gradient)

	// w.Resize(fyne.NewSize(800, 800))

	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err == nil && reader == nil {
			return
		}
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		fmt.Println("File opened")
	}, w)

	fd.Show()
	w.Resize(fyne.NewSize(2000, 1000))
	fd.Resize(fyne.NewSize(2000, 1000))

	w.ShowAndRun()
}
