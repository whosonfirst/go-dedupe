package flags

type RangeFlag interface {
	MatchesAny(...RangeFlag) bool
	MatchesAll(...RangeFlag) bool
	// Above(...RangeFlag) bool
	// Below(...RangeFlag) bool
	// Between(...RangeFlag) bool
	Min() float64
	Max() float64
}

type Float64RangeFlag struct {
	RangeFlag
	min float64
	max float64
}

func NewFloat64RangeFlag(min float64, max float64) RangeFlag {

	fl := &Float64RangeFlag{
		min: min,
		max: max,
	}

	return fl
}

func (fl *Float64RangeFlag) Min() float64 {
	return fl.min
}

func (fl *Float64RangeFlag) Max() float64 {
	return fl.max
}

func (fl *Float64RangeFlag) MatchesAny(others ...RangeFlag) bool {

	matches := false

	for _, o := range others {

		if o.Max() >= fl.Min() && o.Min() <= o.Max() {
			matches = true
			break
		}
	}

	return matches
}

func (fl *Float64RangeFlag) MatchesAll(others ...RangeFlag) bool {

	matches := 0

	for _, o := range others {

		if o.Max() < fl.Min() || o.Min() > o.Max() {
			break
		} else {
			matches += 1
		}
	}

	if matches != len(others) {
		return false
	}

	return true
}
