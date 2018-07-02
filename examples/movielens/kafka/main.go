package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/lovoo/cofire"
	"github.com/lovoo/cofire/examples"
	"github.com/lovoo/cofire/examples/movielens"
	"github.com/lovoo/goka"
	"golang.org/x/sync/errgroup"
)

var (
	input      = flag.String("input", "/tmp/ratings", "input ratings file (Movie Lens format)")
	group      = flag.String("group", "cofire-mlens", "consumer group for learner")
	broker     = flag.String("broker", "localhost:9092", "a bootstrap Kafka broker")
	sample     = flag.Int("sample", 80, "percentage of the input ratings used for training")
	gamma      = flag.Float64("gamma", 0.001, "SGD gamma parameter")
	lambda     = flag.Float64("lambda", 0.01, "SGD lambda parameter")
	rank       = flag.Int("rank", 10, "number of latent features")
	iterations = flag.Int("iterations", 1, "number of iterations")
	delay      = flag.Duration("delay", time.Second, "reiteration delay")
)

func init() {
	flag.Parse()
}

func main() {
	var (
		brokers = []string{*broker}
		ggroup  = goka.Group(*group)
		ratings = movielens.ReadRatings(*input)
		ctx     = context.Background()
		train   = ratings[:len(ratings)**sample/100]
		test    = ratings[len(ratings)**sample/100:]
		params  = cofire.Parameters{
			Gamma:      *gamma,
			Lambda:     *lambda,
			Rank:       *rank,
			Iterations: *iterations,
		}
	)

	fmt.Printf("Train set: %d\n", len(train))
	fmt.Printf("Test  set: %d\n", len(test))
	fmt.Println(train[0:10])

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(examples.StartLearner(ctx, brokers, ggroup, params))
	grp.Go(examples.StartProducer(ctx, brokers, ggroup, train))
	grp.Go(examples.StartRefeeder(ctx, brokers, ggroup, *delay))
	view, startView := examples.CreateView(brokers, ggroup)
	grp.Go(startView(ctx))
	grp.Go(examples.StartValidator(ctx, view, test, params))

	if err := grp.Wait(); err != nil {
		fmt.Println(err)
	}
}
