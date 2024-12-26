package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type SubImage struct {
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Report struct {
	ReportName  string     `json:"reportName"`
	EntryType   string     `json:"entryType"`
	ReportDate  string     `json:"reportDate"`
	ReportType  string     `json:"reportType"`
	TripURL     string     `json:"tripUrl"`
	EventURL    string     `json:"eventUrl"`
	ReportID    string     `json:"reportId"`
	MapURL      string     `json:"mapUrl"`
	MainImage   string     `json:"mainImage"`
	Description string     `json:"description"`
	SubImages   []SubImage `json:"subImages"`
}

func createReportTab() fyne.CanvasObject {
	// Bindings for form fields
	reportName := binding.NewString()
	entryType := binding.NewString()
	reportDate := binding.NewString()
	reportType := binding.NewString()
	tripURL := binding.NewString()
	eventURL := binding.NewString()
	reportID := binding.NewString()
	mapURL := binding.NewString()
	mainImage := binding.NewString()
	description := binding.NewString()

	// Create form widgets
	reportNameEntry := widget.NewEntryWithData(reportName)
	reportNameEntry.Validator = isRequired

	entrySelect := widget.NewSelect([]string{"Event", "Trip", "Report"}, func(value string) {
		entryType.Set(value)
	})

	datePickerBtn := widget.NewButton(time.Now().Format("2006-01-02"), func() {
		showDatePicker(reportDate)
	})

	reportTypeSelect := widget.NewSelect([]string{"Event", "Trip"}, func(value string) {
		reportType.Set(value)
	})

	reportIDEntry := widget.NewEntryWithData(reportID)
	reportIDEntry.Validator = isRequired

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Report Name*", Widget: reportNameEntry},
			{Text: "Entry Type", Widget: entrySelect},
			{Text: "Report Date*", Widget: datePickerBtn},
			{Text: "Report Type*", Widget: reportTypeSelect},
			{Text: "Related Trip URL", Widget: widget.NewEntryWithData(tripURL)},
			{Text: "Related Event URL", Widget: widget.NewEntryWithData(eventURL)},
			{Text: "Unique Report ID*", Widget: reportIDEntry},
			{Text: "Unique Google Map URL", Widget: widget.NewEntryWithData(mapURL)},
			{Text: "Main Image", Widget: createImageUploader(mainImage, "main.webp")},
			{Text: "Description", Widget: widget.NewMultiLineEntry()},
		},
		OnSubmit: func() {
			report := Report{
				ReportName:  getValue(reportName),
				EntryType:   entrySelect.Selected,
				ReportDate:  datePickerBtn.Text,
				ReportType:  reportTypeSelect.Selected,
				TripURL:     getValue(tripURL),
				EventURL:    getValue(eventURL),
				ReportID:    getValue(reportID),
				MapURL:      getValue(mapURL),
				MainImage:   getValue(mainImage),
				Description: description.Get(),
				SubImages:   getSubImages(),
			}
			saveReport(report)
		},
	}

	// Sub images container
	subImagesContainer := container.NewVBox()
	subImagesScroll := container.NewVScroll(subImagesContainer)
	subImagesScroll.SetMinSize(fyne.NewSize(600, 200))

	addSubImageBtn := widget.NewButton("Add Sub Image", func() {
		addSubImageFields(subImagesContainer)
	})

	// Main container with scroll
	mainContent := container.NewVBox(
		form,
		addSubImageBtn,
		subImagesScroll,
		widget.NewButton("Publish", func() {
			form.Submit()
		}),
	)

	return container.NewVScroll(mainContent)
}

type SubImageWidget struct {
	Container *fyne.Container
	URL       binding.String
	Name      binding.String
	Desc      binding.String
}

var subImageWidgets []SubImageWidget

