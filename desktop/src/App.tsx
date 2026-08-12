import { useState } from 'react';
import { Icon } from './components/Icon';
import { ProfileMenu } from './components/ProfileMenu';
import { QueueView } from './components/QueueView';
import { SettingsDialog } from './components/SettingsDialog';
import { Titlebar } from './components/Titlebar';
import { useDesktop } from './hooks/useDesktop';
import { desktop } from './services/desktop';
import discordLogo from './assets/discord.svg';
import lettuceLogo from './assets/lettuce.svg';

export function App() {
  const { snapshot, login, logout, retry, clearCompleted, updateSettings } = useDesktop();
  const [settings, setSettings] = useState(false);
  const [page, setPage] = useState<'queue' | 'activity'>('queue');
  const [loginState, setLoginState] = useState<'idle' | 'opening' | 'error'>('idle');
  const beginLogin = async () => { setLoginState('opening'); try { await login(); setLoginState('idle'); } catch { setLoginState('error'); } };
  if (!snapshot) return <div className="boot">lettuce</div>;
  if (!snapshot.user) return <div className="app login-app"><Titlebar /><main className="login-simple"><div className="login-brand"><span><img src={lettuceLogo} alt="" /></span><strong>lettuce</strong></div><div className="login-copy"><h1>animation uploads,<br />without the mess.</h1><p>connect roblox and get back to studio.</p></div><button className="login-primary" disabled={loginState === 'opening'} onClick={() => void beginLogin()}>{loginState === 'opening' ? 'opening roblox…' : 'continue with roblox'}<Icon name="arrowUpRight" /></button>{loginState === 'error' && <small className="login-error">couldn't open roblox. try again.</small>}<button className="discord-link" onClick={() => void desktop().openExternal('https://discord.gg/4ycV7TUX6G')}><img src={discordLogo} alt="" /><span>join the discord</span></button></main></div>;
  return <div className="app signed-in"><Titlebar />
    <main className="canvas">
      <header className="floating-header">
        <div className="studio-pill"><i className={snapshot.pluginConnected ? 'online' : ''} /><span>{snapshot.pluginConnected ? 'studio connected' : 'waiting for studio'}</span><small>:{snapshot.settings.port}</small></div>
        <ProfileMenu user={snapshot.user} onSettings={() => setSettings(true)} onLogout={() => void logout()} />
      </header>
      <section className="view" key={page}>
        {page === 'queue' ? <QueueView jobs={snapshot.queue} retry={(id) => void retry(id)} clear={() => void clearCompleted()} /> : <><div className="page-heading"><div><h1>activity</h1><p>uploads, retries, and studio stuff.</p></div></div><section className="panel logs">{snapshot.logs.length ? snapshot.logs.map((log) => <div key={log.id}><time>{log.at}</time><span className={log.level} /><p>{log.message}</p></div>) : <div className="empty"><h3>quiet in here</h3><p>events show up when studio connects.</p></div>}</section></>}
      </section>
      <nav className="dock" aria-label="Main navigation">
        <button className={page === 'queue' ? 'active' : ''} onClick={() => setPage('queue')}><Icon name="queue" filled={page === 'queue'} /><span>queue</span></button>
        <button className={page === 'activity' ? 'active' : ''} onClick={() => setPage('activity')}><Icon name="activity" filled={page === 'activity'} /><span>activity</span></button>
        <button onClick={() => setSettings(true)}><Icon name="settings" /><span>settings</span></button>
      </nav>
    </main>
    {settings && <SettingsDialog value={snapshot.settings} close={() => setSettings(false)} save={(value) => { void updateSettings(value); setSettings(false); }} />}
  </div>;
}
