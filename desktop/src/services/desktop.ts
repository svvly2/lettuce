import type { AppSnapshot, DesktopBridge } from '../domain/types';

const defaults: AppSnapshot = {
  user: null, pluginConnected: false, serverRunning: true, queue: [], logs: [], version: '0.1.0',
  settings: { port: 38073, concurrency: 3, maxRetries: 3, launchAtLogin: false, notifications: true, autoUpdate: true },
};

const fallback: DesktopBridge = {
  async snapshot() { return defaults; }, async login() {}, async logout() {}, async retry() {}, async clearCompleted() {},
  async updateSettings() {}, async window() {}, async openExternal() {}, onEvent() { return () => undefined; },
};

export const desktop = (): DesktopBridge => window.lettuce ?? fallback;
