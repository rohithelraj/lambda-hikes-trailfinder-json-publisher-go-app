package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	awsRegion  = "us-east-1"                        // Replace with your AWS region
	bucketName = "hikes-trailfinder-website-images" // Replace with your S3 bucket name
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Report Tab Example")

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

			// Upload to S3
			uploadPath := fmt.Sprintf("%s/main.webp", uniqueReportID.Text)
			url, err := uploadToS3(uploadPath, reader.URI().Path())
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			mainImagePath.SetText(url)
			dialog.ShowInformation("Success", "Main image uploaded successfully", myWindow)
		}, myWindow)
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

				// Upload to S3
				imageIndex := len(subImageContainer.Objects) + 1
				uploadPath := fmt.Sprintf("%s/subImages/image%d.webp", uniqueReportID.Text, imageIndex)
				url, err := uploadToS3(uploadPath, reader.URI().Path())
				if err != nil {
					dialog.ShowError(err, myWindow)
					return
				}

				subImagePath.SetText(url)
				dialog.ShowInformation("Success", "Sub image uploaded successfully", myWindow)
			}, myWindow)
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
			dialog.ShowError(fmt.Errorf("Please fill all required fields"), myWindow)
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
			"SubImages":       getSubImageData(subImageContainer),
		}

		// Convert to JSON
		jsonData, err := json.MarshalIndent(reportData, "", "  ")
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		// Save JSON file
		outputFolder := "output"
		if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
			err = os.Mkdir(outputFolder, os.ModePerm)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to create output folder: %v", err), myWindow)
				return
			}
		}

		fileName := filepath.Join(outputFolder, fmt.Sprintf("%s_%s.json", uniqueReportID.Text, reportName.Text))
		err = os.WriteFile(fileName, jsonData, 0644)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		dialog.ShowInformation("Success", fmt.Sprintf("Report saved as %s", fileName), myWindow)
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

	reportTab := container.NewTabItem("Report", scrollableContent)
	tabs := container.NewAppTabs(reportTab)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(600, 800))
	myWindow.ShowAndRun()
}

// Helper function to upload a file to S3
func uploadToS3(key, filePath string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %v", err)
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	svc := s3.New(sess)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(file),
		ContentType: aws.String("image/webp"), // Ensure correct content type
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	// Generate the S3 URL
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, awsRegion, key), nil
}

// Helper function to get sub-image data from the container
func getSubImageData(container *fyne.Container) []map[string]string {
	var subImages []map[string]string
	for _, obj := range container.Objects {
		if subImageItem, ok := obj.(*fyne.Container); ok {
			subImageData := make(map[string]string)
			for i, subObj := range subImageItem.Objects {
				if entry, ok := subObj.(*widget.Entry); ok {
					if entry.Wrapping == fyne.TextWrapWord {
						subImageData["Description"] = entry.Text
					} else {
						// Check the label that comes before this entry
						if i > 0 {
							if label, ok := subImageItem.Objects[i-1].(*widget.Label); ok {
								if label.Text == "Sub Image Name:" {
									subImageData["Name"] = entry.Text
								} else if label.Text == "Sub Image URL:" {
									subImageData["URL"] = entry.Text
								}
							}
						}
					}
				}
			}
			if len(subImageData) > 0 {
				subImages = append(subImages, subImageData)
			}
		}
	}
	return subImages
}
