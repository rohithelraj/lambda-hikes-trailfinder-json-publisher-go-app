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

func NewEventTab(window fyne.Window) *container.TabItem {
	// Input fields with current date
	currentDate := time.Now().Format("2006-01-02")
	creationDate := widget.NewEntry()
	creationDate.SetText(currentDate)
	creationDate.Disable()

	entryType := widget.NewEntry()
	entryType.SetText("Event")
	entryType.Disable()
	eventName := widget.NewEntry()
	eventDate := widget.NewEntry()
	relatedTripURL := widget.NewEntry()
	uniqueEventID := widget.NewEntry()
	uniqueReportURL := widget.NewEntry()
	uniqueKomootURL := widget.NewEntry()
	mainImagePath := widget.NewEntry()

	// Rich text fields with toolbar
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

	// Main image upload
	mainImageUploadButton := widget.NewButton("Upload Main Image", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			uploadPath := fmt.Sprintf("%s/main.webp", uniqueEventID.Text)
			url, err := helpers.UploadToS3(uploadPath, reader.URI().Path())
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			mainImagePath.SetText(url)
			dialog.ShowInformation("Success", "Main image uploaded successfully", window)
		}, window)
	})

	// Sub images container
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
				uploadPath := fmt.Sprintf("%s/subImages/image%d.webp", uniqueEventID.Text, imageIndex)
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
			widget.NewLabel("Sub Image Name:"),
			subImageName,
			widget.NewLabel("Sub Image URL:"),
			subImagePath,
			subImageUploadButton,
		)

		subImageContainer.Add(subImageItem)
	})

	// Publish button
	publishButton := widget.NewButton("Publish", func() {
		// Validate required fields
		if eventName.Text == "" || eventDate.Text == "" || uniqueEventID.Text == "" || uniqueKomootURL.Text == "" || descriptionEntry.Text == "" || transportationEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("Please fill all required fields"), window)
			return
		}

		eventData := map[string]interface{}{
			"CreationDate":    creationDate.Text,
			"EntryType":       entryType.Text,
			"EventName":       eventName.Text,
			"EventDate":       eventDate.Text,
			"RelatedTripURL":  relatedTripURL.Text,
			"UniqueEventID":   uniqueEventID.Text,
			"UniqueReportURL": uniqueReportURL.Text,
			"UniqueKomootURL": uniqueKomootURL.Text,
			"MainImagePath":   mainImagePath.Text,
			"Description":     descriptionEntry.Text,
			"Costs":           costsEntry.Text,
			"Transportation":  transportationEntry.Text,
			"Equipment":       equipmentEntry.Text,
			"SubImages":       helpers.GetSubImageData(subImageContainer),
		}

		jsonData, err := json.MarshalIndent(eventData, "", "  ")
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		outputFolder := "output/events"
		if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
			err = os.Mkdir(outputFolder, os.ModePerm)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to create output folder: %v", err), window)
				return
			}
		}

		fileName := filepath.Join(outputFolder, fmt.Sprintf("%s_event.json", uniqueEventID.Text))
		err = os.WriteFile(fileName, jsonData, 0644)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		dialog.ShowInformation("Success", fmt.Sprintf("Event saved as %s", fileName), window)
	})

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Creation Date*:"), creationDate,
		widget.NewLabel("Entry Type*:"), entryType,
		widget.NewLabel("Event Name*:"), eventName,
		widget.NewLabel("Event Date*:"), eventDate,
		widget.NewLabel("Related Trip URL:"), relatedTripURL,
		widget.NewLabel("Unique Event ID*:"), uniqueEventID,
		widget.NewLabel("Unique Report URL:"), uniqueReportURL,
		widget.NewLabel("Unique Komoot URL*:"), uniqueKomootURL,
		widget.NewLabel("Main Image:"), container.NewHBox(mainImagePath, mainImageUploadButton),
		widget.NewLabel("Description*:"), container.NewBorder(descriptionToolbar, nil, nil, nil, container.NewVBox(descriptionEntry, description)),
		widget.NewLabel("Costs:"), container.NewBorder(costsToolbar, nil, nil, nil, container.NewVBox(costsEntry, costs)),
		widget.NewLabel("Transportation*:"), container.NewBorder(transportationToolbar, nil, nil, nil, container.NewVBox(transportationEntry, transportation)),
		widget.NewLabel("Equipment:"), container.NewBorder(equipmentToolbar, nil, nil, nil, container.NewVBox(equipmentEntry, equipment)),
		widget.NewLabel("Sub Images:"), subImageContainer, addSubImageButton,
		layout.NewSpacer(),
		publishButton,
	)

	scrollableContent := container.NewVScroll(content)
	return container.NewTabItem("Event", scrollableContent)
}
