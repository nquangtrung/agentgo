package models

type EndCondition interface {
	Condition(context *ToolExecutionsArchive) bool
}
