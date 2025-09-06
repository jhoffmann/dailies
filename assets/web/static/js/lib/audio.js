// Audio notification function
export function beep(frequency = 800, duration = 200) {
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
