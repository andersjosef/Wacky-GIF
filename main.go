package main

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
)

const DELAY = 10 // Delay in 100th of a second

func main() {
	fileSRC, fileDST, err := getArguments()
	if err != nil {
		fmt.Println(err)
		return
	}
	// Open and decode the source image
	img, err := loadImage(fileSRC)
	if err != nil {
		fmt.Printf("Error loading image: %v\n", err)
		return
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// List of transformation functions
	transformations := []func(image.Image, int, int) draw.Image{
		func(img image.Image, width, height int) draw.Image {
			return convertImageHorizontal(img, width, height, 1, 1, 1)
		},
		func(img image.Image, width, height int) draw.Image {
			newImg := convertImageHorizontal(img, width, height, 0, 1, 1)
			newImg = convertImageHorizontal(newImg, width, height, 1, 1, 0)
			newImg = convertImageHorizontal(newImg, width, height, 1, 1, 1)
			return newImg
		},
		convertImageVertical,
		func(img image.Image, width, height int) draw.Image { return adjustBrightness(img, width, height, 4) },
		func(img image.Image, width, height int) draw.Image { return waveImage(img, width, height, 20, 20) },
		func(img image.Image, width, height int) draw.Image {
			return convertImageHorizontal(img, width, height, 1, 0, 1)
		},
		func(img image.Image, width, height int) draw.Image {
			return convertImageHorizontal(img, width, height, 0, 1, 1)
		},
		func(img image.Image, width, height int) draw.Image {
			newImg := kaleidoscopeImage(img, width, height)
			newImg = mergeImages(img, newImg)
			return newImg
		},
		func(img image.Image, width, height int) draw.Image {
			newImg := convertImageHorizontal(img, width, height, 1, 1, 1)
			newImg = convertImageHorizontal(newImg, width, height, 1, 1, 1)
			newImg = convertImageHorizontal(newImg, width, height, 1, 1, 1)
			return newImg
		},
		func(img image.Image, width, height int) draw.Image {
			newImg := waveImage(img, width, height, 100, 20)
			newImg = mergeImages(img, newImg)
			return newImg
		},
		kaleidoscopeImage,
		strong,
		sickTwist,
		func(img image.Image, width, height int) draw.Image {

			newImg := sickTwist(img, width, height)
			newImg = convertImageHorizontal(newImg, width, height, 1, 1, 1)
			newImg = convertImageHorizontal(newImg, width, height, 1, 1, 1)
			return newImg
		},
		func(img image.Image, width, height int) draw.Image {

			newImg := sickTwist(img, width, height)
			newImg = convertImageHorizontal(newImg, width, height, 1, 1, 1)
			return newImg
		},
	}

	// Shuffle the transformations
	shuffle(transformations)

	var images []*image.Paletted
	var delays []int
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create a goroutine for each transformation function
	// then convert them to paletted for the gif format
	for _, transform := range transformations {
		wg.Add(1)
		go func(transform func(image.Image, int, int) draw.Image) {
			defer wg.Done()
			transformedImg := transform(img, width, height)
			palettedImage := convertToPaletted(transformedImg)

			mu.Lock()
			images = append(images, palettedImage)
			delays = append(delays, DELAY)
			mu.Unlock()
		}(transform)
	}

	wg.Wait() // Wait for all the goroutines

	// Create GIF
	outputGif := &gif.GIF{
		Image: images,
		Delay: delays,
	}

	// Make GIF file
	outputFile, err := os.Create(fileDST)
	if err != nil {
		fmt.Println("Error creating GIF file:", err)
		return
	}
	defer outputFile.Close()

	// Write GIF to GIF file
	err = gif.EncodeAll(outputFile, outputGif)
	if err != nil {
		fmt.Println("Error encoding GIF:", err)
	}
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ext := filepath.Ext(path)
	switch ext {
	case ".png":
		return png.Decode(file)
	case ".jpg", ".jpeg":
		return jpeg.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func convertToPaletted(img image.Image) *image.Paletted {
	bounds := img.Bounds()
	paletted := image.NewPaletted(bounds, palette.Plan9)
	draw.FloydSteinberg.Draw(paletted, bounds, img, image.Point{})
	return paletted
}

func shuffle(slice []func(image.Image, int, int) draw.Image) {
	for i := len(slice) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Handeling the arguments for sourcefile destination file and stepsize
func getArguments() (fileSRC, fileDST string, err error) {
	if len(os.Args) != 3 {
		return "", "", fmt.Errorf("usage: ./program /source/path.jpeg /destination/path.gif")
	}

	fileSRC = os.Args[1]
	fileDST = os.Args[2]

	return fileSRC, fileDST, nil
}

/* ---------------- The Transformation Functions ---------------- */

func convertImageHorizontal(img image.Image, width, height int, one, two, three uint8) draw.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			col := img.At(x, y)
			col2 := img.At(width-x, y)
			r, g, _, a := col.RGBA()
			_, _, b2, _ := col2.RGBA()
			fillColor := color.RGBA{uint8(b2>>8) * one, uint8(g>>8) * two, uint8(r>>8) * three, uint8(a >> 8)}
			newImg.SetRGBA(x, y, fillColor)
		}
	}
	return newImg
}

func convertImageVertical(img image.Image, width, height int) draw.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			col := img.At(x, y)
			opCol := img.At(x, height-y-1)
			r, _, _, a := col.RGBA()
			_, og, ob, _ := opCol.RGBA()
			fillColor := color.RGBA{uint8(ob), uint8(og >> 8), uint8(r >> 8), uint8(a >> 8)}
			newImg.SetRGBA(x, y, fillColor)
		}
	}
	return newImg
}