func addSubImageFields(container *fyne.Container) {
	url := binding.NewString()
	name := binding.NewString()
	desc := binding.NewString()

	urlEntry := widget.NewEntryWithData(url)
	nameEntry := widget.NewEntryWithData(name)
	descEntry := widget.NewMultiLineEntry()

	removeBtn := widget.NewButton("Remove", nil)

	subImageBox := container.NewHBox(
		container.NewVBox(
			widget.NewLabel("Sub Image URL:"),
			urlEntry,
			widget.NewLabel("Name:"),
			nameEntry,
			widget.NewLabel("Description:"),
			descEntry,
		),
		removeBtn,
	)

	widget := SubImageWidget{
		Container: subImageBox,
		URL:       url,
		Name:      name,
		Desc:      desc,
	}

	removeBtn.OnTapped = func() {
		container.Remove(subImageBox)
		for i, w := range subImageWidgets {
			if w.Container == subImageBox {
				subImageWidgets = append(subImageWidgets[:i], subImageWidgets[i+1:]...)
				break
			}
		}
	}

	subImageWidgets = append(subImageWidgets, widget)
	container.Add(subImageBox)
}

func getSubImages() []SubImage {
	var subImages []SubImage
	for _, widget := range subImageWidgets {
		url, _ := widget.URL.Get()
		name, _ := widget.Name.Get()
		desc, _ := widget.Desc.Get()

		subImages = append(subImages, SubImage{
			URL:         url,
			Name:        name,
			Description: desc,
		})
	}
	return subImages
}

func isRequired(value string) error {
	if value == "" {
		return fmt.Errorf("This field is required")
	}
	return nil
}

func saveReport(report Report) {
	jsonData, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	// Create output directory if it doesn't exist
	err = os.MkdirAll("output", 0755)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	// Save file as reportId_reportName.json
	filename := fmt.Sprintf("%s_%s.json", report.ReportID, report.ReportName)
	filepath := filepath.Join("output", filename)

	err = ioutil.WriteFile(filepath, jsonData, 0644)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	dialog.ShowInformation("Success", "Report saved successfully", window)
}
func showDatePicker(dateBinding binding.String) {
	// Create entries for day, month, year
	yearEntry := widget.NewEntry()
	monthEntry := widget.NewEntry()
	dayEntry := widget.NewEntry()

	// Pre-fill with current date
	now := time.Now()
	yearEntry.SetText(fmt.Sprintf("%d", now.Year()))
	monthEntry.SetText(fmt.Sprintf("%02d", now.Month()))
	dayEntry.SetText(fmt.Sprintf("%02d", now.Day()))

	dateForm := widget.NewForm(
		&widget.FormItem{Text: "Year", Widget: yearEntry},
		&widget.FormItem{Text: "Month (1-12)", Widget: monthEntry},
		&widget.FormItem{Text: "Day", Widget: dayEntry},
	)

	dialog.ShowCustomConfirm("Select Date", "Confirm", "Cancel", dateForm, func(confirm bool) {
		if !confirm {
			return
		}

		year := yearEntry.Text
		month := monthEntry.Text
		day := dayEntry.Text

		dateStr := fmt.Sprintf("%s-%s-%s", year, month, day)
		dateBinding.Set(dateStr)
	}, window)
}

func makeDatePicker(dateBinding binding.String) fyne.CanvasObject {
	calendar := widget.NewTextGrid()
	// Basic date picker implementation
	// You might want to use a third-party date picker widget for better functionality
	return calendar
}

func getValue(b binding.String) string {
	val, err := b.Get()
	if err != nil {
		return ""
	}
	return val
}

func createImageUploader(binding binding.String, filename string) fyne.CanvasObject {
	upload := widget.NewButton("Upload Image", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if reader == nil {
				return
			}
			// Store filename for now, S3 upload to be implemented
			binding.Set(reader.URI().Name())
		}, window)
		fd.Show()
	})
	return upload
}

var window fyne.Window

func main() {
	myApp := app.New()
	window = myApp.NewWindow("Reports")

	tabs := container.NewAppTabs(
		container.NewTabItem("Report", createReportTab()),
	)

	window.SetContent(tabs)
	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
}
