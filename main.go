package main

import (
	"lambda-hikes-trailfinder-json-publisher-go-app/internal/tabs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Event and Report Publisher")

	reportTab := tabs.NewReportTab(myWindow)
	eventTab := tabs.NewEventTab(myWindow)
	tripTab := tabs.NewTripTab(myWindow)
	tabs := container.NewAppTabs(reportTab, eventTab, tripTab)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(600, 800))
	myWindow.ShowAndRun()
}
