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

func LastStampOf[T api.Stamp](e api.Envelope) (T, bool) {
	stamps := StampsOf[T](e)
	if len(stamps) > 0 {
		return stamps[len(stamps)-1], true
	}
	var zero T
	return zero, false
}

func FirstStampOf[T api.Stamp](e api.Envelope) (T, bool) {
	stamps := StampsOf[T](e)
	if len(stamps) > 0 {
		return stamps[0], true
	}
	var zero T
	return zero, false
}

func HasStampOf[T api.Stamp](e api.Envelope) bool {
	for _, s := range e.Stamps() {
		if _, ok := s.(T); ok {
			return true
		}
	}
	return false
}
