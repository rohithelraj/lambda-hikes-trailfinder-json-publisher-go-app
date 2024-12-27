package tabs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"lambda-hikes-trailfinder-json-publisher-go-app/internal/helpers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewTripTab(window fyne.Window) *container.TabItem {
	currentDate := time.Now().Format("2006-01-02")
	creationDate := widget.NewEntry()
	creationDate.SetText(currentDate)
	creationDate.Disable()

	entryType := widget.NewEntry()
	entryType.SetText("Trip")
	entryType.Disable()
	tripName := widget.NewEntry()
	tripStartDate := widget.NewEntry()
	tripEndDate := widget.NewEntry()
	uniqueTripID := widget.NewEntry()
	uniqueGoogleMapURL := widget.NewEntry()
	uniqueReportURL := widget.NewEntry()
	mainImagePath := widget.NewEntry()

	createRichTextArea := func(label string) (*widget.RichText, *widget.Entry, *widget.Toolbar) {
		richText := widget.NewRichText()
		richText.Resize(fyne.NewSize(0, 150))
		richText.Wrapping = fyne.TextWrapWord
		binding := binding.NewString()
		entry := widget.NewMultiLineEntry()
		entry.Wrapping = fyne.TextWrapWord
		entry.Resize(fyne.NewSize(0, 100))
		entry.OnChanged = func(text string) {
			richText.ParseMarkdown(text)
			binding.Set(text)
		}

		toolbar := widget.NewToolbar(
			widget.NewToolbarAction(theme.GridIcon(), func() {
				currentText := entry.Text
				cursorPos := entry.CursorColumn
				newText := currentText[:cursorPos] + "**bold**" + currentText[cursorPos:]
				entry.SetText(newText)
				entry.CursorColumn = cursorPos + 6
			}),
			widget.NewToolbarAction(theme.MenuIcon(), func() {
				currentText := entry.Text
				cursorPos := entry.CursorColumn
				newText := currentText[:cursorPos] + "*italic*" + currentText[cursorPos:]
				entry.SetText(newText)
				entry.CursorColumn = cursorPos + 6
			}),
			widget.NewToolbarAction(theme.ListIcon(), func() {
				currentText := entry.Text
				cursorPos := entry.CursorColumn
				newText := currentText[:cursorPos] + "# Heading" + currentText[cursorPos:]
				entry.SetText(newText)
				entry.CursorColumn = cursorPos + 9
			}),
			widget.NewToolbarAction(theme.MailAttachmentIcon(), func() {
				currentText := entry.Text
				cursorPos := entry.CursorColumn

				linkTextEntry := widget.NewEntry()
				linkURLEntry := widget.NewEntry()
				linkTextEntry.SetPlaceHolder("Link text")
				linkURLEntry.SetPlaceHolder("URL")

				content := container.NewVBox(
					widget.NewLabel("Link Text:"), linkTextEntry,
					widget.NewLabel("URL:"), linkURLEntry,
				)

				dialog.ShowCustomConfirm("Insert Link", "Insert", "Cancel", content, func(insert bool) {
					if insert && linkTextEntry.Text != "" && linkURLEntry.Text != "" {
						linkMD := fmt.Sprintf("[%s](%s)", linkTextEntry.Text, linkURLEntry.Text)
						newText := currentText[:cursorPos] + linkMD + currentText[cursorPos:]
						entry.SetText(newText)
						entry.CursorColumn = cursorPos + len(linkMD)
					}
				}, window)
			}),
		)

		return richText, entry, toolbar
	}

	description, descriptionEntry, descriptionToolbar := createRichTextArea("Description")
	costs, costsEntry, costsToolbar := createRichTextArea("Costs")
	transportation, transportationEntry, transportationToolbar := createRichTextArea("Transportation")
	equipment, equipmentEntry, equipmentToolbar := createRichTextArea("Equipment")
	accommodation, accommodationEntry, accommodationToolbar := createRichTextArea("Accommodation")

	mainImageUploadButton := widget.NewButton("Upload Main Image", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			uploadPath := fmt.Sprintf("%s/main.webp", uniqueTripID.Text)
			url, err := helpers.UploadToS3(uploadPath, reader.URI().Path())
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			mainImagePath.SetText(url)
			dialog.ShowInformation("Success", "Main image uploaded successfully", window)
		}, window)
	})

	relatedEventsContainer := container.NewVBox()
	addEventButton := widget.NewButton("Add Related Event", func() {
		eventName := widget.NewEntry()
		eventDescription := widget.NewMultiLineEntry()
		eventURL := widget.NewEntry()

		eventItem := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Related Event %d", len(relatedEventsContainer.Objects)+1)),
			widget.NewLabel("Event Name:"), eventName,
			widget.NewLabel("Event Description:"), eventDescription,
			widget.NewLabel("Event URL:"), eventURL,
		)

		relatedEventsContainer.Add(eventItem)
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
				uploadPath := fmt.Sprintf("%s/subImages/image%d.webp", uniqueTripID.Text, imageIndex)
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
			widget.NewLabel(fmt.Sprintf("Sub Image %d Description:", len(subImageContainer.Objects)+1)),
			subImageDescription,
			widget.NewLabel("Sub Image Name:"), subImageName,
			widget.NewLabel("Sub Image URL:"), subImagePath,
			subImageUploadButton,
		)

		subImageContainer.Add(subImageItem)
	})

	getRelatedEventsData := func() []map[string]string {
		var events []map[string]string
		for _, obj := range relatedEventsContainer.Objects {
			if eventItem, ok := obj.(*fyne.Container); ok {
				eventData := make(map[string]string)
				for i, component := range eventItem.Objects {
					if entry, ok := component.(*widget.Entry); ok {
						if entry.MultiLine {
							eventData["Description"] = entry.Text
						} else {
							if i > 0 {
								if label, ok := eventItem.Objects[i-1].(*widget.Label); ok {
									switch label.Text {
									case "Event Name:":
										eventData["Name"] = entry.Text
									case "Event URL:":
										eventData["URL"] = entry.Text
									}
								}
							}
						}
					}
				}
				if len(eventData) > 0 {
					events = append(events, eventData)
				}
			}
		}
		return events
	}

	publishButton := widget.NewButton("Publish", func() {
		// Validate required fields
		if tripName.Text == "" || tripStartDate.Text == "" || tripEndDate.Text == "" || uniqueTripID.Text == "" || descriptionEntry.Text == "" || transportationEntry.Text == "" || accommodationEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("Please fill all required fields"), window)
			return
		}
		tripData := map[string]interface{}{
			"CreationDate":       creationDate.Text,
			"EntryType":          entryType.Text,
			"TripName":           tripName.Text,
			"TripStartDate":      tripStartDate.Text,
			"TripEndDate":        tripEndDate.Text,
			"UniqueTripID":       uniqueTripID.Text,
			"UniqueGoogleMapURL": uniqueGoogleMapURL.Text,
			"UniqueReportURL":    uniqueReportURL.Text,
			"MainImagePath":      mainImagePath.Text,
			"Description":        descriptionEntry.Text,
			"Costs":              costsEntry.Text,
			"Transportation":     transportationEntry.Text,
			"Equipment":          equipmentEntry.Text,
			"Accommodation":      accommodationEntry.Text,
			"RelatedEvents":      getRelatedEventsData(),
			"SubImages":          helpers.GetSubImageData(subImageContainer),
		}

		jsonData, err := json.MarshalIndent(tripData, "", "  ")
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		outputFolder := "output/trips"
		if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
			err = os.MkdirAll(outputFolder, os.ModePerm)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to create output folder: %v", err), window)
				return
			}
		}

		fileName := filepath.Join(outputFolder, fmt.Sprintf("%s_trip.json", uniqueTripID.Text))
		err = os.WriteFile(fileName, jsonData, 0644)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		dialog.ShowInformation("Success", fmt.Sprintf("Trip saved as %s", fileName), window)
	})

	content := container.NewVBox(
		widget.NewLabel("Creation Date*:"), creationDate,
		widget.NewLabel("Entry Type*:"), entryType,
		widget.NewLabel("Trip Name*:"), tripName,
		widget.NewLabel("Trip Start Date*:"), tripStartDate,
		widget.NewLabel("Trip End Date*:"), tripEndDate,
		widget.NewLabel("Unique Trip ID*:"), uniqueTripID,
		widget.NewLabel("Unique Google Map URL:"), uniqueGoogleMapURL,
		widget.NewLabel("Unique Report URL:"), uniqueReportURL,
		widget.NewLabel("Main Image:"), container.NewHBox(mainImagePath, mainImageUploadButton),
		widget.NewLabel("Description*:"), container.NewBorder(descriptionToolbar, nil, nil, nil, container.NewVBox(descriptionEntry, description)),
		widget.NewLabel("Costs:"), container.NewBorder(costsToolbar, nil, nil, nil, container.NewVBox(costsEntry, costs)),
		widget.NewLabel("Transportation*:"), container.NewBorder(transportationToolbar, nil, nil, nil, container.NewVBox(transportationEntry, transportation)),
		widget.NewLabel("Equipment:"), container.NewBorder(equipmentToolbar, nil, nil, nil, container.NewVBox(equipmentEntry, equipment)),
		widget.NewLabel("Accommodation*:"), container.NewBorder(accommodationToolbar, nil, nil, nil, container.NewVBox(accommodationEntry, accommodation)),
		widget.NewLabel("Related Events:"), relatedEventsContainer, addEventButton,
		widget.NewLabel("Sub Images:"), subImageContainer, addSubImageButton,
		layout.NewSpacer(),
		publishButton,
	)

	scrollableContent := container.NewVScroll(content)
	return container.NewTabItem("Trip", scrollableContent)
}
