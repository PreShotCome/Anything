const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('api', {
  getCaptureSources: () => ipcRenderer.invoke('get-capture-sources'),
  analyzeScreen: (imageBase64) => ipcRenderer.invoke('analyze-screen', imageBase64),
  summarizeSession: (summaries) => ipcRenderer.invoke('summarize-session', summaries),
  saveRecording: (bytes) => ipcRenderer.invoke('save-recording', bytes),
  notify: (title, body) => ipcRenderer.invoke('notify', { title, body }),
  onHotkeyToggle: (callback) => ipcRenderer.on('hotkey-toggle', () => callback()),
  onHotkeyFailed: (callback) => ipcRenderer.on('hotkey-failed', () => callback()),
});
