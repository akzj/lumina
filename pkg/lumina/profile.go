package lumina

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// ProfileResult holds profiling data.
type ProfileResult struct {
	TotalFrames  int       `json:"total_frames"`
	AvgFrameTime float64   `json:"avg_frame_time_ms"`
	MinFrameTime float64   `json:"min_frame_time_ms"`
	MaxFrameTime float64   `json:"max_frame_time_ms"`
	SlowFrames   int       `json:"slow_frames_>16ms"`
	FrameTimes   []float64 `json:"frame_times_ms,omitempty"`
}

// FrameTiming tracks timing for a single frame.
type FrameTiming struct {
	Timestamp int64
	Duration  time.Duration
}

// Profile holds performance data.
type Profile struct {
	timings   []FrameTiming
	maxSize   int
	startTime time.Time
}

var globalProfile = &Profile{
	timings: make([]FrameTiming, 0, 1000),
	maxSize: 1000,
}

// RecordFrameTiming records a frame render time.
func RecordFrameTiming(duration time.Duration) {
	timing := FrameTiming{
		Timestamp: time.Now().UnixMilli(),
		Duration:  duration,
	}

	globalProfile.timings = append(globalProfile.timings, timing)

	// Trim if over max size
	if len(globalProfile.timings) > globalProfile.maxSize {
		globalProfile.timings = globalProfile.timings[1:]
	}
}

// ProfileFrames analyzes frame rendering performance.
func ProfileFrames() *ProfileResult {
	result := &ProfileResult{}

	if len(globalProfile.timings) == 0 {
		return result
	}

	result.TotalFrames = len(globalProfile.timings)

	var total float64
	times := make([]float64, len(globalProfile.timings))

	for i, t := range globalProfile.timings {
		ms := t.Duration.Seconds() * 1000
		times[i] = ms
		total += ms

		if ms > 16 {
			result.SlowFrames++
		}
	}

	if result.TotalFrames > 0 {
		result.AvgFrameTime = total / float64(result.TotalFrames)
	}

	// Find min/max
	sort.Float64s(times)
	result.MinFrameTime = times[0]
	result.MaxFrameTime = times[len(times)-1]

	// Include frame times (last 100)
	if len(times) > 100 {
		result.FrameTimes = times[len(times)-100:]
	} else {
		result.FrameTimes = times
	}

	return result
}

// ResetProfile clears profiling data.
func ResetProfile() {
	globalProfile.timings = nil
}

// Lua API

// profile() → JSON string of profile data
func profile(L *lua.State) int {
	result := ProfileFrames()
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		L.PushString(`{"error": "` + err.Error() + `"}`)
		return 1
	}
	L.PushString(string(jsonBytes))
	return 1
}

// profileReset() — clear profiling data
func profileReset(L *lua.State) int {
	ResetProfile()
	return 0
}

// profileSize() → number of recorded frames
func profileSize(L *lua.State) int {
	L.PushInteger(int64(len(globalProfile.timings)))
	return 1
}
