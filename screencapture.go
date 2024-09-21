package main

import (
	"errors"
	"fmt"
	"image"
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