func adjustBrightness(img image.Image, width, height int, factor float64) draw.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			col := img.At(x, y)
			r, g, b, a := col.RGBA()
			fillColor := color.RGBA{
				uint8(clamp(int(float64(r>>8) * factor))),
				uint8(clamp(int(float64(g>>8) * factor))),
				uint8(clamp(int(float64(b>>8) * factor))),
				uint8(a >> 8),
			}
			newImg.Set(x, y, fillColor)
		}
	}
	return newImg
}

func clamp(value int) int {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}

func waveImage(img image.Image, width, height int, amplitude, frequency float64) draw.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := int(amplitude * math.Sin(2*math.Pi*frequency*float64(y)/float64(height)))
			srcX := (x + offset) % width
			if srcX < 0 {
				srcX += width
			}
			newImg.Set(x, y, img.At(srcX, y))
		}
	}

	return newImg
}

func mergeImages(img1, img2 image.Image) draw.Image {
	bounds1 := img1.Bounds()
	width1 := bounds1.Dx()
	height1 := bounds1.Dy()
	bounds2 := img2.Bounds()
	width2 := bounds2.Dx()
	height2 := bounds2.Dy()

	minHeight := min(height1, height2)
	minWidth := min(width1, width2)

	newImg := image.NewRGBA(image.Rect(0, 0, minWidth, minHeight))

	for y := 0; y < minHeight; y++ {
		for x := 0; x < minWidth; x++ {
			col1 := img1.At(x, y)
			col2 := img2.At(x, y)
			_, g1, _, _ := col1.RGBA()
			r2, _, b2, _ := col2.RGBA()

			fillColor := color.RGBA{
				uint8(r2 >> 8),
				uint8(g1 >> 8),
				uint8(b2 >> 8),
				255,
			}
			newImg.Set(x, y, fillColor)
		}
	}
	return newImg
}

func kaleidoscopeImage(img image.Image, width, height int) draw.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x < width/2 {
				if y < height/2 {
					newImg.Set(x, y, img.At(x, y))
				} else {
					newImg.Set(x, y, img.At(x, height-y-1))
				}
			} else {
				if y < height/2 {
					newImg.Set(x, y, img.At(width-x-1, y))
				} else {
					newImg.Set(x, y, img.At(width-x-1, height-y-1))
				}
			}
		}
	}

	return newImg
}

func strong(img image.Image, width, height int) draw.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			col := img.At(x, y)
			r, g, b, _ := col.RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			colorFill := color.RGBA{0, 0, 0, 255}
			switch maxOfThree(r8, g8, b8) {
			case 'r':
				colorFill = color.RGBA{b8 / 2, g8 / 5, r8 / 2, 255}
			case 'g':
				colorFill = color.RGBA{0, b8, g8 / 2, 255}
			case 'b':
				colorFill = color.RGBA{g8 / 10, r8, b8 / 8, 255}
			}
			newImg.Set(x, y, colorFill)
		}
	}
	return newImg
}

func maxOfThree(r, g, b uint8) rune {
	if r >= g && r >= b {
		return 'r'
	} else if g >= r && g >= b {
		return 'g'
	} else {
		return 'b'
	}
}

func sickTwist(img image.Image, width, height int) draw.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			col := img.At(x, y)
			r, g, b, _ := col.RGBA()
			if (x+y)%2 != 0 {
				col = img.At(width-x, height-y)
				g, _, b, _ = col.RGBA()
			}

			fillCol := color.RGBA{
				uint8(r >> 8),
				uint8(g >> 8),
				uint8(b >> 8),
				255,
			}
			newImg.Set(x, y, fillCol)
		}
	}
	return newImg
}
