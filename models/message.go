package models

import (
	"encoding/json"
	"fmt"

	"github.com/nquangtrung/agentgo/utils"
)

type MessageRole string

const (
	MessageRoleHuman     MessageRole = "human"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

type Message interface {
	Type() MessageRole
	Content() BaseMessageContent
}
type MessageContent interface {
	Content() BaseMessageContent
}

type BaseMessage struct {
	messageRole MessageRole
	content     BaseMessageContent
}

type BaseMessageContent struct {
	text string
}

func (mc BaseMessageContent) Text() string {
	return mc.text
}

func (m BaseMessage) Type() MessageRole {
	return m.messageRole
}

func (m BaseMessage) Content() BaseMessageContent {
	return m.content
}

func (m BaseMessage) ContentText() string {
	return m.content.Text()
}

func NewStringMessage(messageRole MessageRole, content string) BaseMessage {
	return BaseMessage{
		messageRole: messageRole,
		content:     BaseMessageContent{text: content},
	}
}

func NewHumanStringMessage(content string) BaseMessage {
	return NewStringMessage(MessageRoleHuman, content)
}

func NewAssistantStringMessage(content string) BaseMessage {
	return NewStringMessage(MessageRoleAssistant, content)
}

func NewMessageFromToolResult(output ToolExecuteOutput) BaseMessage {
	// slog.Warn("output", "output", output)
	toolName := "text"
	if output.ToolCall != nil {
		toolName = output.ToolCall.ToolName
	}
	if output.Error != nil {
		stringError := utils.Must(json.Marshal(output.Error))
		return NewAssistantStringMessage(
			fmt.Sprintf("Tool [%s] execution error: %s",
				toolName,
				string(stringError),
			),
		)
	} else {
		stringResult := utils.Must(json.Marshal(output.Output))
		return NewAssistantStringMessage(fmt.Sprintf("Tool [%s] execution result: %s",
			toolName,
			string(stringResult),
		))
	}
}

func NewSystemStringMessage(content string) BaseMessage {
	return NewStringMessage(MessageRoleSystem, content)
}
