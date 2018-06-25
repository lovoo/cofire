package cofire

import (
	"testing"
	"time"

	"github.com/lovoo/goka"
)

type mockContext struct {
	ts        time.Time
	emitCheck func(goka.Stream, string, interface{})
}

func (c *mockContext) Delete()                                     {}
func (c *mockContext) Emit(t goka.Stream, k string, m interface{}) { c.emitCheck(t, k, m) }
func (c *mockContext) Fail(error)                                  {}
func (c *mockContext) Join(goka.Table) interface{}                 { return nil }
func (c *mockContext) Lookup(goka.Table, string) interface{}       { return nil }
func (c *mockContext) Key() string                                 { return "key" }
func (c *mockContext) Loopback(string, interface{})                {}
func (c *mockContext) SetValue(interface{})                        {}
func (c *mockContext) Timestamp() time.Time                        { return c.ts }
func (c *mockContext) Topic() goka.Stream                          { return "stream" }
func (c *mockContext) Value() interface{}                          { return nil }

func equals(t *testing.T, actual, expected string) {
	if actual != expected {
		t.Errorf(`expected=%s, actual=%s`, expected, actual)
	}
}

func TestRefeeder(t *testing.T) {
	var (
		count int
		delay = 1 * time.Hour
		done  = make(chan bool)
		start = time.Now()
	)

	// create a refeed callback
	cb := refeed("topic", delay, func(ts time.Time) <-chan time.Time {
		if ts != start.Add(delay) {
			t.Errorf("unexpected time: %v (%v)", ts, start)
		}
		return waiter(start) // dont wait anything since start already over
	})

	// the message was created at start - 10 seconds
	ctx := new(mockContext)
	ctx.ts = start
	ctx.emitCheck = func(s goka.Stream, k string, m interface{}) {
		equals(t, string(s), "topic")
		equals(t, k, "key")
		equals(t, m.(string), "some message")
		count++
	}

	// we now process (ie, delay) the message
	go func() {
		cb(ctx, "some message")
		close(done)
	}()

	// wait for cb
	<-done

	if count != 1 {
		t.Errorf("count unexpected: %d times", count)
	}
}
