package domain

type STATE string

const (
	PENDING  STATE = "pending"
	RUNNING  STATE = "running"
	INACTIVE STATE = "inactive"
	ERROR    STATE = "error"
)
