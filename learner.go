package cofire

import (
	fmt "fmt"

	"github.com/lovoo/goka"
)

// NewLearner returns the GroupGraph for a learner processor.
func NewLearner(group goka.Group, validator Validator, params Parameters) *goka.GroupGraph {
	var (
		input  = fmt.Sprintf("%s-input", group)
		update = fmt.Sprintf("%s-update", group)
		refeed = fmt.Sprintf("%s-refeed", group)
	)
	p := newLearner(string(group), validator, params)
	edges := []goka.Edge{
		goka.Input(goka.Stream(input), new(RatingCodec), p.entry),
		goka.Input(goka.Stream(update), new(UpdateCodec), p.update),
		goka.Loop(new(messageCodec), p.stages(goka.Stream(refeed))),
		goka.Persist(new(EntryCodec)),
		goka.Output(goka.Stream(refeed), new(messageCodec)),
	}
	return goka.DefineGroup(group, edges...)
}

// Learner factorizes a user-prodcut rating matrix by learning latent features
// for users and products.
type Learner struct {
	group  string
	params Parameters
	v      Validator
	sgd    *SGD
}

// newLearner creates a new cofire learner.
func newLearner(group string, validator Validator, params Parameters) *Learner {
	return &Learner{
		group:  group,
		params: params,
		v:      validator,
		sgd:    NewSGD(params.Gamma, params.Lambda),
	}
}

// entry receives a Rating message in initiates a learning iteration.
func (l *Learner) entry(ctx goka.Context, m interface{}) {
	msg := m.(*Rating)
	e := getEntry(ctx)

	if e.U == nil {
		e.U = NewFeatures(l.params.Rank).Randomize()
		setEntry(ctx, e)
	}

	// send U to product
	ctx.Loopback(msg.ProductId, &Message{
		Stage:  Stage_PRODUCT,
		Rating: msg,
		F:      e.U,
		Iters:  uint32(l.params.Iterations),
	})
}

//
//     USER
//      |
//      * Entry                     PRODUCT
//      |        U features            |
//      +----------------------------->|
//      |                              |
//      |                              * Update P
//      |        P features            |
//      |<-----------------------------+
//      |
//      * Update U
//
func (l *Learner) stages(refeed goka.Stream) goka.ProcessCallback {
	return func(ctx goka.Context, m interface{}) {
		msg := m.(*Message)
		e := getEntry(ctx)

		switch msg.Stage {
		case Stage_ENTRY:
			// send U to product
			if e.U == nil {
				e.U = NewFeatures(l.params.Rank).Randomize()
				setEntry(ctx, e)
			}
			msg.Stage++
			msg.F = e.U
			ctx.Loopback(msg.Rating.ProductId, msg)

		case Stage_PRODUCT:
			// validate prediction
			if e.P == nil {
				e.P = NewFeatures(l.params.Rank).Randomize()
			}
			l.v.Validate(e.P.Predict(msg.F, l.sgd.Bias()), msg.Rating.Score)

			// update P
			l.sgd.Apply(e.P, msg.F, msg.Rating.Score)
			setEntry(ctx, e)

			// send P to user
			msg.Stage++
			msg.F = e.P
			ctx.Loopback(msg.Rating.UserId, msg)

		case Stage_USER:
			// update U
			if e.U == nil { // should never happen
				e.U = NewFeatures(l.params.Rank).Randomize()
			}
			l.sgd.Apply(e.U, msg.F, msg.Rating.Score)
			setEntry(ctx, e)

			// reiterate?
			if msg.Iters > 1 {
				msg.Iters--
				msg.Stage = Stage_ENTRY
				msg.F = nil
				ctx.Emit(refeed, ctx.Key(), msg)
			}
		}
	}
}

// update updates feature vectors of the model.
func (l *Learner) update(ctx goka.Context, m interface{}) {
	msg := m.(*Update)

	// fetch state
	e := getEntry(ctx)

	// update state
	if msg.U != nil {
		e.U = msg.U
	}
	if msg.P != nil {
		e.P = msg.P
	}

	// save state
	setEntry(ctx, e)
}

func getEntry(ctx goka.Context) *Entry {
	e, ok := ctx.Value().(*Entry)
	if !ok {
		return new(Entry)
	}
	return e
}

func setEntry(ctx goka.Context, e *Entry) {
	ctx.SetValue(e)
}
