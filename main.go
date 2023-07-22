package main

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("No directory path provided.")
		return
	}
	directories := os.Args[1:]
	for _, dir := range directories {
		path, err := extractLatLong(dir)
		if err != nil {
			fmt.Println(fmt.Sprintf("CSV file generation failure for %s : %s", path, err.Error()))
			continue
		}
		fmt.Println(fmt.Sprintf("CSV file generated successfully:%s", path))
	}
}

func extractLatLong(dir string) (string, error) {

	// Open the output CSV file for writing
	outputFilePath := path.Join(dir, "lat_long.csv")
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println(fmt.Sprintf("%s:Error creating output file, %s", dir, err.Error()))
		return dir, err
	}
	defer outputFile.Close()

	writer = csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write the header row to the CSV file
	writer.Write([]string{"File Path", "Latitude", "Longitude"})

	err = filepath.WalkDir(dir, visit)
	if err != nil {
		fmt.Println(fmt.Sprintf("%s:Error iterating directory, %s", dir, err.Error()))
		return dir, err
	}
	return outputFilePath, nil
}

var writer *csv.Writer

var validExt = []string{".jpg", ".gif", ".png", ".jpeg"}

func isImageFile(d fs.DirEntry) bool {
	for _, ext := range validExt {
		if strings.HasSuffix(d.Name(), ext) {
			return true
		}
	}
	return false
}

func visit(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	// Check if the path is a directory
	if d.IsDir() {
		return nil
	}
	// Check if the file is an image
	if isImageFile(d) {
		// Open the image file
		file, err := os.Open(path)
		if err != nil {
			// Skip to the next file
			return nil
		}
		defer file.Close()

		// Decode the EXIF data from the image
		exifData, err := exif.Decode(file)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s:Error decoding EXIF data, %s", path, err.Error()))
			// Skip to the next file
			return nil
		}
		// jsonByte, err := exifData.MarshalJSON()
		// if err != nil {
		// 	log.Fatal(err.Error())
		// }

		// jsonString := string(jsonByte)
		// fmt.Println(jsonString)

		// Get the latitude and longitude from the EXIF data
		lat, long, err := exifData.LatLong()
		if err != nil {
			fmt.Println(fmt.Sprintf("%s:Error getting latitude and longitude, %s", path, err.Error()))
			// Skip to the next file
			return nil
		}
		// Write the latitude and longitude to the CSV file
		writer.Write([]string{path, fmt.Sprintf("%f", lat), fmt.Sprintf("%f", long)})
	}
	return nil
}
