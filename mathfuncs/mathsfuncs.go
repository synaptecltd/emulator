package mathfuncs

import (
	"errors"
	"math"
	"math/rand/v2"

	"github.com/stevenblair/sigourney/fast"
)

// A mathematical function y=f(t,A,T). Takes amplitude, A, and period, T,
// as inputs and returns the value of the function at time, t.
type MathsFunction func(t, A, T float64) float64

// A map between string name and trendFunction pairs
var mathsFunctions = map[string]MathsFunction{
	"linear": linearRamp,
	// "sine": func(t, A, T float64) float64 {
	// return Sine(t, A, T)
	// }
	"sine":                   Sine,
	"cosine":                 cosineWave,
	"exponential":            exponentialRamp,
	"exponential_full":       exponentialRampSaturated,
	"exponential_decay":      exponentialDecay,
	"exponential_decay_full": exponentialDecaySaturated,
	"parabolic":              parabolicRamp,
	"step":                   stepFunction,
	"Lstep":                  LstepFunction,
	"square":                 squareWave,
	"sawtooth":               sawtoothWave,
	"impulse":                impulseTrain,
	"impulse_varying":        impulseTrainVaryingMagnitude,
	"random_noise":           randomNoise,
	"gaussian_noise":         gaussianNoise,
	"exponential_noise":      exponentialNoise,
	"random_walk":            randomWalk,
	"flat":                   flat,
	"warmup_sine":            func(t, A, T float64) float64 { return warmup_sine(t, A, T, 0) },
}

func GetMathsFunctionNames() []string {
	names := make([]string, 0, len(mathsFunctions))
	for name := range mathsFunctions {
		names = append(names, name)
	}
	return names
}

// Returns the named trend function. Defaults to linear if name is empty.
func GetTrendFunctionFromName(name string) (MathsFunction, error) {
	trendFunc, ok := mathsFunctions[name]
	if !ok {
		return nil, errors.New("trend function not found")
	}

	return trendFunc, nil
}

// Returns a linear ramp y=(A/T)*t where A is the magnitude of the ramp, T is
// its duration, and t is elapsed time.
func linearRamp(t, A, T float64) float64 {
	m := A / T // slope of the ramp
	return m * t
}

// Returns a sine wave y = A*sin(2π * t / PeriodDuration)
// PeriodDuration defines the cycle length in seconds.
func Sine(t, A, PeriodDuration float64) float64 {
	if PeriodDuration <= 0 {
		PeriodDuration = 86400.0 // default to 1 day
	}
	return A * math.Sin(2*math.Pi*t/PeriodDuration)
}

// Returns a cosine wave y=A*cos(2*pi*t/T) where A is the amplitude,
// T is the period, and t is elapsed time.
func cosineWave(t, A, T float64) float64 {
	return A * fast.Cos(2*math.Pi*t/T)
}

// Returns an exponential ramp y=A*exp(t/T) - A where A is the amplitude,
// T is the time constant, and t is elapsed time.
func exponentialRamp(t, A, T float64) float64 {
	return A*math.Exp(t/T) - A
}

// Returns an exponential ramp y=A*exp(5*t/T) - A where A is the amplitude,
// T is the time constant, and t is elapsed time.
func exponentialRampSaturated(t, A, T float64) float64 {
	return A*math.Exp(5*t/T) - A
}

// Returns an exponential decay y=A*exp(-t/T) where A is the amplitude,
// T is the time constant, and t is elapsed time.
func exponentialDecay(t, A, T float64) float64 {
	return A * math.Exp(-t/T)
}

// Returns an exponential decay y=A*exp(-t/T) where A is the amplitude,
// T is the time constant, and t is elapsed time.
func exponentialDecaySaturated(t, A, T float64) float64 {
	return A * (1 - math.Exp(-t/T))
}

// Returns a parabolic ramp of amplitude A every period T.
func parabolicRamp(t, A, T float64) float64 {
	return A * (t / T) * (t / T) // faster power of two compared to math.Pow(t/T, 2)
}

// Returns a step function of amplitude A every period T.
func stepFunction(t, A, T float64) float64 {
	if math.Mod(t, T) < T/2 {
		return 0
	} else {
		return A
	}
}

// LstepFunction: creates a one-time downward step that stays flat afterward.
// Produces an 'L' shape — a small drop followed by a flat line.
func LstepFunction(t, A, T float64) float64 {
	if t >= 0 {
		return -A // step down
	}
	return 0
}

// Returns a square wave y=A if sin(2*pi*t/T) >= 0, else -A.
// where A is the amplitude, T is the period, and t is elapsed time.
func squareWave(t, A, T float64) float64 {
	if fast.Sin(2*math.Pi*t/T) >= 0 {
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

// Returns a spike every period T, with an amplitude which is
// normally distributed about A. Each spike has a width of 1 microsecond.
func impulseTrainVaryingMagnitude(t, A, T float64) float64 {
	fixedAmplitudeImpulse := impulseTrain(t, A, T)
	return fixedAmplitudeImpulse * rand.NormFloat64()
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

// WarmupTemp generates a refined sinusoidal warm-up pattern with configurable period and amplitude.
// t: elapsed time in seconds
// A: amplitude of oscillation (e.g., ±1.5°C)
// period: oscillation period (seconds) — typically 3600 for 1-hour cycles
// base: baseline temperature (°C)
func warmup_sine(t, A, period, _ float64) float64 {
	// Primary sine: smooth oscillation with one cycle per `period`
	primary := A * math.Sin(2.0*math.Pi*t/period)

	// Secondary small wave: adds subtle natural variation (higher frequency)
	secondary := 0.5 * A * math.Sin(6.0*math.Pi*t/period)

	// Combine and return
	return primary + secondary
}

// flat returns a constant value equal to A (amplitude),
// independent of time t or period T.
func flat(t, A, T float64) float64 {
	return A
}

// Returns a random walk that lasts for period T. The walk is bounded
// to within +/- amplitude A, and can make steps of maximum size A/20.
// The returned function is stateful, it remembers the previous value.
// This prevents stack overflow errors that occur with recursive implementations.
var randomWalk = func() func(float64, float64, float64) float64 {
	stepFactor := 20.0
	var previousValue float64 = 0
	return func(t, A, T float64) float64 {
		if t != 0 {
			step := A / stepFactor * (rand.Float64()*2 - 1)
			proposedValue := previousValue + step

			// Hold the value within the bounds of +/- A
			previousValue = math.Min(
				math.Max(proposedValue, -A),
				A)
		}
		return previousValue
	}
}()
