package models

type PartEmitter struct {
	channel chan Part
}

func (e *PartEmitter) Emit(part Part) {
	if e.channel == nil {
		return
	}
	e.channel <- part
}

func NewPartEmitter(channel chan Part) *PartEmitter {
	return &PartEmitter{
		channel: channel,
	}
}

func NewEmptyPartEmitter() *PartEmitter {
	return NewPartEmitter(nil)

}
