package utility

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
)

func ExtractFirstPage(pdfBytes []byte) ([]byte, error) {
	doc, err := fitz.NewFromMemory(pdfBytes)
	if err != nil {
		fmt.Println("Error opening PDF:", err)
		return []byte{}, err
	}
	defer doc.Close()

	outputDir := filepath.Join(".", "output_images")

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755)
		if err != nil {
			fmt.Println("Error creating output directory:", err)
			return []byte{}, err

		}
	}

	img, err := doc.Image(0)
	if err != nil {
		fmt.Println("Error extracting image from page:", err)
		return []byte{}, err
	}

	// baseName := filepath.Base(pdfPath)

	// outputFileName := fmt.Sprintf("first_page_%s.jpg", baseName)
	// outputFilePath := filepath.Join(outputDir, outputFileName)

	// f, err := os.Create(outputFilePath)
	// if err != nil {
	// 	fmt.Println("Error creating image file:", err)
	// 	return err
	// }
	// img.
	buf := new(bytes.Buffer)
	// bufio.
	// err := jpeg.Encode(buf, new_image, nil)
	// send_s3 := buf.Bytes()
	// bytes.
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
	if err != nil {
		fmt.Println("Error encoding image to JPEG:", err)
		return []byte{}, err
	}

	// f.Close()

	// fmt.Printf("PDF first page converted to image successfully: %s\n", pdfPath)
	return buf.Bytes(), nil
}
