import { Injectable } from '@angular/core';
import { DOCUMENT } from '@angular/common';
import { Inject } from '@angular/core';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class ThemeService {
  constructor(@Inject(DOCUMENT) private document: Document) {
    this.loadColorTheme();
  }

  private loadColorTheme(): void {
    const themeLink = this.document.createElement('link');
    themeLink.rel = 'stylesheet';
    themeLink.href = `https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.${environment.picoTheme}.min.css`;

    // Remove existing Pico theme if present
    const existingTheme = this.document.querySelector('link[href*="pico"]');
    if (existingTheme) {
      existingTheme.remove();
    }

    this.document.head.appendChild(themeLink);

    // Apply custom CSS variables after theme loads
    themeLink.onload = () => {
      this.applyCustomVariables();
    };
  }

  private applyCustomVariables(): void {
    const style = this.document.createElement('style');
    style.textContent = `
      :root {
        --pico-font-size: 100%;
      }
    `;
    this.document.head.appendChild(style);
  }
}
