package main

import (
	"testing"
	"time"

	"github.com/TwiN/gatus/v5/alerting/alert"
	"github.com/TwiN/gatus/v5/config/endpoint"
	"github.com/TwiN/gatus/v5/storage"
	"github.com/TwiN/gatus/v5/storage/store"
)

// endpointWithAlert returns an endpoint carrying one enabled alert with the given thresholds; the counter restore
// only runs for endpoints that have enabled alerts and loads at most max(failureThreshold, successThreshold) results.
func endpointWithAlert(name string, failureThreshold, successThreshold int) *endpoint.Endpoint {
	enabled := true
	return &endpoint.Endpoint{
		Name:  name,
		Group: "group",
		Alerts: []*alert.Alert{{
			Type:             alert.TypeCustom,
			Enabled:          &enabled,
			FailureThreshold: failureThreshold,
			SuccessThreshold: successThreshold,
		}},
	}
}

func TestRestoreNumberOfEvaluationsInARow(t *testing.T) {
	if err := store.Initialize(nil); err != nil {
		t.Fatal("failed to initialize store:", err.Error())
	}
	defer store.Get().Close()
	ep := endpointWithAlert("endpoint", 3, 3)
	now := time.Now()
	// Simulate two failures in a row that haven't reached an alert's failure threshold yet, followed by a reload:
	// without restoring the in-a-row counters from the persisted results, this progress would be lost.
	for i := 0; i < 2; i++ {
		result := &endpoint.Result{Success: false, Timestamp: now.Add(time.Duration(i) * time.Minute)}
		if err := store.Get().InsertEndpointResult(ep, result); err != nil {
			t.Fatal("failed to insert endpoint result:", err.Error())
		}
	}
	// Simulate a configuration reload creating a brand-new Endpoint struct with the counters reset to 0.
	reloadedEndpoint := endpointWithAlert("endpoint", 3, 3)
	restoreNumberOfEvaluationsInARow(reloadedEndpoint, storage.DefaultMaximumNumberOfResults)
	if reloadedEndpoint.NumberOfFailuresInARow != 2 {
		t.Errorf("expected NumberOfFailuresInARow to be restored to 2, got %d", reloadedEndpoint.NumberOfFailuresInARow)
	}
	if reloadedEndpoint.NumberOfSuccessesInARow != 0 {
		t.Errorf("expected NumberOfSuccessesInARow to remain 0, got %d", reloadedEndpoint.NumberOfSuccessesInARow)
	}
}

func TestRestoreNumberOfEvaluationsInARowWithTrailingSuccesses(t *testing.T) {
	if err := store.Initialize(nil); err != nil {
		t.Fatal("failed to initialize store:", err.Error())
	}
	defer store.Get().Close()
	ep := endpointWithAlert("endpoint", 3, 3)
	now := time.Now()
	results := []bool{false, false, false, true, true}
	for i, success := range results {
		result := &endpoint.Result{Success: success, Timestamp: now.Add(time.Duration(i) * time.Minute)}
		if err := store.Get().InsertEndpointResult(ep, result); err != nil {
			t.Fatal("failed to insert endpoint result:", err.Error())
		}
	}
	reloadedEndpoint := endpointWithAlert("endpoint", 3, 3)
	restoreNumberOfEvaluationsInARow(reloadedEndpoint, storage.DefaultMaximumNumberOfResults)
	if reloadedEndpoint.NumberOfSuccessesInARow != 2 {
		t.Errorf("expected NumberOfSuccessesInARow to be restored to 2, got %d", reloadedEndpoint.NumberOfSuccessesInARow)
	}
	if reloadedEndpoint.NumberOfFailuresInARow != 0 {
		t.Errorf("expected NumberOfFailuresInARow to remain 0, got %d", reloadedEndpoint.NumberOfFailuresInARow)
	}
}

func TestRestoreNumberOfEvaluationsInARowWithNoHistory(t *testing.T) {
	if err := store.Initialize(nil); err != nil {
		t.Fatal("failed to initialize store:", err.Error())
	}
	defer store.Get().Close()
	ep := endpointWithAlert("unknown-endpoint", 3, 3)
	restoreNumberOfEvaluationsInARow(ep, storage.DefaultMaximumNumberOfResults)
	if ep.NumberOfFailuresInARow != 0 || ep.NumberOfSuccessesInARow != 0 {
		t.Errorf("expected counters to remain 0 when there's no persisted history, got failures=%d successes=%d", ep.NumberOfFailuresInARow, ep.NumberOfSuccessesInARow)
	}
}

func TestRestoreNumberOfEvaluationsInARowSkipsEndpointsWithoutAlerts(t *testing.T) {
	if err := store.Initialize(nil); err != nil {
		t.Fatal("failed to initialize store:", err.Error())
	}
	defer store.Get().Close()
	ep := &endpoint.Endpoint{Name: "endpoint", Group: "group"}
	now := time.Now()
	for i := 0; i < 2; i++ {
		result := &endpoint.Result{Success: false, Timestamp: now.Add(time.Duration(i) * time.Minute)}
		if err := store.Get().InsertEndpointResult(ep, result); err != nil {
			t.Fatal("failed to insert endpoint result:", err.Error())
		}
	}
	reloadedEndpoint := &endpoint.Endpoint{Name: "endpoint", Group: "group"}
	restoreNumberOfEvaluationsInARow(reloadedEndpoint, storage.DefaultMaximumNumberOfResults)
	if reloadedEndpoint.NumberOfFailuresInARow != 0 || reloadedEndpoint.NumberOfSuccessesInARow != 0 {
		t.Errorf("expected counters to stay 0 for an endpoint without alerts (nothing to restore for), got failures=%d successes=%d", reloadedEndpoint.NumberOfFailuresInARow, reloadedEndpoint.NumberOfSuccessesInARow)
	}
}

func TestRestoreNumberOfEvaluationsInARowIsBoundedByAlertThreshold(t *testing.T) {
	if err := store.Initialize(nil); err != nil {
		t.Fatal("failed to initialize store:", err.Error())
	}
	defer store.Get().Close()
	ep := endpointWithAlert("endpoint", 2, 1)
	now := time.Now()
	for i := 0; i < 5; i++ {
		result := &endpoint.Result{Success: false, Timestamp: now.Add(time.Duration(i) * time.Minute)}
		if err := store.Get().InsertEndpointResult(ep, result); err != nil {
			t.Fatal("failed to insert endpoint result:", err.Error())
		}
	}
	reloadedEndpoint := endpointWithAlert("endpoint", 2, 1)
	restoreNumberOfEvaluationsInARow(reloadedEndpoint, storage.DefaultMaximumNumberOfResults)
	// Only the largest threshold (2) results are loaded: enough to know the threshold is reached, no more.
	if reloadedEndpoint.NumberOfFailuresInARow != 2 {
		t.Errorf("expected NumberOfFailuresInARow to be bounded by the alert threshold (2), got %d", reloadedEndpoint.NumberOfFailuresInARow)
	}
}
