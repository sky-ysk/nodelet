package internalversion

import (
	"hit.edu/framework/pkg/apis/meta"
)

func SetListOptionsDefaults(obj *ListOptions, isWatchListFeatureEnabled bool) {
	if !isWatchListFeatureEnabled {
		return
	}
	if obj.SendInitialEvents != nil || len(obj.ResourceVersionMatch) != 0 {
		return
	}
	legacy := obj.ResourceVersion == "" || obj.ResourceVersion == "0"
	if obj.Watch && legacy {
		turnOnInitialEvents := true
		obj.SendInitialEvents = &turnOnInitialEvents
		obj.ResourceVersionMatch = meta.ResourceVersionMatchNotOlderThan
	}
}
