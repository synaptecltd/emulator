package emulator

import "math"

// Returns the magntiude of a linear ramp y=(A/T)*t where A is the magntiude of the ramp, T is
// its duration, and t is elapsed time.
func linearRamp(t float64, A float64, T float64) float64 {
	m := A / T // slope of the ramp
	return m * t
}

// Returns the magnitude of a sinusoid y=A*sin(2*pi*t/T) where A is the amplitude,
// T is the period, and t is elapsed time.
func sinusoid(t float64, A float64, T float64) float64 {
	return A * math.Sin(2*math.Pi*t/T)
}

// Returns the magnitude of an exponential ramp y=A*exp(t/T) - A where A is the amplitude,
// T is the time constant, and t is elapsed time.
func exponentialRamp(t, A, T float64) float64 {
	return A*math.Exp(t/T) - A
}

// Returns the magnitude of a square wave y=A if sin(2*pi*t/T) >= 0, else -A.
// where A is the amplitude, T is the period, and t is elapsed time.
func squareWave(t, A, T float64) float64 {
	if math.Sin(2*math.Pi*t/T) >= 0 {
		return A
	} else {
		return -A
	}
}

// Returns the magnitude of a sawtooth wave y=(2*A/pi)*atan(tan(pi*t/T)),
// where A is the amplitude, T is the period, and t is elapsed time.
func sawtoothWave(t, A, T float64) float64 {
	return (2 * A / math.Pi) * math.Atan(math.Tan(math.Pi*t/T))
}
