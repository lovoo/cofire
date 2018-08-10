package cofire

// Parameters configure the SGD algorithm.
type Parameters struct {
	// Rank is the number of latent factors (features).
	Rank int
	// Gamma is the learning step.
	Gamma float64
	// Lambda is the regularization parameter.
	Lambda float64
	// Iterations is the number of times the data will be used for training.
	Iterations int
	// Clip defines the maximum absolute error to be applied.
	// If 0 the error is not limited. Default is 0.
	Clip float64
}

// DefaultParams return the default parameters of SGD.
func DefaultParams() Parameters {
	return Parameters{10, 0.01, 0.001, 1, 0}
}
