import { useState } from 'react';
import { Icon } from './Icon';

const releases = [
  { version: '0.1.16', changes: ['failed uploads now stand out in red', 'queue and activity stay fixed while their lists scroll', 'smoother overflow with stable, lightweight scrollbars', 'a cleaner complete update history'] },
  { version: '0.1.15', changes: ['cleaner navigation and account menu', 'a simpler studio connection indicator', 'release notes now appear on the main page', 'automatic update logic moved into its own module'] },
  { version: '0.1.14', changes: ['roblox sign-in with oauth and pkce', 'secure sessions with automatic token refresh', 'automatic animation and sound uploads', 'rebuilt studio plugin and automatic updates'] },
  { version: '0.1.0', changes: ['first lettuce desktop and studio release'] },
] as const;

export function ReleaseNotice({ version }: { version: string }) {
  const storageKey = `lettuce:release-notes:${version}`;
  const [visible, setVisible] = useState(() => localStorage.getItem(storageKey) !== 'seen');
  if (!visible) return null;
  const current = releases.find((release) => release.version === version) ?? { version, changes: ['fixes and improvements'] };
  const previous = releases.filter((release) => release.version !== current.version);
  const dismiss = () => { localStorage.setItem(storageKey, 'seen'); setVisible(false); };
  return <aside className="release-notice" aria-label={`What's new in Lettuce ${version}`}>
    <div className="release-notice-head"><span>what's new</span><button className="icon-button" aria-label="Dismiss" onClick={dismiss}><Icon name="close" /></button></div>
    <strong>lettuce v{current.version}</strong>
    <ul>{current.changes.map((change) => <li key={change}>{change}</li>)}</ul>
    <details>
      <summary>previous updates</summary>
      <div className="release-history scroll-area">
        {previous.map((release) => <section key={release.version}><strong>v{release.version}</strong><ul>{release.changes.map((change) => <li key={change}>{change}</li>)}</ul></section>)}
      </div>
    </details>
  </aside>;
}
