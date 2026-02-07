package things

import (
	"fmt"
	"os/exec"
	"regexp"
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

var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// IsSpecificDate returns true if the value is a YYYY-MM-DD date string
func IsSpecificDate(when string) bool {
	return datePattern.MatchString(when)
}

// TaskUpdateParams contains parameters for updating a task via AppleScript
type TaskUpdateParams struct {
	UUID      string
	Name      string // title
	Notes     string
	DueDate   string // YYYY-MM-DD format
	When      string // "today", "tomorrow", "evening", "anytime", "someday", or YYYY-MM-DD
	TagNames  string // comma-separated
	AuthToken string // required for specific date scheduling via URL scheme
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
			if !datePattern.MatchString(params.When) {
				return fmt.Errorf("invalid when value '%s': use 'today', 'tomorrow', 'anytime', 'someday', or YYYY-MM-DD", params.When)
			}
			// Specific dates require the URL scheme (AppleScript activation date is read-only)
			if params.AuthToken == "" {
				return fmt.Errorf("auth token required for specific date scheduling")
			}
			return updateViaURLScheme(params)
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

// updateViaURLScheme updates a task using the things:///update URL scheme.
// Used for specific date scheduling since AppleScript's activation date is read-only.
func updateViaURLScheme(params TaskUpdateParams) error {
	updateParams := UpdateParams{
		ID:        params.UUID,
		AuthToken: params.AuthToken,
		Title:     params.Name,
		Notes:     params.Notes,
		When:      params.When,
		Deadline:  params.DueDate,
		Tags:      params.TagNames,
	}
	url := BuildUpdateURL(updateParams)
	return OpenURL(url)
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

// CancelProject marks a project as canceled
func CancelProject(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	set status of project id "%s" to canceled
end tell`, uuid)
	return runAppleScript(script)
}

// CreateArea creates a new area and returns its UUID
func CreateArea(name string) (string, error) {
	script := fmt.Sprintf(`tell application "Things3"
	set newArea to make new area with properties {name:%q}
	return id of newArea
end tell`, name)
	return runAppleScriptWithOutput(script)
}

// UpdateArea updates an area's name
func UpdateArea(uuid, name string) error {
	script := fmt.Sprintf(`tell application "Things3"
	set name of area id "%s" to %q
end tell`, uuid, name)
	return runAppleScript(script)
}

// DeleteArea deletes an area by UUID
func DeleteArea(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	delete area id "%s"
end tell`, uuid)
	return runAppleScript(script)
}

// CreateTag creates a new tag and returns its UUID
func CreateTag(name string, parentUUID string) (string, error) {
	var script string
	if parentUUID != "" {
		script = fmt.Sprintf(`tell application "Things3"
	set parentTag to tag id "%s"
	set newTag to make new tag with properties {name:%q, parent tag:parentTag}
	return id of newTag
end tell`, parentUUID, name)
	} else {
		script = fmt.Sprintf(`tell application "Things3"
	set newTag to make new tag with properties {name:%q}
	return id of newTag
end tell`, name)
	}
	return runAppleScriptWithOutput(script)
}

// UpdateTag updates a tag's name
func UpdateTag(uuid, name string) error {
	script := fmt.Sprintf(`tell application "Things3"
	set name of tag id "%s" to %q
end tell`, uuid, name)
	return runAppleScript(script)
}

// DeleteTag deletes a tag by UUID
func DeleteTag(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	delete tag id "%s"
end tell`, uuid)
	return runAppleScript(script)
}

// DeleteHeading deletes a heading by UUID
// Tasks under the heading move to project root
func DeleteHeading(uuid string) error {
	script := fmt.Sprintf(`tell application "Things3"
	delete to do id "%s"
end tell`, uuid)
	return runAppleScript(script)
}

// RenameHeading renames a heading by UUID
func RenameHeading(uuid, newTitle string) error {
	script := fmt.Sprintf(`tell application "Things3"
	set name of to do id "%s" to %q
end tell`, uuid, newTitle)
	return runAppleScript(script)
}

// runAppleScriptWithOutput executes AppleScript and returns the output
func runAppleScriptWithOutput(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("applescript error: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return strings.TrimSpace(string(output)), nil
}

// MoveTaskToArea moves a task to an area by UUID
func MoveTaskToArea(taskUUID, areaUUID string) error {
	script := fmt.Sprintf(`tell application "Things3"
	set aToDo to to do id "%s"
	set area of aToDo to area id "%s"
end tell`, taskUUID, areaUUID)
	return runAppleScript(script)
}

// MoveTaskToProject moves a task to a project by UUID
func MoveTaskToProject(taskUUID, projectUUID string) error {
	script := fmt.Sprintf(`tell application "Things3"
	set aToDo to to do id "%s"
	set project of aToDo to project id "%s"
end tell`, taskUUID, projectUUID)
	return runAppleScript(script)
}

// DeleteAllOpenTasks deletes all open tasks (moves to trash) and returns the count
func DeleteAllOpenTasks() (int, error) {
	script := `tell application "Things3"
	set openTodos to every to do whose status is open
	set countDeleted to count of openTodos
	repeat with t in openTodos
		delete t
	end repeat
	return countDeleted
end tell`
	result, err := runAppleScriptWithOutput(script)
	if err != nil {
		return 0, err
	}
	var count int
	_, err = fmt.Sscanf(result, "%d", &count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}
