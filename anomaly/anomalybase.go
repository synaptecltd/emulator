package anomaly

import (
	"errors"

	"github.com/google/uuid"
	"github.com/synaptecltd/emulator/mathfuncs"
)

// AnomalyBase is the base struct for all anomaly types.
type AnomalyBase struct {
	Uuid   uuid.UUID // identifier for the anomaly, store as string for compatibility with yaml
	Repeat uint64    // the number of times the anomalies repeat, 0 for infinite
	Off    bool      // true: anomaly deactivated, false: activated

	// Setters with error checking should be provided for private fields below
	typeName   string  // the type of anomaly as a string, e.g. "trend", "spike".
	startDelay float64 // the delay before anomalies begin (and between anomaly repeats) in seconds
	duration   float64 // the duration of anomaly each anomaly repeat in seconds

	// internal state
	isAnomalyActive       bool    // whether the anomaly is actively modulating the waveform in this timestep
	startDelayIndex       int     // startDelay converted to time steps, used to track delay period between anomaly repeats
	elapsedActivatedIndex int     // number of time steps since start of this active anomaly repeat, used to track the progress within an anomaly burst/trend
	elapsedActivatedTime  float64 // time elapsed since the start of this active anomaly repeat
	countRepeats          uint64  // counter for number of times the anomaly trend/burst has repeated
}

// Returns the type of anomaly as a string.
func (a *AnomalyBase) GetTypeAsString() string {
	return a.typeName
}

// Returns the uuid of the anomaly.
func (a *AnomalyBase) GetUuid() uuid.UUID {
	return a.Uuid
}

// Sets the uuid of the anomaly to uuid.
func (a *AnomalyBase) SetUuid(uuidId uuid.UUID) {
	a.Uuid = uuidId
}

// Set the uuid of the anomaly from a string representation. Returns an error if the string is not a valid uuid.
// If the string is empty, the uuid is set to 00000000-0000-0000-0000-000000000000.
func (a *AnomalyBase) SetUuidFromString(uuidString string) error {
	if uuidString == "" {
		a.Uuid = uuid.Nil
		return nil
	}
	uuidId, err := uuid.Parse(uuidString)
	if err != nil {
		return err
	}
	a.Uuid = uuidId
	return nil
}

// Returns the start delay of anomaly in seconds
func (a *AnomalyBase) GetStartDelay() float64 {
	return a.startDelay
}

// Returns the duration of the anomaly in seconds.
func (a *AnomalyBase) GetDuration() float64 {
	return a.duration
}

// Returns whether the anomaly is actively actuating the waveform output in this timestep.
func (a *AnomalyBase) GetIsAnomalyActive() bool {
	return a.isAnomalyActive
}

// Returns the start delay of the anomaly as a number of time steps.
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
	moreRepeatsAllowed := a.countRepeats < a.Repeat || a.Repeat == 0 // 0 means infinite repetitions
	if !moreRepeatsAllowed {
		a.Off = true // switch the anomaly off if all repetitions are complete to save future computation
		return false
	}

	hasAnomalyStarted := a.startDelayIndex >= int(a.startDelay/Ts)-1
	return hasAnomalyStarted
}

// Set the fields funcName and funcVar of an anomaly by looking up a function name.
func (a *AnomalyBase) SetFunctionByName(name string, funcSetter func(string) (mathfuncs.MathsFunction, error), funcName *string, funcVar *mathfuncs.MathsFunction) error {
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
