package models

type MessageRole string

const (
	MessageRoleHuman     MessageRole = "human"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

type Message interface {
	GetType() MessageRole
	GetContent() MessageContentImpl
}
type MessageContent interface {
	GetContent() MessageContentImpl
}

type MessageImpl struct {
	messageRole MessageRole
	content     MessageContentImpl
}

type MessageContentImpl struct {
	text string
}

func (mc MessageContentImpl) GetText() string {
	return mc.text
}

func (m MessageImpl) GetType() MessageRole {
	return m.messageRole
}

func (m MessageImpl) GetContent() MessageContentImpl {
	return m.content
}

func (m MessageImpl) GetContentText() string {
	return m.content.GetText()
}

func NewStringMessage(messageRole MessageRole, content string) Message {
	return &MessageImpl{
		messageRole: messageRole,
		content:     MessageContentImpl{text: content},
	}
}

func NewHumanStringMessage(content string) Message {
	return NewStringMessage(MessageRoleHuman, content)
}

func NewAssistantStringMessage(content string) Message {
	return NewStringMessage(MessageRoleAssistant, content)
}

func NewSystemStringMessage(content string) Message {
	return NewStringMessage(MessageRoleSystem, content)
}
