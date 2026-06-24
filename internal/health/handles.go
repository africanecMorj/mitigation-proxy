package health

import "time"

func (b *Backend) Close() {
	b.cancel()
}

func (b *Backend) StateValue() BackendState {
	return BackendState(b.State.Load())
}

func (b *Backend) SetDraining(v bool) {
	b.draining.Store(v)
}

func (b *Backend) Failures() uint64 {
	return uint64(b.TotalFailures.Load())
}

func (b *Backend) AvgLatency() uint64 {
	return uint64(b.TotalLatency.Load())
}

func (b *Backend) Successes() uint64 {
	return uint64(b.TotalSuccesses.Load())
}

func (b *Backend) SetState(state BackendState) {
	b.State.Store(int32(state))
	b.LastStateChange.Store(time.Now().UnixNano())
}

func (b *Backend) TTFBValue() int64 {
	return int64(b.ttfb.Load())
}

func (b *Backend) SetTTFB(t int64) {
	b.ttfb.Store(t)
}

func (b *Backend) WeightValue() int64 {
	return int64(b.Weight.Load())
}

func (b *Backend) SetWeight(weight int64) {
	b.Weight.Store(weight)
}

func (b *Backend) CurrentWeightValue() int64 {
	return b.currentWeight.Load()
}

func (b *Backend) AddCurrentWeight(delta int64) int64 {
	return b.currentWeight.Add(delta)
}

func (s BackendState) String() string {
	switch s {
	case Healthy:
		return "healthy"

	case Suspect:
		return "suspect"

	case Recovering:
		return "recovering"

	case Unhealthy:
		return "unhealthy"

	case Draining:
		return "draining"

	case Removed:
		return "removed"

	default:
		return "unknown"
	}
}
