package domain

type STATE string

const (
	PENDING  STATE = "pending"
	RUNNING  STATE = "running"
	ACTIVE   STATE = "active"
	COMPLETE STATE = "complete"
	INACTIVE STATE = "inactive"
	ERROR    STATE = "error"
)
