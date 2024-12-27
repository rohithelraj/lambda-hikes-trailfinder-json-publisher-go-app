package helpers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func GetSubImageData(container *fyne.Container) []map[string]string {
	var subImages []map[string]string
	for _, obj := range container.Objects {
		if subImageItem, ok := obj.(*fyne.Container); ok {
			subImageData := make(map[string]string)
			for i, subObj := range subImageItem.Objects {
				if entry, ok := subObj.(*widget.Entry); ok {
					if entry.Wrapping == fyne.TextWrapWord {
						subImageData["Description"] = entry.Text
					} else {
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
