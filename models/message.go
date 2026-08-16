package models

import (
	"encoding/json"
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

func NewStringMessage(messageRole MessageRole, content string) *BaseMessage {
	return &BaseMessage{
		messageRole: messageRole,
		content:     BaseMessageContent{text: content},
	}
}

func NewHumanStringMessage(content string) *BaseMessage {
	return NewStringMessage(MessageRoleHuman, content)
}

func NewAssistantStringMessage(content string) *BaseMessage {
	return NewStringMessage(MessageRoleAssistant, content)
}

func NewMessageFromToolResult(output ToolExecuteOutput) *BaseMessage {
	if output.Error != nil {
		stringError, _ := json.Marshal(output.Error)
		return NewAssistantStringMessage("Tool execution error: " + string(stringError))
	} else {
		stringResult, _ := json.Marshal(output.Result)
		return NewAssistantStringMessage("Tool execution result: " + string(stringResult))
	}
}

func NewSystemStringMessage(content string) *BaseMessage {
	return NewStringMessage(MessageRoleSystem, content)
}
