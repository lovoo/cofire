package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"math/rand"
	"strconv"
	"time"

	"github.com/lovoo/cofire"
	"github.com/lovoo/cofire/examples/pixelreco"
)

var (
	input      = flag.String("input", "/tmp/figure.jpg", "input figure to be used as rating")
	sample     = flag.Int("sample", 80, "percentage of the input ratings used for training")
	gamma      = flag.Float64("gamma", 0.001, "SGD gamma parameter")
	lambda     = flag.Float64("lambda", 0.01, "SGD lambda parameter")
	rank       = flag.Int("rank", 10, "number of latent features")
	iterations = flag.Int("iterations", 1, "number of iterations")
	delay      = flag.Duration("delay", time.Second, "reiteration delay")
)

func init() {
	flag.Parse()
	rand.Seed(int64(time.Now().Unix()))
}

func main() {
	var (
		ratings, img = pixelreco.ReadRatings(*input)
		train        = ratings[:len(ratings)**sample/100]
		test         = ratings[len(ratings)**sample/100:]
		params       = cofire.Parameters{
			Gamma:      *gamma,
			Lambda:     *lambda,
			Rank:       *rank,
			Iterations: *iterations,
		}
		trainError = cofire.NewErrorValidator()
		testError  = cofire.NewErrorValidator()
		sgd        = cofire.NewSGD(params)
		model      = make(map[string]*cofire.Entry)
		animated   = &gif.GIF{}
		k          = 0
		n          = len(ratings) * *iterations / 20
	)

	for i := 0; i < *iterations; i++ {

		// train model
		for _, r := range train {
			user := model[r.UserId]
			if user == nil {
				user = new(cofire.Entry)
			}
			if user.U == nil {
				user.U = cofire.NewFeatures(*rank).Randomize()
			}
			product := model[r.ProductId]
			if product == nil {
				product = new(cofire.Entry)
			}
			if product.P == nil {
				product.P = cofire.NewFeatures(*rank).Randomize()
			}

			// train
			sgd.Apply(user.U, product.P, r.Score)
			sgd.Apply(product.P, user.U, r.Score)

			// update table
			model[r.UserId] = user
			model[r.ProductId] = product

			trainError.Validate(user.U.Predict(product.P, sgd.Bias()), r.Score)
			fmt.Printf("RSME: %.8f Count: %d\n", trainError.RMSE(), trainError.Count())

			if k%n == 0 {
				// check with test set
				output := image.NewGray(img.Bounds())
				for _, r := range ratings {
					user := model[r.UserId]
					product := model[r.ProductId]
					if user == nil || user.U == nil || product == nil || product.P == nil {
						continue
					}

					p := user.U.Predict(product.P, sgd.Bias())
					x, _ := strconv.Atoi(r.UserId)
					y, _ := strconv.Atoi(r.ProductId)
					output.SetGray(x, y, color.Gray{Y: uint8(p)})
				}
				animated = pixelreco.AppendGif(animated, output)
			}
			k++
		}

		// check with test set
		output := image.NewGray(img.Bounds())
		_ = test
		for _, r := range ratings {
			user := model[r.UserId]
			product := model[r.ProductId]
			if user == nil || user.U == nil || product == nil || product.P == nil {
				continue
			}

			p := user.U.Predict(product.P, sgd.Bias())
			x, _ := strconv.Atoi(r.UserId)
			y, _ := strconv.Atoi(r.ProductId)
			output.SetGray(x, y, color.Gray{Y: uint8(p)})

			testError.Validate(p, r.Score)
		}
		fmt.Printf("TEST RSME: %.8f Count: %d\n", testError.RMSE(), testError.Count())
		animated = pixelreco.AppendGif(animated, output)
	}
	pixelreco.SaveGif(animated, "output")
}
