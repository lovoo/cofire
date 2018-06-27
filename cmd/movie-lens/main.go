package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/lovoo/cofire"
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
	params     cofire.Parameters
)

func init() {
	flag.Parse()
	params = cofire.Parameters{
		Gamma:      *gamma,
		Lambda:     *lambda,
		Rank:       *rank,
		Iterations: *iterations,
	}
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

// starts a learner processor
func startLearner(ctx context.Context, brokers []string, group goka.Group) func() error {
	return func() error {
		// validator prints the RMSE every couple of seconds
		validator := cofire.NewErrorValidator()
		go func() {
			for {
				cnt := validator.Count()
				if cnt > 0 {
					rmse := validator.RMSE()
					fmt.Printf("RMSE: %.8f Count: %d \n", rmse, cnt)
				}
				time.Sleep(1 * time.Second)
			}
		}()

		// create a new group graph with the cofire group, validator and SGD
		// parameters.
		gg := cofire.NewLearner(group, validator, params)
		p, err := goka.NewProcessor(brokers, gg)
		if err != nil {
			return err
		}

		// start processor
		return p.Run(ctx)
	}
}

// start a refeeder processor
func startRefeeder(ctx context.Context, brokers []string, group goka.Group) func() error {
	return func() error {
		gg := cofire.NewRefeeder(group, time.Second)
		p, err := goka.NewProcessor(brokers, gg)
		if err != nil {
			return err
		}
		return p.Run(ctx)
	}
}

// start a producer of ratings
func startProducer(ctx context.Context, brokers []string, group goka.Group, ratings []cofire.Rating) func() error {
	return func() error {
		emitter, err := goka.NewEmitter(brokers,
			goka.Stream(fmt.Sprintf("%s-input", string(group))),
			new(cofire.RatingCodec))
		if err != nil {
			return err
		}
		defer emitter.Finish()

		// wait until processor is up
		time.Sleep(5 * time.Second)
		for _, r := range ratings {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			emitter.EmitSync(r.UserId, &r)
		}
		log.Println("finished loading input data")
		return nil
	}
}

// create a view of the model
func createView(brokers []string, group goka.Group) (*goka.View, func(ctx context.Context) func() error) {
	view, err := goka.NewView(brokers, goka.GroupTable(group), new(cofire.EntryCodec))
	return view, func(ctx context.Context) func() error {
		return func() error {
			if err != nil {
				return err
			}
			return view.Run(ctx)
		}
	}
}

// start validator
func startValidator(ctx context.Context, view *goka.View, ratings []cofire.Rating) func() error {
	return func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			var (
				v   = cofire.NewErrorValidator()
				u   *cofire.Features
				p   *cofire.Features
				sgd = cofire.NewSGD(cofire.DefaultParams().Gamma, cofire.DefaultParams().Lambda)
			)

			for _, r := range ratings {
				eu, err := view.Get(r.UserId)
				if err != nil {
					return err
				}
				if eu == nil {
					continue
				}
				u = eu.(*cofire.Entry).U
				if u == nil {
					continue
				}
				ep, err := view.Get(r.ProductId)
				if err != nil {
					return err
				}
				if ep == nil {
					continue
				}
				p = ep.(*cofire.Entry).P
				if p == nil {
					continue
				}

				sgd.Add(r.Score)
				v.Validate(u.Predict(p, sgd.Bias()), r.Score)
			}
			time.Sleep(3 * time.Second)
			fmt.Printf("TEST RSME: %.8f Count: %d\n", v.RMSE(), v.Count())
		}
		return nil
	}
}

func main() {
	var (
		brokers = []string{*broker}
		ggroup  = goka.Group(*group)
		ratings = readRatings(*input)
		ctx     = context.Background()
		train   = ratings[:len(ratings)**sample/100]
		test    = ratings[len(ratings)**sample/100:]
	)

	fmt.Printf("Train set: %d\n", len(train))
	fmt.Printf("Test  set: %d\n", len(test))
	fmt.Println(train[0:10])

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(startLearner(ctx, brokers, ggroup))
	grp.Go(startProducer(ctx, brokers, ggroup, train))
	grp.Go(startRefeeder(ctx, brokers, ggroup))
	view, startView := createView(brokers, ggroup)
	grp.Go(startView(ctx))
	grp.Go(startValidator(ctx, view, test))

	if err := grp.Wait(); err != nil {
		fmt.Println(err)
	}
}
