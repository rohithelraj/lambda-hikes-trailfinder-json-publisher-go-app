package tabs

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"lambda-hikes-trailfinder-json-publisher-go-app/internal/helpers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewReportTab(window fyne.Window) *container.TabItem {
	// Input fields
	entryType := widget.NewEntry()
	entryType.SetText("Report")
	entryType.Disable()

	reportDate := widget.NewEntry()
	reportType := widget.NewSelect([]string{"Event", "Trip"}, nil)
	reportName := widget.NewEntry()
	relatedTripURL := widget.NewEntry()
	relatedEventURL := widget.NewEntry()
	uniqueReportID := widget.NewEntry()
	googleMapURL := widget.NewEntry()

	// Rich text description
	description := widget.NewRichText()
	descriptionBinding := binding.NewString()
	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.OnChanged = func(text string) {
		description.ParseMarkdown(text)
		descriptionBinding.Set(text)
	}

	// Toolbar with supported formatting
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			currentText := descriptionEntry.Text
			descriptionEntry.SetText(currentText + "**bold**")
		}),
		widget.NewToolbarAction(theme.DocumentIcon(), func() {
			currentText := descriptionEntry.Text
			descriptionEntry.SetText(currentText + "*italic*")
		}),
		widget.NewToolbarAction(theme.DocumentIcon(), func() {
			currentText := descriptionEntry.Text
			descriptionEntry.SetText(currentText + "# Heading")
		}),
	)

	descriptionContainer := container.NewBorder(toolbar, nil, nil, nil, container.NewVBox(descriptionEntry, description))

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
		if reportDate.Text == "" || reportType.Selected == "" || reportName.Text == "" || uniqueReportID.Text == "" {
			dialog.ShowError(fmt.Errorf("Please fill all required fields"), window)
			return
		}

		reportData := map[string]interface{}{
			"EntryType":       entryType.Text,
			"ReportDate":      reportDate.Text,
			"ReportType":      reportType.Selected,
			"ReportName":      reportName.Text,
			"RelatedTripURL":  relatedTripURL.Text,
			"RelatedEventURL": relatedEventURL.Text,
			"UniqueReportID":  uniqueReportID.Text,
			"GoogleMapURL":    googleMapURL.Text,
			"MainImagePath":   mainImagePath.Text,
			"Description":     descriptionEntry.Text,
			"SubImages":       helpers.GetSubImageData(subImageContainer),
		}

		jsonData, err := json.MarshalIndent(reportData, "", "  ")
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

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
		widget.NewLabel("Description:"), descriptionContainer,
		widget.NewLabel("Sub Images:"), subImageContainer, addSubImageButton,
		layout.NewSpacer(),
		publishButton,
	)

	scrollableContent := container.NewVScroll(content)
	return container.NewTabItem("Report", scrollableContent)
}

func richTextToHTML(segments []widget.RichTextSegment) string {
	var html strings.Builder

	for _, seg := range segments {
		if textSeg, ok := seg.(*widget.TextSegment); ok {
			text := textSeg.Text

			if textSeg.Style.ColorName != "" || textSeg.Style.SizeName != "" || textSeg.Style.TextStyle != (fyne.TextStyle{}) {
				html.WriteString("<span style=\"")
				if textSeg.Style.ColorName != "" {
					html.WriteString(fmt.Sprintf("color:%s;", textSeg.Style.ColorName))
				}
				switch textSeg.Style.SizeName {
				case "h1":
					html.WriteString("font-size:2em;")
				case "h2":
					html.WriteString("font-size:1.5em;")
				case "h3":
					html.WriteString("font-size:1.17em;")
				case "large":
					html.WriteString("font-size:1.2em;")
				case "small":
					html.WriteString("font-size:0.8em;")
				}
				if textSeg.Style.TextStyle.Bold {
					html.WriteString("font-weight:bold;")
				}
				if textSeg.Style.TextStyle.Italic {
					html.WriteString("font-style:italic;")
				}
				html.WriteString("\">")
				html.WriteString(text)
				html.WriteString("</span>")
			} else {
				html.WriteString(text)
			}
		}
	}

	return html.String()
}

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}
