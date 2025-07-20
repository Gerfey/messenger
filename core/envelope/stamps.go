package envelope

import (
	"github.com/gerfey/messenger/api"
)

func StampsOf[T api.Stamp](e api.Envelope) []T {
	var filtered []T
	for _, s := range e.Stamps() {
		if stamp, ok := s.(T); ok {
			filtered = append(filtered, stamp)
		}
	}

	return filtered
}

func FirstStampOf[T api.Stamp](e api.Envelope) (T, bool) {
	var zero T
	stamps := e.Stamps()
	for _, stamp := range stamps {
		if s, ok := stamp.(T); ok {
			return s, true
		}
	}

	return zero, false
}

func LastStampOf[T api.Stamp](e api.Envelope) (T, bool) {
	var zero T
	stamps := e.Stamps()
	for i := len(stamps) - 1; i >= 0; i-- {
		if s, ok := stamps[i].(T); ok {
			return s, true
		}
	}

	return zero, false
}

func HasStampOf[T api.Stamp](e api.Envelope) bool {
	stamps := e.Stamps()
	for _, stamp := range stamps {
		if _, ok := stamp.(T); ok {
			return true
		}
	}

	return false
}
