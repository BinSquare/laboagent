package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"time"

	"github.com/kbinani/screenshot"
)

func captureDesktop() (error, *image.RGBA) {
	// Take a screenshot of the first display
	n := screenshot.NumActiveDisplays()
	if n <= 0 {

		return errors.New("Error: no active displays found."), nil
	}

	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return errors.New("Error: Failed to capture screenshot."), nil
	}

	return nil, img
}

func imgToPNG(img *image.RGBA) error {
	// Save the screenshot to a PNG file
	fileName := fmt.Sprintf("screenshot_%d.png", time.Now().Unix())
	file, err := os.Create(fileName)

	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return err
	}

	fmt.Printf("Screenshot saved to %s\n", fileName)
	return nil
}

func imgToBase64(img *image.RGBA) (error, string) {
	// Create a buffer to hold the encoded JPEG image
	var buf bytes.Buffer

	// Encode the image to JPEG format (with specified quality) and write it to the buffer
	options := &jpeg.Options{Quality: 30} // Quality ranges from 1 (low) to 100 (high)
	if err := jpeg.Encode(&buf, img, options); err != nil {
		return fmt.Errorf("failed to encode image: %w", err), ""
	}

	// Convert the buffer data to a base64 string
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	return nil, base64Str
}
