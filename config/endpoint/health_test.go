package endpoint

import "testing"

func TestResult_ComputeHealth(t *testing.T) {
	scenarios := []struct {
		name     string
		result   *Result
		expected Health
	}{
		{name: "failed", result: &Result{Success: false}, expected: HealthUnhealthy},
		{name: "failed-with-conditions", result: &Result{Success: false, ConditionResults: []*ConditionResult{{Condition: "lag 900 exceeds max 500", Success: false}}}, expected: HealthUnhealthy},
		{name: "healthy-no-conditions", result: &Result{Success: true}, expected: HealthHealthy},
		{name: "healthy-all-conditions-pass", result: &Result{Success: true, ConditionResults: []*ConditionResult{{Condition: "Idle (lag 12 <= 500)", Success: true}}}, expected: HealthHealthy},
		{name: "healthy-with-warning", result: &Result{Success: true, ConditionResults: []*ConditionResult{{Condition: "[STATUS] == 200", Success: true}, {Condition: "Warning: certificate expires in 3 days", Success: false}}}, expected: HealthHealthyWithWarnings},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if health := scenario.result.ComputeHealth(); health != scenario.expected {
				t.Errorf("expected %s, got %s", scenario.expected, health)
			}
		})
	}
}

func TestStatus_WithHealth(t *testing.T) {
	status := NewStatus("group", "name")
	if status.WithHealth().Health != "" {
		t.Errorf("expected no health without results, got %s", status.Health)
	}
	status.Results = append(status.Results,
		&Result{Success: false},
		&Result{Success: true, ConditionResults: []*ConditionResult{{Condition: "warning", Success: false}}},
	)
	status.WithHealth()
	if status.Results[0].Health != HealthUnhealthy || status.Results[1].Health != HealthHealthyWithWarnings {
		t.Errorf("expected per-result health to be set, got %s and %s", status.Results[0].Health, status.Results[1].Health)
	}
	if status.Health != HealthHealthyWithWarnings {
		t.Errorf("expected the status health to follow the latest result, got %s", status.Health)
	}
	var nilStatus *Status
	if nilStatus.WithHealth() != nil {
		t.Error("expected nil status to stay nil")
	}
}
