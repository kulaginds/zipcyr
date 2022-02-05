package main

import (
	"archive/zip"
	"fmt"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Help:")
		fmt.Println("./zipcyr file.zip")
		fmt.Println("It creates file_transcoded.zip file with UTF-8 charset of filenames in archive.")
		os.Exit(1)
	}

	sourceFilename := os.Args[1]
	destinationFilename := convertToTranscodedFilename(os.Args[1])

	sourceArchive, err := zip.OpenReader(sourceFilename)
	if err != nil {
		errorExit(err)
	}

	destinationArchive, err := OpenZipWriter(destinationFilename)
	if err != nil {
		errorExit(err)
	}

	var (
		writer io.Writer
		reader io.ReadCloser
	)

	for _, file := range sourceArchive.File {
		fileHeader := file.FileHeader
		fileHeader.Name, err = transcode(fileHeader.Name)
		if err != nil {
			errorExit(err)
		}

		writer, err = destinationArchive.CreateHeader(&fileHeader)
		if err != nil {
			errorExit(err)
		}

		reader, err = file.Open()
		if err != nil {
			errorExit(err)
		}

		_, err = io.Copy(writer, reader)
		if err != nil {
			errorExit(err)
		}

		err = reader.Close()
		if err != nil {
			errorExit(err)
		}
	}

	if err = destinationArchive.Close(); err != nil {
		errorExit(err)
	}

	if err = sourceArchive.Close(); err != nil {
		errorExit(err)
	}

	fmt.Println("Done!")
}

func errorExit(err error) {
	fmt.Println(fmt.Errorf("error: %w", err))
	os.Exit(1)
}

func OpenZipWriter(name string) (*zip.Writer, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	return zip.NewWriter(f), nil
}

func convertToTranscodedFilename(name string) string {
	filename := name
	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		filename = parts[len(parts)-1]
	}

	filenameParts := strings.Split(filename, ".")
	filename = filename + "_transcoded"
	if len(filenameParts) == 2 {
		filename = filenameParts[0] + "_transcoded." + filenameParts[1]
	}

	if len(parts) == 1 {
		return filename
	}

	parts[len(parts)-1] = filename

	return strings.Join(parts, "/")
}

func transcode(input string) (string, error) {
	enc, _, certain := charset.DetermineEncoding([]byte(input), "text/plain")
	if !certain {
		enc = charmap.CodePage866
	}

	return enc.NewDecoder().String(input)
}
