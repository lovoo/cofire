package cofire

import "sync"

// SGD is a helper to apply stochastic gradient descent.
type SGD struct {
	// Gamma constant for learning speed
	Gamma float64

	// Lambda constant for regularization
	Lambda float64

	// average bias calculated in runtime
	bias   float64
	bsum   float64
	bcount int
	m      sync.RWMutex
}

// NewSGD returns a configured SGD helper.
func NewSGD(gamma, lambda float64) *SGD {
	return &SGD{Gamma: gamma, Lambda: lambda}
}

// Add adds bias to the prediction error. If Add is called multiple times, the average bias
// is computed.
func (s *SGD) Add(bias float64) {
	s.m.Lock()
	s.bsum += bias
	s.bcount++
	s.bias = s.bsum / float64(s.bcount)
	s.m.Unlock()
}

// Bias returns the average bias added to the sgd object.
func (s *SGD) Bias() float64 {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.bias
}

// Error computes the error between the score prediction (with f and o) and the
// real score.
func (s *SGD) Error(f, o *Features, score float64) float64 {
	return score - f.Predict(o, s.Bias())
}

// ApplyError applies the stochastic gradient descent on features f with o and
// error e.
func (s *SGD) ApplyError(f, o *Features, e float64) {
	update := o.mult(e * s.Gamma)
	regularization := f.mult(-s.Lambda * s.Gamma)

	// update features
	f.add(update)
	f.add(regularization)

	// update bias
	f.Bias += s.Gamma * (e - s.Lambda*f.Bias)
}

// Apply applies the stochastic gradient descent on features f with o and a
// score. Apply also adds the score to the bias.
func (s *SGD) Apply(f, o *Features, score float64) {
	s.Add(score)
	e := s.Error(f, o, score)
	s.ApplyError(f, o, e)
}
