package domain

type STATE string

const (
	STATE_PENDING   STATE = "pending"
	STATE_PROCESSED STATE = "processed"
	STATE_RUNNING   STATE = "running"
	STATE_ACTIVE    STATE = "active"
	STATE_COMPLETE  STATE = "complete"
	STATE_INACTIVE  STATE = "inactive"
	STATE_ARCHIVED  STATE = "archived"
	STATE_ERROR     STATE = "error"
)
