package endconditions

import (
	"errors"
	"testing"

	"trontria.com/agentgo/models"
)

func makeArchive(records ...models.ToolExecutionRecord) *models.ToolExecutionsArchive {
	archive := models.NewExecutionContextFromLanguageModelContext(models.LanguageModelContext{ModelName: "test"})
	for _, record := range records {
		if record.ToolResult != nil {
			archive.AddToolCallWithResult(record.Name, record.ToolResult)
			continue
		}
		archive.AddToolCall(record.Name, record.ToolCalled)
	}
	return archive
}

func TestMaxStepsEndCondition(t *testing.T) {
	archive := makeArchive(models.ToolExecutionRecord{Name: "first", ToolCalled: &models.ToolCall{ToolName: "tool"}})
	if !NewMaxStepsEndCondition(1).Condition(archive) {
		t.Fatal("expected max steps condition to trigger")
	}
	if NewMaxStepsEndCondition(2).Condition(archive) {
		t.Fatal("expected max steps condition to wait for threshold")
	}
}

func TestNoProgressEndCondition(t *testing.T) {
	archive := makeArchive(
		models.ToolExecutionRecord{Name: "a", ToolCalled: &models.ToolCall{ToolName: "read_file"}},
		models.ToolExecutionRecord{Name: "b", ToolCalled: &models.ToolCall{ToolName: "read_file"}},
		models.ToolExecutionRecord{Name: "c", ToolCalled: &models.ToolCall{ToolName: "read_file"}},
	)
	if !NewNoProgressEndCondition(3).Condition(archive) {
		t.Fatal("expected no-progress condition to trigger on repeated tool usage")
	}
}

func TestConsecutiveFailureEndCondition(t *testing.T) {
	archive := makeArchive(
		models.ToolExecutionRecord{Name: "a", ToolResult: &models.ToolExecuteOutput{Error: errors.New("nope")}},
		models.ToolExecutionRecord{Name: "b", ToolResult: &models.ToolExecuteOutput{Error: errors.New("still nope")}},
		models.ToolExecutionRecord{Name: "c", ToolResult: &models.ToolExecuteOutput{Error: errors.New("still nope")}},
	)
	if !NewConsecutiveFailureEndCondition(3).Condition(archive) {
		t.Fatal("expected consecutive failure condition to trigger")
	}
}

func TestRepeatedStateMutationEndCondition(t *testing.T) {
	archive := makeArchive(
		models.ToolExecutionRecord{Name: "a", ToolCalled: &models.ToolCall{ToolName: "write_file", Params: map[string]any{"path": "/tmp/a.txt"}}},
		models.ToolExecutionRecord{Name: "b", ToolCalled: &models.ToolCall{ToolName: "write_file", Params: map[string]any{"path": "/tmp/a.txt"}}},
		models.ToolExecutionRecord{Name: "c", ToolCalled: &models.ToolCall{ToolName: "write_file", Params: map[string]any{"path": "/tmp/a.txt"}}},
	)
	if !NewRepeatedStateMutationEndCondition(3, 2).Condition(archive) {
		t.Fatal("expected repeated state mutation condition to trigger")
	}
}

func TestMaxTotalTokensEndCondition(t *testing.T) {
	archive := makeArchive(
		models.ToolExecutionRecord{Name: "a", ToolResult: &models.ToolExecuteOutput{Usage: models.LanguageModelUsage{TotalTokens: 100}}},
		models.ToolExecutionRecord{Name: "b", ToolResult: &models.ToolExecuteOutput{Usage: models.LanguageModelUsage{TotalTokens: 250}}},
	)
	if !NewMaxTotalTokensEndCondition(300).Condition(archive) {
		t.Fatal("expected token budget condition to trigger once total usage reaches the threshold")
	}
	if NewMaxTotalTokensEndCondition(400).Condition(archive) {
		t.Fatal("expected token budget condition to wait for threshold")
	}
}
