package main

import (
	"lambda-hikes-trailfinder-json-publisher-go-app/internal/tabs" // adjust this import path based on your module name

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Report Tab Example")

	reportTab := tabs.NewReportTab(myWindow)
	tabs := container.NewAppTabs(reportTab)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(600, 800))
	myWindow.ShowAndRun()
}
