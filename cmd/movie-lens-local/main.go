package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/lovoo/cofire"
)

var (
	input      = flag.String("input", "/tmp/ratings", "input ratings file (Movie Lens format)")
	iterations = flag.Int("iterations", 1, "number of iterations")
	sample     = flag.Int("sample", 80, "percentage of the input ratings used for training")
	gamma      = flag.Float64("gamma", 0.001, "SGD gamma parameter")
	lambda     = flag.Float64("lambda", 0.01, "SGD lambda parameter")
	rank       = flag.Int("rank", 10, "number of latent features")
)

func init() {
	flag.Parse()
}

// read ratings from file, and drop
func readRatings(fname string) []cofire.Rating {
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	var ratings []cofire.Rating
	for _, l := range strings.Split(string(dat), "\n") {
		if l == "" {
			continue
		}
		e := strings.Split(l, ",")
		if len(e) != 4 {
			log.Print(l)
			log.Fatal("!= 4")
		}

		s, _ := strconv.ParseFloat(e[2], 64)
		ratings = append(ratings, cofire.Rating{
			UserId:    e[0],
			ProductId: e[1],
			Score:     s,
		})
	}

	rand.Shuffle(len(ratings), func(i, j int) {
		ratings[i], ratings[j] = ratings[j], ratings[i]
	})
	return ratings
}

func main() {
	var (
		model      = make(map[string]*cofire.Entry)
		sgd        = cofire.NewSGD(*gamma, *lambda)
		ratings    = readRatings(*input)
		train      = ratings[:len(ratings)**sample/100]
		test       = ratings[len(ratings)**sample/100:]
		trainError = cofire.NewErrorValidator()
		testError  = cofire.NewErrorValidator()
	)

	fmt.Printf("Train set: %d\n", len(train))
	fmt.Printf("Test  set: %d\n", len(test))
	fmt.Println(train[0:10])

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
		}

		// check with test set
		for _, r := range test {
			user := model[r.UserId]
			product := model[r.ProductId]
			if user == nil || user.U == nil || product == nil || product.P == nil {
				continue
			}
			testError.Validate(user.U.Predict(product.P, sgd.Bias()), r.Score)
		}
		fmt.Printf("TEST RSME: %.8f Count: %d\n", testError.RMSE(), testError.Count())

	}
}
