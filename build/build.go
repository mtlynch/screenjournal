package build

import (
	"runtime/debug"
	"time"
)

// Time returns the source revision time embedded by Go's VCS stamping.
func Time() time.Time {
	v := buildSetting("vcs.time")
	if v == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}
	}
	return t
}

// Revision returns the short commit hash embedded by Go's VCS stamping.
func Revision() string {
	full := FullRevision()
	if len(full) > 10 {
		return full[:10]
	}
	return full
}

// FullRevision returns the full commit hash embedded by Go's VCS stamping.
func FullRevision() string {
	return buildSetting("vcs.revision")
}

func buildSetting(key string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, s := range info.Settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}
