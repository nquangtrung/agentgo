package models

type MessageRole string

const (
	MessageRoleHuman     MessageRole = "human"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
	// MessageRoleTool carries the results of tool executions. Anthropic models
	// these as tool_result blocks on a user turn, but they are results rather
	// than user input, so they get their own role here and each provider maps it
	// onto its own wire format.
	MessageRoleTool MessageRole = "tool"
)

// Message is one turn of conversation. Content is a list of typed parts so a
// single message can carry text alongside tool calls or tool results.
type Message interface {
	Type() MessageRole
	Content() BaseMessageContent
}

type BaseMessage struct {
	messageRole MessageRole
	content     BaseMessageContent
}

func (m BaseMessage) Type() MessageRole {
	return m.messageRole
}

func (m BaseMessage) Content() BaseMessageContent {
	return m.content
}

func NewStringMessage(messageRole MessageRole, content string) BaseMessage {
	return NewMessageWithParts(messageRole, NewTextContentPart(content))
}

func NewMessageWithParts(messageRole MessageRole, parts ...MessageContentPart) BaseMessage {
	return BaseMessage{
		messageRole: messageRole,
		content:     NewMessageContent(parts...),
	}
}

func NewHumanStringMessage(content string) BaseMessage {
	return NewStringMessage(MessageRoleHuman, content)
}

func NewAssistantStringMessage(content string) BaseMessage {
	return NewStringMessage(MessageRoleAssistant, content)
}

func NewSystemStringMessage(content string) BaseMessage {
	return NewStringMessage(MessageRoleSystem, content)
}

// NewAssistantToolCallsMessage builds the assistant turn recording the tool
// calls the model requested. Providers that require structured tool protocols
// reject a tool result that does not reference a preceding call, so this
// message must precede the matching results in the conversation.
//
// When text is non-empty the model narrated alongside its calls, so the text is
// carried as a leading text part.
func NewAssistantToolCallsMessage(calls []ToolCall, text string) BaseMessage {
	parts := make([]MessageContentPart, 0, len(calls)+1)
	if text != "" {
		parts = append(parts, NewTextContentPart(text))
	}
	for _, call := range calls {
		parts = append(parts, NewToolCallContentPart(call.ID, call.ToolName, call.Params))
	}

	return NewMessageWithParts(MessageRoleAssistant, parts...)
}

// NewMessageFromToolResult builds the tool turn holding a single execution
// result, correlated to its originating call by ToolCall.ID.
func NewMessageFromToolResult(output ToolExecuteOutput) BaseMessage {
	toolName := "text"
	var toolCallID string
	if output.ToolCall != nil {
		toolName = output.ToolCall.ToolName
		toolCallID = output.ToolCall.ID
	}

	return NewMessageWithParts(
		MessageRoleTool,
		NewToolResultContentPart(toolCallID, toolName, output.Output, output.Error),
	)
}
