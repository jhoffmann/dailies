import { Component, OnInit } from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { ToastNotificationComponent } from './components/toast-notification/toast-notification.component';
import { FooterComponent } from './components/footer/footer.component';
import { WebSocketService } from './services/websocket.service';
import { ThemeService } from './services/theme.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    RouterOutlet,
    RouterLink,
    RouterLinkActive,
    ToastNotificationComponent,
    FooterComponent,
  ],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App implements OnInit {
  title = 'Daily Tracker';
  isDarkMode = false;

  constructor(
    private wsService: WebSocketService,
    private themeService: ThemeService,
  ) {}

  ngOnInit(): void {
    // WebSocket service starts automatically via constructor
    console.log('App initialized with WebSocket support');

    // Initialize theme from localStorage
    this.initializeTheme();
  }

  private initializeTheme(): void {
    const savedTheme = localStorage.getItem('theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

    this.isDarkMode = savedTheme ? savedTheme === 'dark' : prefersDark;
    this.applyTheme();
  }

  private applyTheme(): void {
    const htmlElement = document.documentElement;
    if (this.isDarkMode) {
      htmlElement.setAttribute('data-theme', 'dark');
    } else {
      htmlElement.removeAttribute('data-theme');
    }
  }

  onThemeToggle(): void {
    this.isDarkMode = !this.isDarkMode;
    this.applyTheme();
    localStorage.setItem('theme', this.isDarkMode ? 'dark' : 'light');
  }
}
