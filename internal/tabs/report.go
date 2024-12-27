package tabs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"lambda-hikes-trailfinder-json-publisher-go-app/internal/helpers" // adjust this import path

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func NewReportTab(window fyne.Window) *container.TabItem {
	// Input fields
	entryType := widget.NewSelect([]string{"Event", "Trip", "Report"}, nil)
	reportDate := widget.NewEntry()
	reportType := widget.NewSelect([]string{"Event", "Trip"}, nil)
	reportName := widget.NewEntry()
	relatedTripURL := widget.NewEntry()
	relatedEventURL := widget.NewEntry()
	uniqueReportID := widget.NewEntry()
	googleMapURL := widget.NewEntry()
	description := widget.NewMultiLineEntry()
	description.Wrapping = fyne.TextWrapWord

	// S3 File Uploads
	mainImagePath := widget.NewEntry()
	mainImageUploadButton := widget.NewButton("Upload Main Image", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			uploadPath := fmt.Sprintf("%s/main.webp", uniqueReportID.Text)
			url, err := helpers.UploadToS3(uploadPath, reader.URI().Path())
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			mainImagePath.SetText(url)
			dialog.ShowInformation("Success", "Main image uploaded successfully", window)
		}, window)
	})

	subImageContainer := container.NewVBox()
	addSubImageButton := widget.NewButton("Add Sub Image", func() {
		subImagePath := widget.NewEntry()
		subImageName := widget.NewEntry()
		subImageDescription := widget.NewMultiLineEntry()
		subImageDescription.Wrapping = fyne.TextWrapWord
		subImageUploadButton := widget.NewButton("Upload Sub Image", func() {
			dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil || reader == nil {
					return
				}
				defer reader.Close()

				imageIndex := len(subImageContainer.Objects) + 1
				uploadPath := fmt.Sprintf("%s/subImages/image%d.webp", uniqueReportID.Text, imageIndex)
				url, err := helpers.UploadToS3(uploadPath, reader.URI().Path())
				if err != nil {
					dialog.ShowError(err, window)
					return
				}

				subImagePath.SetText(url)
				dialog.ShowInformation("Success", "Sub image uploaded successfully", window)
			}, window)
		})

		subImageItem := container.NewVBox(
			widget.NewLabel("Sub Image Description:"), subImageDescription,
			widget.NewLabel("Sub Image Name:"), subImageName,
			widget.NewLabel("Sub Image URL:"), subImagePath,
			subImageUploadButton,
		)

		subImageContainer.Add(subImageItem)
	})

	// Publish button logic
	publishButton := widget.NewButton("Publish", func() {
		// Validate required fields
		if entryType.Selected == "" || reportDate.Text == "" || reportType.Selected == "" || reportName.Text == "" || uniqueReportID.Text == "" {
			dialog.ShowError(fmt.Errorf("Please fill all required fields"), window)
			return
		}

		// Collect data into a JSON object
		reportData := map[string]interface{}{
			"EntryType":       entryType.Selected,
			"ReportDate":      reportDate.Text,
			"ReportType":      reportType.Selected,
			"ReportName":      reportName.Text,
			"RelatedTripURL":  relatedTripURL.Text,
			"RelatedEventURL": relatedEventURL.Text,
			"UniqueReportID":  uniqueReportID.Text,
			"GoogleMapURL":    googleMapURL.Text,
			"MainImagePath":   mainImagePath.Text,
			"Description":     description.Text,
			"SubImages":       helpers.GetSubImageData(subImageContainer),
		}

		// Convert to JSON
		jsonData, err := json.MarshalIndent(reportData, "", "  ")
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		// Save JSON file
		outputFolder := "output"
		if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
			err = os.Mkdir(outputFolder, os.ModePerm)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to create output folder: %v", err), window)
				return
			}
		}

		fileName := filepath.Join(outputFolder, fmt.Sprintf("%s_%s.json", uniqueReportID.Text, reportName.Text))
		err = os.WriteFile(fileName, jsonData, 0644)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		dialog.ShowInformation("Success", fmt.Sprintf("Report saved as %s", fileName), window)
	})

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Entry Type*:"), entryType,
		widget.NewLabel("Report Date*:"), reportDate,
		widget.NewLabel("Report Type*:"), reportType,
		widget.NewLabel("Report Name*:"), reportName,
		widget.NewLabel("Related Trip URL:"), relatedTripURL,
		widget.NewLabel("Related Event URL:"), relatedEventURL,
		widget.NewLabel("Unique Report ID*:"), uniqueReportID,
		widget.NewLabel("Unique Google Map URL:"), googleMapURL,
		widget.NewLabel("Main Image:"), container.NewHBox(mainImagePath, mainImageUploadButton),
		widget.NewLabel("Description:"), description,
		widget.NewLabel("Sub Images:"), subImageContainer, addSubImageButton,
		layout.NewSpacer(),
		publishButton,
	)

	// Make content scrollable
	scrollableContent := container.NewVScroll(content)

	return container.NewTabItem("Report", scrollableContent)
}
