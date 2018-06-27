package pixelreco

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"log"
	"math/rand"
	"os"

	"github.com/lovoo/cofire"
)

// Reads a jpeg or gif image and returns a list of ratings using x values as
// users, y values as products, and gray value as score.
func ReadRatings(fname string) ([]cofire.Rating, image.Image) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	var ratings []cofire.Rating
	for x := 0; x <= img.Bounds().Max.X; x++ {
		for y := 0; y <= img.Bounds().Max.Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// convert RGB to gray level
			lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			pixel := color.Gray{uint8(lum / 256)}

			// convert pixel to score
			ratings = append(ratings, cofire.Rating{
				UserId:    fmt.Sprintf("%d", x),
				ProductId: fmt.Sprintf("%d", y),
				Score:     float64(pixel.Y),
			})
		}
	}

	rand.Shuffle(len(ratings), func(i, j int) {
		ratings[i], ratings[j] = ratings[j], ratings[i]
	})
	return ratings, img
}

func SaveImage(img image.Image, name string, iteration int) error {
	outFile, err := os.Create(fmt.Sprintf("%s_%d.jpg", name, iteration))
	if err != nil {
		return err
	}
	defer outFile.Close()
	jpeg.Encode(outFile, img, nil)
	return nil
}

func AppendGif(g *gif.GIF, img image.Image) *gif.GIF {
	var pal []color.Color
	for i := 0; i < 256; i++ {
		pal = append(pal, color.Gray{Y: uint8(i)})
	}
	imgNew := image.NewPaletted(img.Bounds(), pal)
	for x := 0; x <= img.Bounds().Max.X; x++ {
		for y := 0; y <= img.Bounds().Max.Y; y++ {
			imgNew.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}

	g.Image = append(g.Image, imgNew)
	g.Delay = append(g.Delay, 1)

	return g
}

func SaveGif(g *gif.GIF, name string) error {
	outFile, err := os.Create(fmt.Sprintf("%s.gif", name))
	if err != nil {
		return err
	}
	defer outFile.Close()
	g.Delay[0] = 50
	g.Delay[len(g.Delay)-1] = 100
	gif.EncodeAll(outFile, g)

	return nil
}
