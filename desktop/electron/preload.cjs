const { contextBridge, ipcRenderer } = require('electron');
const command = (action, payload) => ipcRenderer.invoke('lettuce:command', action, payload);
contextBridge.exposeInMainWorld('lettuce', {
  snapshot: () => ipcRenderer.invoke('lettuce:snapshot'), login: () => ipcRenderer.invoke('lettuce:login'),
  logout: () => command('logout'), retry: jobId => command('retry', { jobId }), clearCompleted: () => command('clear_completed'),
  updateSettings: settings => command('update_settings', { port: String(settings.port), settings }), window: action => ipcRenderer.invoke('lettuce:window', action),
  openExternal: url => ipcRenderer.invoke('lettuce:open-external', url),
  onEvent: callback => { const listener = (_event, value) => callback(value); ipcRenderer.on('lettuce:event', listener); return () => ipcRenderer.removeListener('lettuce:event', listener); },
});
