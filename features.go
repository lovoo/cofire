package cofire

import (
	"math/rand"
)

// NewFeatures creates a feature vector with rank features.
func NewFeatures(rank int) *Features {
	return &Features{
		V: make([]float64, rank),
	}
}

func (f *Features) clone() *Features {
	o := NewFeatures(len(f.V))
	copy(o.V, f.V)
	o.Bias = f.Bias
	return o
}

// Randomize initializes the feature vector with random numbers
func (f *Features) Randomize() *Features {
	for i := range f.V {
		f.V[i] = rand.Float64()
	}
	return f
}

// predict predicts a r^ for a given user and a product
func (f *Features) dot(o *Features) float64 {
	score := 0.0
	n := len(f.V)
	if len(o.V) < n {
		n = len(o.V)
	}
	for i := 0; i < n; i++ {
		score += f.V[i] * o.V[i]
	}
	return score
}

// mult multiplies the features by a scalar and returns the resulting vector
func (f *Features) mult(scalar float64) []float64 {
	r := make([]float64, len(f.V))
	for i := range f.V {
		r[i] = f.V[i] * scalar
	}
	return r
}

// add adds a vector to the features
func (f *Features) add(v []float64) {
	n := len(f.V)
	if len(v) < n {
		n = len(v)
	}
	for i := 0; i < n; i++ {
		f.V[i] += v[i]
	}
}

// Predict predicts a r^ for a given user and a product
func (f *Features) Predict(o *Features, bias float64) float64 {
	return bias + f.Bias + o.Bias + f.dot(o)
}
