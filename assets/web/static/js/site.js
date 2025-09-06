import htmx from "htmx.org";
import "htmx-ext-json-enc";
import "htmx-ext-ws";
import _hyperscript from "hyperscript.org";

// Expose htmx globally
window.htmx = htmx;

_hyperscript.browserInit();

// Audio notification function
function beep(frequency = 800, duration = 200) {
  try {
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const oscillator = ctx.createOscillator();
    const gainNode = ctx.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(ctx.destination);

    oscillator.frequency.value = frequency;
    oscillator.type = "sine";

    gainNode.gain.setValueAtTime(0.3, ctx.currentTime);
    gainNode.gain.exponentialRampToValueAtTime(
      0.01,
      ctx.currentTime + duration / 1000,
    );

    oscillator.start();
    oscillator.stop(ctx.currentTime + duration / 1000);
  } catch (e) {
    console.warn("Audio context not available:", e);
  }
}

// WebSocket notification functionality
window.handleWebSocketMessage = function (event) {
  const message = JSON.parse(event.detail.message);

  // Show notification
  showNotification(message);

  // Handle task list refresh
  if (message.type === "task-list-refresh") {
    const taskListElement = document.getElementById("task-list");

    if (taskListElement) {
      if (window.htmx && window.htmx.trigger) {
        window.htmx.trigger(taskListElement, "refresh");
      } else {
        // Fallback: dispatch a custom event
        taskListElement.dispatchEvent(new CustomEvent("refresh"));
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
