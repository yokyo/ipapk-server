package utils

import (
	"image/png"
	"image"
	"os"
	"path/filepath"
)

func SaveIcon(icon image.Image, filename string) error  {
	file, err := os.Create(GetIconPath(filename))
	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, icon); err != nil {
		return err
	}
	return nil
}

func GetAppPath(filename string) string {
	return filepath.Join(".data", "app", filename)
}

func GetIconPath(filename string) string {
	return filepath.Join(".data", "icon", filename)
}