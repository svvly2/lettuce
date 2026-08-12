import { useState } from 'react';
import { Icon } from './Icon';

const notes: Record<string, readonly string[]> = {
  '0.1.15': [
    'cleaner navigation and account menu',
    'a simpler studio connection indicator',
    'release notes now appear right here',
  ],
  '0.1.14': [
    'roblox sign-in with oauth and pkce',
    'automatic animation and sound uploads',
    'rebuilt studio plugin and automatic updates',
  ],
};

export function ReleaseNotice({ version }: { version: string }) {
  const storageKey = `lettuce:release-notes:${version}`;
  const [visible, setVisible] = useState(() => localStorage.getItem(storageKey) !== 'seen');
  if (!visible) return null;
  const changes = notes[version] ?? ['lettuce was updated with a few fixes and improvements'];
  const dismiss = () => { localStorage.setItem(storageKey, 'seen'); setVisible(false); };
  return <aside className="release-notice" aria-label={`What's new in Lettuce ${version}`}>
    <div className="release-notice-head"><span>what's new</span><button className="icon-button" aria-label="Dismiss" onClick={dismiss}><Icon name="close" /></button></div>
    <strong>lettuce v{version}</strong>
    <ul>{changes.map((change) => <li key={change}>{change}</li>)}</ul>
  </aside>;
}
