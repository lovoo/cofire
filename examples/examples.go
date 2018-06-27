package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lovoo/cofire"
	"github.com/lovoo/goka"
)

// starts a learner processor
func StartLearner(ctx context.Context, brokers []string, group goka.Group, params cofire.Parameters) func() error {
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
func StartRefeeder(ctx context.Context, brokers []string, group goka.Group, delay time.Duration) func() error {
	return func() error {
		gg := cofire.NewRefeeder(group, delay)
		p, err := goka.NewProcessor(brokers, gg)
		if err != nil {
			return err
		}
		return p.Run(ctx)
	}
}

// start a producer of ratings
func StartProducer(ctx context.Context, brokers []string, group goka.Group, ratings []cofire.Rating) func() error {
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
			emitter.Emit(r.UserId, &r)
			time.Sleep(5 * time.Millisecond)
		}
		log.Println("finished loading input data")
		return nil
	}
}

// create a view of the model
func CreateView(brokers []string, group goka.Group) (*goka.View, func(ctx context.Context) func() error) {
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
func StartValidator(ctx context.Context, view *goka.View, ratings []cofire.Rating, params cofire.Parameters) func() error {
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
				sgd = cofire.NewSGD(params.Gamma, params.Lambda)
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
