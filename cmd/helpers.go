package main

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"image"
	"image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"strings"
)

type Printer struct{}

func (p Printer) Walk(name exif.FieldName, tag *tiff.Tag) error {
	fmt.Printf("%v %v\n", name, tag)
	return nil
}

func stripMetadata(inputPath, outputPath string) error {
	input, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("error opening file %v", err)
	}
	defer input.Close()

	img, format, err := image.Decode(input)
	if err != nil {
		return fmt.Errorf("error decoding image: %v", err)
	}

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer output.Close()

	switch strings.ToLower(format) {
	case "jpeg":
		jpeg.Encode(output, img, &jpeg.Options{Quality: 100})
	case "png":
		png.Encode(output, img)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err := verifyNoMetadata(outputPath); err != nil {
		return fmt.Errorf("verification failed: %v", err)
	}
	return nil
}

func verifyNoMetadata(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	header := make([]byte, 50)
	_, err = f.Read(header)
	if err != nil {
		return err
	}

	// Check for common metadata markers
	markers := [][]byte{
		{0xFF, 0xE1}, // EXIF
		{0xFF, 0xE2}, // ICC Profile
		{0xFF, 0xED}, // IPTC
		{0xFF, 0xE0}, // JFIF
	}

	for i := 0; i < len(header)-1; i++ {
		for _, marker := range markers {
			if header[i] == marker[0] && header[i+1] == marker[1] {
				return fmt.Errorf("found metadata marker %X%X at position %d", marker[0], marker[1], i)
			}
		}
	}
	return nil

}
func readMetadata(fileName string) {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("can't open file")
		return
	}
	defer f.Close()

	// Check if it's a JPEG by reading the first few bytes
	header := make([]byte, 2)
	if _, err := f.Read(header); err != nil {
		fmt.Printf("Error reading file header: %v\n", err)
		return
	}

	// Check JPEG signature (0xFF 0xD8)
	if header[0] == 0xFF && header[1] == 0xD8 {
		fmt.Println("File is JPEG")
	} else {
		fmt.Println("File is not a JPEG - EXIF data can only be read from JPEGs")
		return
	}

	// Return to the start of the file before reading EXIF
	if _, err := f.Seek(0, 0); err != nil {
		fmt.Printf("Error seeking to start of file: %v\n", err)
		return
	}

	// Try to decode EXIF data
	x, err := exif.Decode(f)

	//If we get here, we have valid EXIF data
	fmt.Println("Successfully found EXIF data. Reading tags...")

	var p Printer
	if x != nil {
		err = x.Walk(p)
	}

	if err != nil {
		fmt.Printf("Error walking through EXIF data: %v\n", err)
		return
	}

}
