/// <reference types="vite/client" />

import type { DesktopBridge } from './domain/types';
declare global { interface Window { lettuce?: DesktopBridge } }
export {};

type DaemonMessage = {
  type: 'log' | 'state' | 'error';
  level?: 'info' | 'success' | 'warn' | 'error';
  message?: string;
  time?: string;
  isLoggedIn?: boolean;
  username?: string;
  displayName?: string;
  serverRunning?: boolean;
  serverPort?: string;
  busy?: boolean;
};

declare global {
  interface Window {
    lettuce: {
      send: (cmd: Record<string, unknown>) => Promise<void>;
      onMessage: (cb: (line: string) => void) => () => void;
    };
  }
}

export {};
