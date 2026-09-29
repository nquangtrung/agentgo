package models

type ObjectPart interface {
	Part
	Object() any
	Raw() string
}

type BaseObjectPart struct {
	BasePart
	object any
	raw    string
}

func (p BaseObjectPart) Object() any { return p.object }
func (p BaseObjectPart) Raw() string { return p.raw }

func NewObjectPart(context LanguageModelContext, object any, raw string) *BaseObjectPart {
	return &BaseObjectPart{
		BasePart: BasePart{partType: PartTypeObject, context: context},
		object:   object,
		raw:      raw,
	}
}

type ObjectErrorPart interface {
	Part
	Error() error
	Raw() string
}

type BaseObjectErrorPart struct {
	BasePart
	err error
	raw string
}

func (p BaseObjectErrorPart) Error() error { return p.err }
func (p BaseObjectErrorPart) Raw() string  { return p.raw }

func NewObjectErrorPart(context LanguageModelContext, err error, raw string) *BaseObjectErrorPart {
	return &BaseObjectErrorPart{
		BasePart: BasePart{partType: PartTypeObjectError, context: context},
		err:      err,
		raw:      raw,
	}
}
