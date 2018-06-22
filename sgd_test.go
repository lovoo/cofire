package cofire

import (
	"testing"
)

func TestGradientDescent(t *testing.T) {
	user := makeFeatures([]float64{0.5, 0.5, 0.5})
	product := makeFeatures([]float64{0.1, 0.1, 0.1})
	sgd := SGD{Gamma: DefaultParams().Gamma, Lambda: DefaultParams().Lambda}

	cuser := user.clone()
	sgd.Apply(cuser, product, -1)
	a := cuser.dot(product)

	cuser = user.clone()
	sgd.Apply(cuser, product, 1)
	b := cuser.dot(product)

	if a >= b {
		t.Errorf("a >= b (%f >= %f)", a, b)
	}
}
