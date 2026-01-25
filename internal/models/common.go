package models

// TaskStatus represents the status of a task or project
type TaskStatus int

const (
	StatusIncomplete TaskStatus = 0
	StatusCanceled   TaskStatus = 2
	StatusCompleted  TaskStatus = 3
)

func (s TaskStatus) String() string {
	switch s {
	case StatusIncomplete:
		return "incomplete"
	case StatusCanceled:
		return "canceled"
	case StatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

func (s TaskStatus) Icon() string {
	switch s {
	case StatusIncomplete:
		return "○"
	case StatusCompleted:
		return "✓"
	case StatusCanceled:
		return "✗"
	default:
		return "?"
	}
}

// TaskType represents the type of item in TMTask table
type TaskType int

const (
	TypeTask    TaskType = 0
	TypeProject TaskType = 1
	TypeHeading TaskType = 2
)

func (t TaskType) String() string {
	switch t {
	case TypeTask:
		return "Task"
	case TypeProject:
		return "Project"
	case TypeHeading:
		return "Heading"
	default:
		return "Unknown"
	}
}
