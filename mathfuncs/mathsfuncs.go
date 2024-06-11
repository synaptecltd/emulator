package mathfuncs

import (
	"errors"
	"math"
	"math/rand/v2"
)

// A map between string names and functions
var trendFunctions = map[string]func(float64, float64, float64) float64{
	"linear":            linearRamp, // default
	"sine":              sinusoid,
	"cosine":            cosine,
	"exponential":       exponentialRamp,
	"parabolic":         parabolicRamp,
	"step":              stepFunction,
	"square":            squareWave,
	"sawtooth":          sawtoothWave,
	"impulse":           impulseTrain,
	"random_noise":      randomNoise,
	"gaussian_noise":    gaussianNoise,
	"exponential_noise": exponentialNoise,
	"random_walk":       randomWalk,
}

// Returns the named trend function. Defaults to linear if name is empty.
func GetTrendFunctionFromName(name string) (func(float64, float64, float64) float64, error) {
	// Default to linear if no name is provided
	if name == "" {
		return trendFunctions["linear"], nil
	}
	trendFunc, ok := trendFunctions[name]
	if !ok {
		return nil, errors.New("trend function not found")
	}

	return trendFunc, nil
}

// Returns a linear ramp y=(A/T)*t where A is the magntiude of the ramp, T is
// its duration, and t is elapsed time.
func linearRamp(t, A, T float64) float64 {
	m := A / T // slope of the ramp
	return m * t
}

// Returns a sinusoid y=A*sin(2*pi*t/T) where A is the amplitude,
// T is the period, and t is elapsed time.
func sinusoid(t, A, T float64) float64 {
	return A * math.Sin(2*math.Pi*t/T)
}

// Returns a cosine wave y=A*cos(2*pi*t/T) where A is the amplitude,
// T is the period, and t is elapsed time.
func cosine(t, A, T float64) float64 {
	return A * math.Cos(2*math.Pi*t/T)
}

// Returns an exponential ramp y=A*exp(t/T) - A where A is the amplitude,
// T is the time constant, and t is elapsed time.
func exponentialRamp(t, A, T float64) float64 {
	return A*math.Exp(t/T) - A
}

// Returns a parabolic ramp of amplitude A every period T.
func parabolicRamp(t, A, T float64) float64 {
	return A * math.Pow(t/T, 2)
}

// Returns a step function of amplitude A every period T.
func stepFunction(t, A, T float64) float64 {
	if math.Mod(t, T) < T/2 {
		return 0
	} else {
		return A
	}
}

// Returns a square wave y=A if sin(2*pi*t/T) >= 0, else -A.
// where A is the amplitude, T is the period, and t is elapsed time.
func squareWave(t, A, T float64) float64 {
	if math.Sin(2*math.Pi*t/T) >= 0 {
		return A
	} else {
		return -A
	}
}

// Returns a sawtooth wave y=(2*A/pi)*atan(tan(pi*t/T)),
// where A is the amplitude, T is the period, and t is elapsed time.
func sawtoothWave(t, A, T float64) float64 {
	return (2 * A / math.Pi) * math.Atan(math.Tan(math.Pi*t/T))
}

// Returns a spike of amplitude A every period T.
// Each spike has a width of 1 microsecond.
func impulseTrain(t, A, T float64) float64 {
	spikeWidth := 1e-6
	if math.Mod(t, T) < spikeWidth {
		return A
	} else {
		return 0
	}
}

// Returns additional random (uniform) noise of amplitude A.
func randomNoise(_, A, _ float64) float64 {
	return A * (rand.Float64()*2 - 1) // A random number between -A and A
}

// Returns additional Gaussian noise of amplitude A.
func gaussianNoise(_, A, _ float64) float64 {
	return rand.NormFloat64() * A
}

// Returns additional exponential noise of amplitude A.
func exponentialNoise(_, A, _ float64) float64 {
	return -A * math.Log(rand.Float64())
}

// Returns a random walk of amplitude A every period T.
// The returned function is stateful, it remembers the previous value.
// This prevents stack overflow errors that occur with recursive implementations.
var randomWalk = func() func(float64, float64, float64) float64 {
	var previousValue float64 = 0
	return func(t, A, T float64) float64 {
		if t != 0 {
			previousValue += A * (rand.Float64()*2 - 1)
		}
		return previousValue
	}
}()
