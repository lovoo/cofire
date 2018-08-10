package cofire

import (
	"math"
	"testing"
)

func TestGradientDescent(t *testing.T) {
	user := makeFeatures([]float64{0.5, 0.5, 0.5})
	product := makeFeatures([]float64{0.1, 0.1, 0.1})
	sgd := NewSGD(DefaultParams())

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

func TestSGD_Clip0(t *testing.T) {
	u := &Features{V: []float64{0.5, 0.5, 0.5}}
	p := &Features{V: []float64{0.1, 0.1, 0.1}}
	params := DefaultParams()
	params.Gamma = 1e33
	sgd := NewSGD(params)

	var v float64
	for i := 0; i < 11; i++ {
		sgd.Apply(u, p, -1)
		v = u.Predict(p, 0)
	}
	if !math.IsNaN(v) {
		t.Errorf("v should NaN with Clip 0")
	}
}

func TestSGD_ClipNot0(t *testing.T) {
	u := &Features{V: []float64{0.5, 0.5, 0.5}}
	p := &Features{V: []float64{0.1, 0.1, 0.1}}
	params := DefaultParams()
	params.Gamma = 1e33
	params.Clip = 0.2
	sgd := NewSGD(params)

	var v float64
	for i := 0; i < 11; i++ {
		sgd.Apply(u, p, -1)

		v = u.Predict(p, 0)
	}
	if math.IsNaN(v) {
		t.Errorf("v should not NaN with Clip 0.1")
	}
}
