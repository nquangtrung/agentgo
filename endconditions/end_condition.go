package endconditions

import (
	"encoding/json"
	"slices"

	"trontria.com/agentgo/models"
)

type MaxStepsEndCondition struct {
	MaxSteps int
}

func (m MaxStepsEndCondition) Condition(context *models.ToolExecutionsArchive) bool {
	if m.MaxSteps <= 0 {
		return true
	}
	return context.RecordCount() >= m.MaxSteps
}

func NewMaxStepsEndCondition(maxSteps int) MaxStepsEndCondition {
	return MaxStepsEndCondition{MaxSteps: maxSteps}
}

type NoProgressEndCondition struct {
	Window int
}

func (n NoProgressEndCondition) Condition(context *models.ToolExecutionsArchive) bool {
	if n.Window <= 0 {
		return false
	}
	records := context.Records()
	if len(records) < n.Window {
		return false
	}

	windowStart := max(len(records)-n.Window, 0)

	uniqueToolNames := map[string]struct{}{}
	for _, record := range records[windowStart:] {
		if record.ToolCalled == nil {
			continue
		}
		uniqueToolNames[record.ToolCalled.ToolName] = struct{}{}
	}
	return len(uniqueToolNames) <= 1
}

func NewNoProgressEndCondition(window int) NoProgressEndCondition {
	return NoProgressEndCondition{Window: window}
}

type ConsecutiveFailureEndCondition struct {
	Failures int
}

func (c ConsecutiveFailureEndCondition) Condition(context *models.ToolExecutionsArchive) bool {
	if c.Failures <= 0 {
		return false
	}
	records := context.Records()
	if len(records) < c.Failures {
		return false
	}

	failCount := 0

	for _, record := range slices.Backward(records) {
		if record.ToolResult == nil || record.ToolResult.Error == nil {
			return false
		}
		failCount++
		if failCount >= c.Failures {
			return true
		}
	}
	return false
}

func NewConsecutiveFailureEndCondition(failures int) ConsecutiveFailureEndCondition {
	return ConsecutiveFailureEndCondition{Failures: failures}
}

type RepeatedStateMutationEndCondition struct {
	Window      int
	Repetitions int
}

func (r RepeatedStateMutationEndCondition) Condition(context *models.ToolExecutionsArchive) bool {
	if r.Window <= 0 || r.Repetitions <= 0 {
		return false
	}
	records := context.Records()
	if len(records) < r.Repetitions {
		return false
	}

	windowStart := max(len(records)-r.Window, 0)

	counts := map[string]int{}
	for _, record := range records[windowStart:] {
		fingerprint := recordFingerprint(record)
		if fingerprint == "" {
			continue
		}
		counts[fingerprint]++
	}
	for _, count := range counts {
		if count >= r.Repetitions {
			return true
		}
	}
	return false
}

func NewRepeatedStateMutationEndCondition(window, repetitions int) RepeatedStateMutationEndCondition {
	return RepeatedStateMutationEndCondition{Window: window, Repetitions: repetitions}
}

type MaxTotalTokensEndCondition struct {
	MaxTokens int64
}

func (m MaxTotalTokensEndCondition) Condition(context *models.ToolExecutionsArchive) bool {
	if m.MaxTokens <= 0 {
		return true
	}
	return totalTokenUsage(context) >= m.MaxTokens
}

func NewMaxTotalTokensEndCondition(maxTokens int64) MaxTotalTokensEndCondition {
	return MaxTotalTokensEndCondition{MaxTokens: maxTokens}
}

func totalTokenUsage(context *models.ToolExecutionsArchive) int64 {
	if context == nil {
		return 0
	}
	var total int64
	for _, record := range context.Records() {
		if record.ToolResult == nil {
			continue
		}
		total += record.ToolResult.Usage.TotalTokens
	}
	return total
}

func recordFingerprint(record models.ToolExecutionRecord) string {
	if record.ToolCalled == nil {
		return ""
	}
	payload := map[string]any{
		"tool":   record.ToolCalled.ToolName,
		"params": record.ToolCalled.Params,
	}
	if record.ToolResult != nil && len(record.ToolResult.Output) > 0 {
		payload["output"] = record.ToolResult.Output
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return record.ToolCalled.ToolName
	}
	return string(encoded)
}
