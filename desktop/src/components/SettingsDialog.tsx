import { useState } from 'react';
import type { Settings } from '../domain/types';
import { Icon } from './Icon';

export function SettingsDialog({ value, close, save }: { value: Settings; close(): void; save(value: Settings): void }) {
  const [settings, setSettings] = useState(value);
  return <div className="scrim" onMouseDown={close}><section className="dialog" role="dialog" aria-modal="true" aria-label="Settings" onMouseDown={(event) => event.stopPropagation()}>
    <div className="dialog-head"><h2>Settings</h2><button className="icon-button" aria-label="Close" onClick={close}><Icon name="close" /></button></div>
    <div className="setting"><div><strong>Studio port</strong><span>Only accepts requests from localhost.</span></div><input type="number" value={settings.port} onChange={(event) => setSettings({ ...settings, port: Number(event.target.value) })} /></div>
    <div className="setting"><div><strong>Concurrent uploads</strong><span>Balance speed against Roblox rate limits.</span></div><input type="number" min="1" max="8" value={settings.concurrency} onChange={(event) => setSettings({ ...settings, concurrency: Number(event.target.value) })} /></div>
    <div className="setting"><div><strong>Retry attempts</strong><span>Failed uploads retry automatically.</span></div><input type="number" min="0" max="10" value={settings.maxRetries} onChange={(event) => setSettings({ ...settings, maxRetries: Number(event.target.value) })} /></div>
    {(['notifications', 'autoUpdate', 'launchAtLogin'] as const).map((key) => <label className="setting switch-row" key={key}><div><strong>{{ notifications: 'Notifications', autoUpdate: 'Automatic updates', launchAtLogin: 'Launch at login' }[key]}</strong></div><input type="checkbox" checked={settings[key]} onChange={(event) => setSettings({ ...settings, [key]: event.target.checked })} /><i /></label>)}
    <footer><button className="secondary" onClick={close}>Cancel</button><button className="primary" onClick={() => save(settings)}>Save</button></footer>
  </section></div>;
}
