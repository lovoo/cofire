package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"

	"github.com/lovoo/cofire"
	"github.com/lovoo/cofire/examples"
	"github.com/lovoo/cofire/examples/pixelreco"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/web/index"
	"github.com/lovoo/goka/web/monitor"
	"github.com/lovoo/goka/web/query"
)

var (
	input      = flag.String("input", "/tmp/figure.jpg", "input figure to be used as rating")
	group      = flag.String("group", "cofire-image", "consumer group for learner")
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
	rand.Seed(int64(time.Now().Unix()))
}

// start validator
func startPicValidator(ctx context.Context, view *goka.View, ratings []cofire.Rating, params cofire.Parameters, createImage func() *image.Gray) func() error {
	return func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			var (
				output = createImage()
				u      *cofire.Features
				p      *cofire.Features
				sgd    = cofire.NewSGD(params.Gamma, params.Lambda)
				//g      = &gif.GIF{}
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

				prediction := u.Predict(p, sgd.Bias())
				if prediction < 0.0 {
					prediction = 0.0
				}
				prediction *= 256.0
				x, _ := strconv.Atoi(r.UserId)
				y, _ := strconv.Atoi(r.ProductId)
				fmt.Println(x, r.UserId, y, r.ProductId, color.Gray{Y: uint8(prediction)})
				//output.SetGray(x, y, color.Gray{Y: uint8(r.Score * 256)})
				output.SetGray(x, y, color.Gray{Y: uint8(prediction)})
				//g = appendGif(g, output)
			}
			err := pixelreco.SaveImage(output, "lovoo_gray", 100)
			if err != nil {
				log.Fatal(err)
			}
			//saveGif(g, "lovoo_grayy")
			//return errors.New("end")
			fmt.Println("END")
			time.Sleep(10 * time.Second)
		}
	}
}

func main() {
	var (
		brokers = []string{*broker}
		ggroup  = goka.Group(*group)
		ratings = pixelreco.ReadRatings(*input)
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

	root := mux.NewRouter()
	monitorServer := monitor.NewServer("/monitor", root)
	queryServer := query.NewServer("/query", root)
	idxServer := index.NewServer("/", root)
	idxServer.AddComponent(monitorServer, "Monitor")
	idxServer.AddComponent(queryServer, "Query")
	monitorServer.AttachView(view)
	queryServer.AttachSource("table", view.Get)
	root.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		value, _ := view.Get(mux.Vars(r)["key"])
		data, _ := json.Marshal(value)
		w.Write(data)
	})

	fmt.Println("View opened at http://localhost:9095/")
	go http.ListenAndServe(":9095", root)

	grp.Go(examples.StartValidator(ctx, view, test, params))
	grp.Go(startPicValidator(ctx, view, test, params, func() *image.Gray {
		f, err := os.Open(*input)
		if err != nil {
			log.Fatal(err)
		}
		img, _, err := image.Decode(f)
		f.Close()

		return image.NewGray(img.Bounds())
	}))

	if err := grp.Wait(); err != nil {
		fmt.Println(err)
	}
}
