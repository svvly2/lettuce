export type JobStatus = 'queued' | 'uploading' | 'retrying' | 'complete' | 'failed';
export type LogLevel = 'info' | 'success' | 'warning' | 'error';

export interface RobloxUser { id: string; username: string; displayName: string; avatarUrl?: string }
export interface UploadJob { id: string; name: string; sourceAssetId: string; resultAssetId?: string; progress: number; attempt: number; status: JobStatus }
export interface LogEntry { id: string; at: string; level: LogLevel; message: string }
export interface Settings { port: number; concurrency: number; maxRetries: number; launchAtLogin: boolean; notifications: boolean; autoUpdate: boolean }
export interface AppSnapshot { user: RobloxUser | null; pluginConnected: boolean; serverRunning: boolean; queue: UploadJob[]; logs: LogEntry[]; settings: Settings; version: string }

export type DesktopEvent = { type: 'snapshot'; payload: AppSnapshot } | { type: 'toast'; payload: { level: LogLevel; message: string } };

export interface DesktopBridge {
  snapshot(): Promise<AppSnapshot>;
  login(): Promise<void>;
  logout(): Promise<void>;
  retry(jobId: string): Promise<void>;
  clearCompleted(): Promise<void>;
  updateSettings(settings: Settings): Promise<void>;
  window(action: 'minimize' | 'close'): Promise<void>;
  openExternal(url: string): Promise<void>;
  onEvent(callback: (event: DesktopEvent) => void): () => void;
}
