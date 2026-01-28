package things

import (
	"fmt"
	"os/exec"
	"strings"
)

// runAppleScript executes AppleScript code
func runAppleScript(script string) error {
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("applescript error: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// DeleteTask deletes (trashes) a task by UUID
func DeleteTask(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	delete to do id "%s"
end tell`, uuid)
	return runAppleScript(script)
}

// DeleteProject deletes (trashes) a project by UUID
func DeleteProject(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	delete project id "%s"
end tell`, uuid)
	return runAppleScript(script)
}

// CompleteTask marks a task as complete by UUID
func CompleteTask(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	set status of to do id "%s" to completed
end tell`, uuid)
	return runAppleScript(script)
}

// CancelTask marks a task as canceled by UUID
func CancelTask(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	set status of to do id "%s" to canceled
end tell`, uuid)
	return runAppleScript(script)
}

// MoveTaskToToday moves a task to the Today list
func MoveTaskToToday(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	move to do id "%s" to list "Today"
end tell`, uuid)
	return runAppleScript(script)
}

// TaskUpdateParams contains parameters for updating a task via AppleScript
type TaskUpdateParams struct {
	UUID     string
	Name     string // title
	Notes    string
	DueDate  string // YYYY-MM-DD format
	When     string // "today", "tomorrow", "evening", "anytime", "someday", or YYYY-MM-DD
	TagNames string // comma-separated
}

// UpdateTask updates a task's properties via AppleScript
func UpdateTask(params TaskUpdateParams) error {
	var statements []string

	if params.Name != "" {
		statements = append(statements, fmt.Sprintf(`set name of theTodo to %q`, params.Name))
	}
	if params.Notes != "" {
		statements = append(statements, fmt.Sprintf(`set notes of theTodo to %q`, params.Notes))
	}
	if params.DueDate != "" {
		statements = append(statements, fmt.Sprintf(`set due date of theTodo to date "%s"`, params.DueDate))
	}
	if params.When != "" {
		switch params.When {
		case "today", "evening":
			// Note: "evening" just moves to Today; Things 3 doesn't expose evening scheduling via AppleScript
			statements = append(statements, `move theTodo to list "Today"`)
		case "tomorrow":
			statements = append(statements, `move theTodo to list "Tomorrow"`)
		case "anytime":
			statements = append(statements, `move theTodo to list "Anytime"`)
		case "someday":
			statements = append(statements, `move theTodo to list "Someday"`)
		default:
			// Specific dates (YYYY-MM-DD) cannot be set via AppleScript
			// The activation date property is read-only
			return fmt.Errorf("cannot set specific date '%s' via AppleScript; Things 3 activation date is read-only. Use 'today', 'tomorrow', 'anytime', or 'someday' instead", params.When)
		}
	}
	if params.TagNames != "" {
		statements = append(statements, fmt.Sprintf(`set tag names of theTodo to %q`, params.TagNames))
	}

	if len(statements) == 0 {
		return fmt.Errorf("no update parameters provided")
	}

	script := fmt.Sprintf(`tell application "Things3"
	set theTodo to to do id "%s"
	%s
end tell`, params.UUID, strings.Join(statements, "\n\t"))

	return runAppleScript(script)
}

// ProjectUpdateParams contains parameters for updating a project via AppleScript
type ProjectUpdateParams struct {
	UUID     string
	Name     string // title
	Notes    string
	DueDate  string // YYYY-MM-DD format
	TagNames string // comma-separated
}

// UpdateProject updates a project's properties via AppleScript
func UpdateProject(params ProjectUpdateParams) error {
	var statements []string

	if params.Name != "" {
		statements = append(statements, fmt.Sprintf(`set name of theProject to %q`, params.Name))
	}
	if params.Notes != "" {
		statements = append(statements, fmt.Sprintf(`set notes of theProject to %q`, params.Notes))
	}
	if params.DueDate != "" {
		statements = append(statements, fmt.Sprintf(`set due date of theProject to date "%s"`, params.DueDate))
	}
	if params.TagNames != "" {
		statements = append(statements, fmt.Sprintf(`set tag names of theProject to %q`, params.TagNames))
	}

	if len(statements) == 0 {
		return fmt.Errorf("no update parameters provided")
	}

	script := fmt.Sprintf(`tell application "Things3"
	set theProject to project id "%s"
	%s
end tell`, params.UUID, strings.Join(statements, "\n\t"))

	return runAppleScript(script)
}

// CompleteProject marks a project as complete
func CompleteProject(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	set status of project id "%s" to completed
end tell`, uuid)
	return runAppleScript(script)
}
