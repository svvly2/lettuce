const { app, BrowserWindow, ipcMain, shell } = require('electron');
const { spawn } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const { startUpdater } = require('./updater.cjs');

const API = 'http://127.0.0.1:39000/api';
let windowRef; let daemon; let timer; let stopUpdater = () => {}; let snapshot = null;
const defaultSettings = { port: 38073, concurrency: 3, maxRetries: 3, launchAtLogin: false, notifications: true, autoUpdate: true };

function backendPath() {
  const roots = [process.resourcesPath, app.getAppPath(), path.resolve(app.getAppPath(), '..')];
  return roots.map(root => path.join(root, process.platform === 'win32' ? 'lettuce-daemon.exe' : 'lettuce-daemon')).find(fs.existsSync);
}
function startDaemon() {
  const executable = backendPath(); if (!executable) return;
  const runtimeDir = app.isPackaged ? process.resourcesPath : path.resolve(app.getAppPath(), '..');
  daemon = spawn(executable, [], { cwd: runtimeDir, windowsHide: true, stdio: 'ignore' });
  daemon.once('exit', () => { daemon = undefined; });
}
function installStudioPlugin() {
  if (!app.isPackaged || process.platform !== 'win32') return;
  const source = path.join(process.resourcesPath, 'Lettuce-Plugin.rbxm');
  const targetDir = path.join(process.env.LOCALAPPDATA || app.getPath('appData'), 'Roblox', 'Plugins');
  const target = path.join(targetDir, 'Lettuce-Plugin.rbxm');
  try {
    if (!fs.existsSync(source)) return;
    fs.mkdirSync(targetDir, { recursive: true });
    const sourceData = fs.readFileSync(source);
    const targetData = fs.existsSync(target) ? fs.readFileSync(target) : null;
    if (!targetData || !sourceData.equals(targetData)) fs.copyFileSync(source, target);
  } catch { /* Studio can keep using its current copy until the next launch. */ }
}
function normalize(raw = {}) {
  const settings = { ...defaultSettings, port: Number(raw.serverPort || defaultSettings.port) };
  const logs = (raw.logs || []).map(entry => ({ id: String(entry.id), at: entry.time, level: entry.level === 'warn' ? 'warning' : entry.level, message: entry.message }));
  return { user: raw.isLoggedIn ? { id: String(raw.userId || ''), username: raw.username || '', displayName: raw.displayName || raw.username || '', avatarUrl: raw.avatarUrl || undefined } : null, pluginConnected: Boolean(raw.pluginConnected || raw.busy), serverRunning: raw.serverRunning !== false, queue: raw.queue || [], logs, settings, version: app.getVersion() };
}
async function request(route, init) { const response = await fetch(`${API}${route}`, init); if (!response.ok) throw new Error(`Backend request failed (${response.status})`); return response.headers.get('content-type')?.includes('json') ? response.json() : undefined; }
async function poll() { try { snapshot = normalize(await request('/state')); windowRef?.webContents.send('lettuce:event', { type: 'snapshot', payload: snapshot }); } catch { if (!snapshot) snapshot = normalize({ serverRunning: false }); } }
function createWindow() {
  windowRef = new BrowserWindow({ width: 1120, height: 760, minWidth: 760, minHeight: 600, frame: false, titleBarStyle: 'hidden', backgroundColor: '#050505', show: false, webPreferences: { preload: path.join(__dirname, 'preload.cjs'), contextIsolation: true, nodeIntegration: false, sandbox: true } });
  windowRef.once('ready-to-show', () => windowRef.show());
  if (process.env.VITE_DEV_SERVER_URL) void windowRef.loadURL(process.env.VITE_DEV_SERVER_URL); else void windowRef.loadFile(path.join(app.getAppPath(), 'dist', 'index.html'));
}
ipcMain.handle('lettuce:snapshot', async () => { if (!snapshot) await poll(); return snapshot; });
ipcMain.handle('lettuce:login', async () => { const { url } = await request('/oauth/start'); if (url) await shell.openExternal(url); });
ipcMain.handle('lettuce:command', (_event, action, payload = {}) => request('/command', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ action, ...payload }) }));
ipcMain.handle('lettuce:window', (_event, action) => { if (action === 'minimize') windowRef?.minimize(); if (action === 'close') windowRef?.close(); });
ipcMain.handle('lettuce:open-external', (_event, url) => { if (url === 'https://discord.gg/4ycV7TUX6G') return shell.openExternal(url); });

app.whenReady().then(() => { installStudioPlugin(); startDaemon(); createWindow(); void poll(); timer = setInterval(poll, 1000); stopUpdater = startUpdater({ packaged: app.isPackaged }); });
app.on('window-all-closed', () => { clearInterval(timer); stopUpdater(); daemon?.kill(); if (process.platform !== 'darwin') app.quit(); });
