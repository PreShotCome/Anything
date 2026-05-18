const {
  app,
  BrowserWindow,
  desktopCapturer,
  ipcMain,
  Notification,
  globalShortcut,
  dialog,
} = require('electron');
const path = require('path');
const fs = require('fs');
const { analyzeScreenshot, summarizeSession } = require('./analyzer');

const HOTKEY = 'CommandOrControl+Shift+R';
let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1040,
    height: 880,
    backgroundColor: '#0e1116',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });
  mainWindow.loadFile(path.join(__dirname, 'renderer', 'index.html'));
}

app.whenReady().then(() => {
  ipcMain.handle('get-capture-sources', async () => {
    const sources = await desktopCapturer.getSources({
      types: ['window', 'screen'],
      thumbnailSize: { width: 0, height: 0 },
    });
    return sources.map((s) => ({
      id: s.id,
      name: s.name,
      kind: s.id.startsWith('screen') ? 'screen' : 'window',
    }));
  });

  ipcMain.handle('analyze-screen', (_event, imageBase64) => analyzeScreenshot(imageBase64));

  ipcMain.handle('summarize-session', (_event, summaries) => summarizeSession(summaries));

  ipcMain.handle('save-recording', async (_event, bytes) => {
    const defaultPath = path.join(app.getPath('videos'), `recording-${Date.now()}.webm`);
    const { canceled, filePath } = await dialog.showSaveDialog(mainWindow, {
      title: 'Save recording',
      defaultPath,
      filters: [{ name: 'WebM video', extensions: ['webm'] }],
    });
    if (canceled || !filePath) return { ok: false, canceled: true };
    fs.writeFileSync(filePath, Buffer.from(bytes));
    return { ok: true, path: filePath };
  });

  ipcMain.handle('notify', (_event, { title, body }) => {
    new Notification({ title, body }).show();
  });

  const registered = globalShortcut.register(HOTKEY, () => {
    if (mainWindow) mainWindow.webContents.send('hotkey-toggle');
  });
  if (!registered && mainWindow) {
    mainWindow.webContents.send('hotkey-failed');
  }

  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on('will-quit', () => globalShortcut.unregisterAll());

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
