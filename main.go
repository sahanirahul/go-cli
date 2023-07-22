package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rwcarlsen/goexif/exif"
)

/*
use the below command to run this go program (html is optional, pass that if you need the exif data to be written in html file too)

go run main.go /images/directory/path html

** the program will generate a csv file with the exif data of the images in the provided directory and its sub-directory **
** if html is passed in the command line argument then the same data is also written in a html file **
*/
func main() {
	if len(os.Args) == 1 {
		fmt.Println("No directory path provided.")
		return
	}
	dir := os.Args[1]
	_csv := true
	var _html bool
	if len(os.Args) > 2 && os.Args[2] == "html" {
		_html = true
	}
	res, err := extractLatLong(dir, _csv, _html)
	if err != nil {
		if len(res.CSVFilePath) == 0 && len(res.HTMLFilePath) == 0 {
			fmt.Println(fmt.Sprintf("file generation failure for %s : %s", res.Directory, err.Error()))
		} else {
			fmt.Println(fmt.Sprintf("file generation partial success %s,%s,%s", res.CSVFilePath, res.HTMLFilePath, err.Error()))
		}
		return
	}
	fmt.Println(fmt.Sprintf("File generated successfully:%s,%s", res.CSVFilePath, res.HTMLFilePath))
}

type response struct {
	Directory    string
	CSVFilePath  string
	HTMLFilePath string
}

func extractLatLong(dir string, _csv, _html bool) (response, error) {
	res := response{
		Directory: dir,
	}
	if !_csv && !_html {
		return res, errors.New("no file format provided to write lat-long data")
	}
	if _csv {
		csvOutputFilePath := path.Join(dir, "lat_long.csv")
		outputFile, err := os.Create(csvOutputFilePath)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s:Error creating output file, %s", dir, err.Error()))
			return res, err
		}
		defer outputFile.Close()

		writer = csv.NewWriter(outputFile)
		defer writer.Flush()

		// csv file header
		writer.Write([]string{"File Path", "Latitude", "Longitude"})

		err = filepath.WalkDir(dir, visit)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s:Error iterating directory, %s", dir, err.Error()))
			return res, err
		}
		res.CSVFilePath = csvOutputFilePath
	}
	if _html {
		outputFileHtmlPath := path.Join(dir, "lat_long.html")
		err := writeToHTMLFile(outputFileHtmlPath, lat_long_data)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s:Error generating html file, %s", dir, err.Error()))
			return res, err
		}
		res.HTMLFilePath = outputFileHtmlPath
	}
	return res, nil
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
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		// Decode the EXIF data from the image
		exifData, err := exif.Decode(file)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s:Error decoding EXIF data, %s", path, err.Error()))
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
			return nil
		}
		// writing the latitude and longitude to the CSV file
		writer.Write([]string{path, fmt.Sprintf("%f", lat), fmt.Sprintf("%f", long)})
		lat_long_data = append(lat_long_data, FilePathWithLatLong{FilePath: path, Latitute: lat, Longitude: long})
	}
	return nil
}

// Create a struct to hold data for the table
type FilePathWithLatLong struct {
	FilePath  string
	Latitute  float64
	Longitude float64
}

var lat_long_data []FilePathWithLatLong

func writeToHTMLFile(filepath string, data []FilePathWithLatLong) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Pparsing html template string
	tmpl, err := template.New("html_page").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(file, map[string]interface{}{
		"Rows": data,
	})
	if err != nil {
		return err
	}
	return nil
}

// HTML template
var htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>EXIF DATA</title>
</head>
<body>
	<h1>Exif Data</h1>
	<table border="1">
		<tr>
			<th>Image Path</th>
			<th>Latitude</th>
			<th>Longitude</th>
		</tr>
		{{range .Rows}}
		<tr>
			<td>{{.FilePath}}</td>
			<td>{{.Latitute}}</td>
			<td>{{.Longitude}}</td>
		</tr>
		{{end}}
	</table>
</body>
</html>`
