package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"
)

const threshold = 150
const sigma = 1

func rgbToGray(c color.Color) uint8 {
	r, g, b, _ := c.RGBA()
	return uint8((r + g + b) / 3 >> 8)
}

func gaussian(x, sigma float64) float64 {
	return math.Exp(-(x*x)/(2*sigma*sigma)) / (math.Sqrt(2*math.Pi) * sigma)
}

func applyGaussianBlur(src image.Image, sigma float64) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y

	kernelSize := int(6*sigma + 1)
	kernel := make([][]float64, kernelSize)
	for i := range kernel {
		kernel[i] = make([]float64, kernelSize)
	}

	var sum float64
	for y := -kernelSize / 2; y <= kernelSize/2; y++ {
		for x := -kernelSize / 2; x <= kernelSize/2; x++ {
			g := gaussian(math.Sqrt(float64(x*x+y*y)), sigma)
			kernel[y+kernelSize/2][x+kernelSize/2] = g
			sum += g
		}
	}

	// Normalize the kernel
	for i := range kernel {
		for j := range kernel[i] {
			kernel[i][j] /= sum
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, b, a float64
			for ky := 0; ky < kernelSize; ky++ {
				for kx := 0; kx < kernelSize; kx++ {
					srcX := x + kx - kernelSize/2
					srcY := y + ky - kernelSize/2

					if srcX >= 0 && srcX < w && srcY >= 0 && srcY < h {
						cr, cg, cb, ca := src.At(srcX, srcY).RGBA()
						r += float64(cr) * kernel[ky][kx]
						g += float64(cg) * kernel[ky][kx]
						b += float64(cb) * kernel[ky][kx]
						a += float64(ca) * kernel[ky][kx]
					}
				}
			}
			dst.SetRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.RGBA{
				R: uint8(r / 0x101),
				G: uint8(g / 0x101),
				B: uint8(b / 0x101),
				A: uint8(a / 0x101),
			})
		}
	}

	return dst
}

func removeWhiteBackground(im image.Image) image.Image {
	bounds := im.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := rgbToGray(im.At(x, y))
			if gray > threshold {
				dst.Set(x, y, color.Transparent)
			} else {
				dst.Set(x, y, im.At(x, y))
			}
		}
	}

	return dst
}

func getNextOutputFileName(baseFileName string) string {
	outputDir := "saved"
	fileExt := filepath.Ext(baseFileName)
	baseName := baseFileName[:len(baseFileName)-len(fileExt)]
	fileNum := 1
	for {
		outputFileName := fmt.Sprintf("%s%d%s", baseName, fileNum, fileExt)
		outputFilePath := filepath.Join(outputDir, outputFileName)
		if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
			return outputFilePath
		}
		fileNum++
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go input.jpg output.png")
		return
	}

	inputFileName := os.Args[1]
	outputFileName := os.Args[2]

	inputFile, err := os.Open(inputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	inputImage, _, err := image.Decode(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	// Mengukur waktu eksekusi
	startTime := time.Now()

	backgroundRemovedImage := removeWhiteBackground(inputImage)
	smoothedImage := applyGaussianBlur(backgroundRemovedImage, sigma)

	// Dapatkan nama file output berikutnya dengan nomor increment jika file sudah ada
	outputFilePath := getNextOutputFileName(outputFileName) // Change the output file extension to .png
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	err = png.Encode(outputFile, smoothedImage) // Use png.Encode instead of jpeg.Encode
	if err != nil {
		log.Fatal(err)
	}

	// Menghentikan pengukuran waktu
	elapsedTime := time.Since(startTime)

	fmt.Printf("Background removed and smoothed, saved to %s\n", outputFilePath)
	fmt.Printf("Execution time: %s\n", elapsedTime)
}
