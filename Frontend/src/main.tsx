import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from './App';
import './index.css';
import { initGA4 } from './services/analytics';
import { wsClient } from './services/websocket';
import { getAuthToken } from './services/api';
import { authService } from './services/authService';
import { useAuthStore } from './store/authStore';

// Initialize Analytics & Real-Time connection
initGA4();
wsClient.connect();

// Restore auth session on page load.
// authStore is persisted, so user & token are rehydrated from localStorage.
// But we still verify the token is valid against the backend in the background.
(async () => {
  const token = getAuthToken();
  if (token && token !== 'mock-jwt-token') {
    const freshUser = await authService.getCurrentUser();
    if (freshUser) {
      useAuthStore.getState().signIn(freshUser, token);
    } else {
      // Token expired or invalid — sign out cleanly
      useAuthStore.getState().signOut();
    }
  }
})();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
