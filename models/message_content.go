package models

import (
	"encoding/json"
	"strings"

	"github.com/nquangtrung/agentgo/utils"
)

// MessageContentPartType discriminates the parts a message can carry.
type MessageContentPartType string

const (
	// MessageContentPartText is plain text typed by a user or the model.
	MessageContentPartText MessageContentPartType = "text"
	// MessageContentPartToolCall is a tool invocation requested by the model.
	MessageContentPartToolCall MessageContentPartType = "tool_call"
	// MessageContentPartToolResult is the outcome of executing a tool call.
	MessageContentPartToolResult MessageContentPartType = "tool_result"
)

// MessageContentPart is one piece of a message's content. Providers that
// require structured tool protocols need to distinguish a tool call from a tool
// result and correlate them by the provider-assigned call id, which a single
// text blob cannot express.
type MessageContentPart interface {
	PartType() MessageContentPartType
}

// TextContentPart is plain text.
type TextContentPart struct {
	text string
}

func NewTextContentPart(text string) TextContentPart {
	return TextContentPart{text: text}
}

func (p TextContentPart) PartType() MessageContentPartType {
	return MessageContentPartText
}

func (p TextContentPart) Text() string {
	return p.text
}

// ToolCallContentPart is a tool invocation requested by the model. ID carries
// the provider-assigned call id that the matching ToolResultContentPart must
// echo back; it is empty when a provider did not supply one.
type ToolCallContentPart struct {
	id    string
	name  string
	input map[string]any
}

func NewToolCallContentPart(id, name string, input map[string]any) ToolCallContentPart {
	return ToolCallContentPart{id: id, name: name, input: input}
}

func (p ToolCallContentPart) PartType() MessageContentPartType {
	return MessageContentPartToolCall
}

func (p ToolCallContentPart) ID() string            { return p.id }
func (p ToolCallContentPart) Name() string          { return p.name }
func (p ToolCallContentPart) Input() map[string]any { return p.input }

// ToolResultContentPart is the outcome of executing a tool call. ToolCallID
// echoes the id from the originating ToolCallContentPart so providers that pair
// calls with results can do so.
type ToolResultContentPart struct {
	toolCallID string
	name       string
	output     map[string]any
	err        error
}

func NewToolResultContentPart(toolCallID, name string, output map[string]any, err error) ToolResultContentPart {
	return ToolResultContentPart{
		toolCallID: toolCallID,
		name:       name,
		output:     output,
		err:        err,
	}
}

func (p ToolResultContentPart) PartType() MessageContentPartType {
	return MessageContentPartToolResult
}

func (p ToolResultContentPart) ToolCallID() string     { return p.toolCallID }
func (p ToolResultContentPart) Name() string           { return p.name }
func (p ToolResultContentPart) Output() map[string]any { return p.output }
func (p ToolResultContentPart) Err() error             { return p.err }

// Payload returns the result in the form providers expect on the wire: the
// error's message when the tool failed, otherwise the JSON encoding of the
// output map. Providers that accept structured output prefer Output directly.
func (p ToolResultContentPart) Payload() string {
	if p.err != nil {
		// err.Error() rather than json.Marshal: most error types carry no
		// exported fields, so marshaling them yields "{}" and loses the message.
		return p.err.Error()
	}
	return string(utils.Must(json.Marshal(p.output)))
}

// BaseMessageContent is the ordered list of parts making up a message.
type BaseMessageContent struct {
	parts []MessageContentPart
}

func NewMessageContent(parts ...MessageContentPart) BaseMessageContent {
	return BaseMessageContent{parts: parts}
}

// Text returns the concatenation of the message's text parts, or "" when it has
// none. Tool calls and tool results are excluded: callers that need them should
// use Parts.
func (mc BaseMessageContent) Text() string {
	var builder strings.Builder
	for _, part := range mc.parts {
		if textPart, ok := part.(TextContentPart); ok {
			builder.WriteString(textPart.Text())
		}
	}
	return builder.String()
}

// Parts returns the message's content parts in order.
func (mc BaseMessageContent) Parts() []MessageContentPart {
	return mc.parts
}

// ToolCalls returns only the tool call parts, which providers need in order to
// replay the model's own requests back to it.
func (mc BaseMessageContent) ToolCalls() []ToolCallContentPart {
	var calls []ToolCallContentPart
	for _, part := range mc.parts {
		if call, ok := part.(ToolCallContentPart); ok {
			calls = append(calls, call)
		}
	}
	return calls
}

// ToolResults returns only the tool result parts.
func (mc BaseMessageContent) ToolResults() []ToolResultContentPart {
	var results []ToolResultContentPart
	for _, part := range mc.parts {
		if result, ok := part.(ToolResultContentPart); ok {
			results = append(results, result)
		}
	}
	return results
}

// IsEmpty reports whether the message carries no parts at all.
func (mc BaseMessageContent) IsEmpty() bool {
	return len(mc.parts) == 0
}
