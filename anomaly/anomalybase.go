package anomaly

import (
	"errors"

	"github.com/synaptecltd/emulator/mathfuncs"
)

// AnomalyBase is the base struct for all anomaly types.
type AnomalyBase struct {
	// Setters and getters are provided for private fields below to allow for error checking
	typeName   string  // the type of anomaly
	startDelay float64 // how many seconds before the anomalies begin
	duration   float64 // the duration the anomalies occur for in seconds
	Repeats    uint64  // the number of times the anomalies repeat, 0 for infinite
	Off        bool    // true: anomaly deactivated, false: activated

	// internal state
	isAnomalyActive       bool    // whether the anomaly is actively modulating the waveform in this timestep
	startDelayIndex       int     // startDelay converted to time steps, used to track delay period between anomaly repeats
	elapsedActivatedIndex int     // number of time steps since start of active anomaly trend/burst, used to track the progress of anomaly trends/bursts
	elapsedActivatedTime  float64 // time elapsed since the start of the active anomaly trend/burst
	countRepeats          uint64  // counter for number of times the anomaly trend/burst has repeated
}

// Returns the type of anomaly as a string.
func (a *AnomalyBase) GetTypeAsString() string {
	return a.typeName
}

// Returns the start delay of anomaly trend/burst in seconds.
func (a *AnomalyBase) GetStartDelay() float64 {
	return a.startDelay
}

// Returns the duration of each anomaly trend/burst in seconds.
func (a *AnomalyBase) GetDuration() float64 {
	return a.duration
}

// Returns whether the anomaly is actively actuating the waveform output in this timestep.
func (a *AnomalyBase) GetIsAnomalyActive() bool {
	return a.isAnomalyActive
}

// Returns a.startDelay as a number of time steps.
func (a *AnomalyBase) GetStartDelayIndex() int {
	return a.startDelayIndex
}

// Returns the number of time steps since the start of the active anomaly trend/burst.
func (a *AnomalyBase) GetElapsedActivatedIndex() int {
	return a.elapsedActivatedIndex
}

// Returns the time elapsed since the start of the active anomaly trend/burst.
func (a *AnomalyBase) GetElapsedActivatedTime() float64 {
	return a.elapsedActivatedTime
}

// Returns the number of times the anomaly trend/burst has repeated so far.
func (a *AnomalyBase) GetCountRepeats() uint64 {
	return a.countRepeats
}

// Sets the start time of anomalies in seconds if delay >= 0.
func (a *AnomalyBase) SetStartDelay(startDelay float64) error {
	if startDelay < 0 {
		return errors.New("startDelay must be greater than or equal to 0")
	}

	a.startDelay = startDelay
	return nil
}

// Returns whether anomalies should be active this timestep. This is true if:
//  1. Enough time has elapsed for the anomaly to start, and;
//  2. The anomaly has not yet completed all repetitions.
func (a *AnomalyBase) CheckAnomalyActive(Ts float64) bool {
	moreRepeatsAllowed := a.countRepeats < a.Repeats || a.Repeats == 0 // 0 means infinite repetitions
	if !moreRepeatsAllowed {
		a.Off = true // switch the anomaly off if all repetitions are complete to save future computation
		return false
	}

	hasAnomalyStarted := a.startDelayIndex >= int(a.startDelay/Ts)-1
	return hasAnomalyStarted
}

// Set the fields funcName and funcVar of an anomaly by looking up the name of the function.
func (a *AnomalyBase) SetFunctionByName(name string, funcSetter func(string) (mathfuncs.TrendFunction, error), funcName *string, funcVar *mathfuncs.TrendFunction) error {
	if name == "" {
		*funcName = name
		*funcVar = nil
		return nil
	}

	trendFunc, err := funcSetter(name)
	if err != nil {
		return err
	}
	*funcVar = trendFunc
	*funcName = name
	return nil
}
