package models

type EndCondition interface {
	Condition(archive *ToolExecutionsArchive) bool
}
