// Wails uses an event-based communication between Go backend and JavaScript frontend
// The runtime is automatically injected by Wails

// Listen for lyrics updates from Go backend
// The Go backend will emit "lyrics-update" events with current and next lyrics
window.runtime.EventsOn("lyrics-update", (data) => {
  // Update the current line element
  const currentLine = document.getElementById("current-line");
  currentLine.textContent = data.current || "Waiting for music...";

  // Update the next line element (can be empty)
  const nextLine = document.getElementById("next-line");
  nextLine.textContent = data.next || "";
});

// TODO: Add manual search functionality

// TODO: Add keyboard shortcuts
document.addEventListener("keydown", (event) => {
  // ?? later
});
