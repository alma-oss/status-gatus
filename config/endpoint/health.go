package endpoint

// Health is the three-state reading of a Result derived from its success flag and its condition results:
// a successful result whose conditions all passed is HEALTHY, a successful result with at least one failed
// condition is HEALTHY_WITH_WARNINGS (the pusher decided the endpoint is up but flagged something), and a
// failed result is UNHEALTHY.
type Health string

const (
	HealthHealthy             Health = "HEALTHY"
	HealthHealthyWithWarnings Health = "HEALTHY_WITH_WARNINGS"
	HealthUnhealthy           Health = "UNHEALTHY"
)

// ComputeHealth derives the Health of a result
func (r *Result) ComputeHealth() Health {
	if !r.Success {
		return HealthUnhealthy
	}
	for _, conditionResult := range r.ConditionResults {
		if conditionResult != nil && !conditionResult.Success {
			return HealthHealthyWithWarnings
		}
	}
	return HealthHealthy
}

// WithHealth sets the derived Health on every result and on the status itself (the health of the latest result).
// Results are stored oldest first, so the latest one is the last element.
func (s *Status) WithHealth() *Status {
	if s == nil {
		return nil
	}
	for _, result := range s.Results {
		if result != nil {
			result.Health = result.ComputeHealth()
		}
	}
	if len(s.Results) > 0 && s.Results[len(s.Results)-1] != nil {
		s.Health = s.Results[len(s.Results)-1].Health
	} else {
		s.Health = ""
	}
	return s
}
