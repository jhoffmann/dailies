// Package websocket provides global WebSocket notification management
package websocket

import "sync"

var (
	globalHub *Hub
	once      sync.Once
)

// GetHub returns the global WebSocket hub instance
func GetHub() *Hub {
	once.Do(func() {
		globalHub = NewHub()
		go globalHub.Run()
	})
	return globalHub
}

// BroadcastTaskReset broadcasts a task reset notification
func BroadcastTaskReset(taskName, frequencyName string, resetCount int) {
	hub := GetHub()
	hub.Broadcast("task-reset",
		frequencyName+" tasks have been reset",
		map[string]any{
			"task_name":      taskName,
			"frequency_name": frequencyName,
			"reset_count":    resetCount,
		})
}

// BroadcastTaskUpdated broadcasts a task status change notification
func BroadcastTaskUpdated(taskID, taskName string, completed bool) {
	hub := GetHub()
	action := "updated"
	if completed {
		action = "completed"
	} else {
		action = "uncompleted"
	}

	hub.Broadcast("task-updated",
		"Task "+action+": "+taskName,
		map[string]any{
			"task_id":   taskID,
			"task_name": taskName,
			"completed": completed,
			"action":    action,
		})
}

// BroadcastTaskListRefresh broadcasts a task list refresh request
func BroadcastTaskListRefresh() {
	hub := GetHub()
	hub.Broadcast("task-list-refresh", "Task list updated", nil)
}
