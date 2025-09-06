import { beep } from "./audio.js";

// WebSocket notification functionality
window.handleWebSocketMessage = function (event) {
  const message = JSON.parse(event.detail.message);

  // Show notification
  showNotification(message);

  // Handle task list refresh
  if (message.type === "task-list-refresh") {
    const taskListElement = document.getElementById("refresh-trigger");

    if (taskListElement) {
      if (window.htmx && window.htmx.trigger) {
        window.htmx.trigger(taskListElement, "click");
      } else {
        // Fallback: dispatch a custom event
        taskListElement.dispatchEvent(new CustomEvent("click"));
      }
    }
  }
};

// Show notification with slide animation
function showNotification(message) {
  const container = document.getElementById("notification-container");
  if (!container) return;

  // Play beep sound for task-reset notifications
  if (message.type === "task-reset") {
    beep(600, 350);
  }

  const notification = document.createElement("div");
  notification.className = `notification ${message.type}`;
  notification.textContent = message.message;

  // Add click to dismiss
  notification.addEventListener("click", function () {
    hideNotification(notification);
  });

  container.appendChild(notification);

  // Trigger show animation
  requestAnimationFrame(() => {
    notification.classList.add("show");
  });

  // Auto-hide after 15 seconds
  setTimeout(() => {
    hideNotification(notification);
  }, 15000);
}

// Hide notification with slide animation
function hideNotification(notification) {
  notification.classList.remove("show");
  setTimeout(() => {
    if (notification.parentNode) {
      notification.parentNode.removeChild(notification);
    }
  }, 300);
}
