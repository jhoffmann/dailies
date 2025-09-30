import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root',
})
export class AudioService {
  private audioContext: AudioContext | null = null;
  private isEnabled = true;

  constructor() {
    this.initAudioContext();
  }

  private initAudioContext(): void {
    try {
      this.audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
    } catch (error) {
      console.warn('Web Audio API not supported:', error);
    }
  }

  public enable(): void {
    this.isEnabled = true;
  }

  public disable(): void {
    this.isEnabled = false;
  }

  public isAudioEnabled(): boolean {
    return this.isEnabled && this.audioContext !== null;
  }

  public async playNotification(): Promise<void> {
    if (!this.isEnabled || !this.audioContext) {
      return;
    }

    try {
      // Resume audio context if it's suspended (common after user interaction)
      if (this.audioContext.state === 'suspended') {
        await this.audioContext.resume();
      }

      // Create a gentle notification sound using Web Audio API
      this.playTone(800, 0.1, 'sine'); // Higher pitch

      setTimeout(() => {
        this.playTone(600, 0.1, 'sine'); // Lower pitch for a pleasant two-tone chime
      }, 100);
    } catch (error) {
      console.warn('Error playing notification sound:', error);
    }
  }

  private playTone(frequency: number, duration: number, type: OscillatorType = 'sine'): void {
    if (!this.audioContext) return;

    const oscillator = this.audioContext.createOscillator();
    const gainNode = this.audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(this.audioContext.destination);

    oscillator.frequency.setValueAtTime(frequency, this.audioContext.currentTime);
    oscillator.type = type;

    // Create a gentle envelope to avoid harsh starts/stops
    gainNode.gain.setValueAtTime(0, this.audioContext.currentTime);
    gainNode.gain.linearRampToValueAtTime(0.1, this.audioContext.currentTime + 0.01); // Quick fade in
    gainNode.gain.exponentialRampToValueAtTime(0.01, this.audioContext.currentTime + duration); // Gentle fade out

    oscillator.start(this.audioContext.currentTime);
    oscillator.stop(this.audioContext.currentTime + duration);
  }

  public async testNotificationSound(): Promise<void> {
    if (!this.audioContext) {
      console.warn('Audio context not available');
      return;
    }

    // User interaction is required for audio to play in modern browsers
    if (this.audioContext.state === 'suspended') {
      await this.audioContext.resume();
    }

    await this.playNotification();
  }
}
