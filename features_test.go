package cofire

import (
	"testing"

	"github.com/facebookgo/ensure"
)

// make features
func makeFeatures(v []float64) *Features {
	f := NewFeatures(DefaultParams().Rank)
	for i := range v {
		f.V[i] = v[i]
	}
	return f
}

func TestDot(t *testing.T) {
	user := makeFeatures([]float64{1.0, 2.0, 3.0})
	product := makeFeatures([]float64{3.0, 2.0, 1.0})

	score := user.dot(product)
	ensure.DeepEqual(t, score, 10.0)
}
