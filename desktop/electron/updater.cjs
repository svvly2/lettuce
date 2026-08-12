const { autoUpdater } = require('electron-updater');

const CHECK_DELAY_MS = 8_000;
const CHECK_INTERVAL_MS = 4 * 60 * 60 * 1_000;

function startUpdater({ packaged, enabled = true } = {}) {
  if (!packaged || !enabled) return () => {};

  autoUpdater.autoDownload = true;
  autoUpdater.autoInstallOnAppQuit = true;
  autoUpdater.logger = null;

  const check = () => autoUpdater.checkForUpdatesAndNotify().catch(() => {});
  const firstCheck = setTimeout(check, CHECK_DELAY_MS);
  const recurringCheck = setInterval(check, CHECK_INTERVAL_MS);

  return () => {
    clearTimeout(firstCheck);
    clearInterval(recurringCheck);
  };
}

module.exports = { startUpdater };
