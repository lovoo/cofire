package cofire

import (
	"math"
	"sync"
)

// Validator is used by the processor to validate each incoming rating.
type Validator interface {
	// Validate validates the prediction given a score.
	Validate(prediction, score float64)
}

// ErrorValidator validates each prediction calculating the root mean square
// error.
type ErrorValidator struct {
	sum   float64
	count int
	m     sync.RWMutex
}

// NewErrorValidator create a new ErrorValidator.
func NewErrorValidator() *ErrorValidator {
	return new(ErrorValidator)
}

// Validate validates the prediction given a score.
func (v *ErrorValidator) Validate(prediction, score float64) {
	e := score - prediction
	v.m.Lock()
	v.sum += math.Pow(e, 2)
	v.count++
	v.m.Unlock()
}

// RMSE returns the current root mean square error.
func (v *ErrorValidator) RMSE() float64 {
	v.m.RLock()
	defer v.m.RUnlock()
	if v.count == 0 {
		return 0.0
	}
	return math.Sqrt(v.sum / float64(v.count))
}

// Count returns the number of values validated.
func (v *ErrorValidator) Count() int {
	v.m.RLock()
	defer v.m.RUnlock()
	return v.count
}

// Reset resets the state of the ErrorValidator and returns the last value.
func (v *ErrorValidator) Reset() float64 {
	v.m.Lock()
	defer v.m.Unlock()
	if v.count == 0 {
		return 0.0
	}
	rmse := math.Sqrt(v.sum / float64(v.count))
	v.sum = 0
	v.count = 0
	return rmse
}
