package cofire

import (
	"fmt"
	"time"

	"github.com/lovoo/goka"
)

type waitUntil func(time.Time) <-chan time.Time

// waiter until absolute time t
func waiter(t time.Time) <-chan time.Time {
	return time.After(time.Until(t))
}

func refeed(loop goka.Stream, delay time.Duration, wait waitUntil) goka.ProcessCallback {
	return func(ctx goka.Context, m interface{}) {
		<-wait(ctx.Timestamp().Add(delay))
		ctx.Emit(loop, ctx.Key(), m)
	}
}

// NewRefeeder returns the GroupGraph for a processor that refeeds the input of
// the learner after a specified delay.
func NewRefeeder(cofireGroup goka.Group, delay time.Duration) *goka.GroupGraph {
	var (
		group = fmt.Sprintf("%s-refeed", cofireGroup)
		input = fmt.Sprintf("%s-refeed", cofireGroup)
		loop  = fmt.Sprintf("%s-loop", cofireGroup)
	)
	return goka.DefineGroup(goka.Group(group),
		goka.Input(
			goka.Stream(input),
			new(messageCodec),
			refeed(goka.Stream(loop), delay, waiter),
		),
		goka.Output(goka.Stream(loop), new(messageCodec)),
	)
}
