let lyricsLines = [];
let currentLineIndex = -1;
let startTime = 0;
let pausedAt = null; // performance.now() timestamp when paused
let animationFrameId = null;
let detectionEnabled = true;

// ── Now Playing ──────────────────────────────────────────────────────────────

window.runtime.EventsOn("now-playing", (info) => {
  const label = info.artist ? `${info.artist} — ${info.title}` : info.title;
  document.getElementById("song-info").textContent = label;
  document.getElementById("current-line").textContent = "Loading lyrics…";
  document.getElementById("next-line").textContent = "";
  // Clear old sync so stale lines don't flash
  lyricsLines = [];
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
    animationFrameId = null;
  }
});

// ── Lyrics ───────────────────────────────────────────────────────────────────

// Backend emits { lines: [{Timestamp, Text}], offsetMs: number }
// Timestamp is in nanoseconds (Go time.Duration).
// offsetMs is how far into the song we are when lyrics were fetched.
window.runtime.EventsOn("lyrics-found", (data) => {
  lyricsLines = data.lines || [];
  const offsetMs = data.offsetMs || 0;

  // Shift startTime so elapsed time in updateLyrics() matches song position.
  startTime = performance.now() - offsetMs - 400;
  currentLineIndex = -1;

  document.getElementById("current-line").textContent = "";
  document.getElementById("next-line").textContent = "";

  if (animationFrameId) cancelAnimationFrame(animationFrameId);
  updateLyrics();
});

// Plain (unsynced) lyrics — display first line as static text
window.runtime.EventsOn("lyrics-plain", (text) => {
  lyricsLines = [];
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
    animationFrameId = null;
  }
  const firstLine = (text || "").split(/\r?\n/).find((l) => l.trim()) || "";
  document.getElementById("current-line").textContent = firstLine;
  document.getElementById("next-line").textContent = "(no sync available)";
});

window.runtime.EventsOn("playback-paused", () => {
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
    animationFrameId = null;
  }
  pausedAt = performance.now();
});

window.runtime.EventsOn("playback-resumed", (posMs) => {
  if (lyricsLines.length === 0) return;
  // Recalibrate from SMTC position if available, otherwise shift by pause duration.
  if (posMs > 0) {
    startTime = performance.now() - posMs - 400; // Nudge back 400ms to better align with lyric timing
  } else if (pausedAt !== null) {
    startTime += performance.now() - pausedAt - 400; // Nudge back 400ms to better align with lyric timing
  }
  pausedAt = null;
  if (!animationFrameId) updateLyrics();
});

window.runtime.EventsOn("playback-seeked", (posMs) => {
  if (lyricsLines.length === 0 || posMs <= 0) return;
  startTime = performance.now() - posMs - 400;
});

window.runtime.EventsOn("lyrics-error", (msg) => {
  lyricsLines = [];
  document.getElementById("current-line").textContent = msg;
  document.getElementById("next-line").textContent = "";
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
    animationFrameId = null;
  }
});

// ── Sync loop ────────────────────────────────────────────────────────────────

function updateLyrics() {
  if (lyricsLines.length === 0) return;

  const elapsedMs = performance.now() - startTime;

  let activeIndex = -1;
  for (let i = 0; i < lyricsLines.length; i++) {
    const lineMs = lyricsLines[i].Timestamp / 1e6; // ns → ms
    if (elapsedMs >= lineMs) {
      activeIndex = i;
    } else {
      break;
    }
  }

  if (activeIndex !== currentLineIndex) {
    currentLineIndex = activeIndex;
    if (activeIndex >= 0) {
      document.getElementById("current-line").textContent =
        lyricsLines[activeIndex].Text;
      const next = lyricsLines[activeIndex + 1];
      document.getElementById("next-line").textContent = next ? next.Text : "";
    }
  }

  animationFrameId = requestAnimationFrame(updateLyrics);
}

// ── Controls ─────────────────────────────────────────────────────────────────

window.toggleDetection = function () {
  detectionEnabled = !detectionEnabled;
  document.getElementById("toggle-btn").textContent = detectionEnabled
    ? "⏸"
    : "▶";

  if (window.go?.main?.App) {
    window.go.main.App.ToggleDetection(detectionEnabled);
  }

  if (!detectionEnabled) {
    lyricsLines = [];
    if (animationFrameId) {
      cancelAnimationFrame(animationFrameId);
      animationFrameId = null;
    }
    document.getElementById("song-info").textContent = "Detection paused";
    document.getElementById("current-line").textContent = "";
    document.getElementById("next-line").textContent = "";
  }
};

window.testSong = function (title, artist) {
  if (window.go?.main?.App) {
    window.go.main.App.TriggerTest(title, artist);
  } else {
    console.error("Wails runtime not ready");
  }
};

// Spacebar toggles detection; no other default shortcuts needed.
document.addEventListener("keydown", (e) => {
  if (e.key === " " && e.target === document.body) {
    e.preventDefault();
    toggleDetection();
  }
});
