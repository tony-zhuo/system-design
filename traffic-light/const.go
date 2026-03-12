package main

type SignalState int

const (
	Red SignalState = iota
	Yellow
	Green
	BlinkingYellow
	BlinkingRed
)

func (s SignalState) String() string {
	switch s {
	case Red:
		return "RED"
	case Yellow:
		return "YELLOW"
	case Green:
		return "GREEN"
	case BlinkingYellow:
		return "BLINKING_YELLOW"
	case BlinkingRed:
		return "BLINKING_RED"
	default:
		return "UNKNOWN"
	}
}

type Direction int

const (
	North Direction = iota
	South
	East
	West
)

func isValidTransition(from, to SignalState) bool {
	allowed := map[SignalState][]SignalState{
		Green:  {Yellow},
		Yellow: {Red},
		Red:    {Green},
		// 緊急模式可以跳過驗證，另外處理
	}
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeEmergency
	ModeManual
)
