import { useEffect, useRef, useState } from 'react';
import type { RobloxUser } from '../domain/types';
import { Icon } from './Icon';

export function ProfileMenu({ user, onSettings, onLogout }: { user: RobloxUser; onSettings(): void; onLogout(): void }) {
  const [open, setOpen] = useState(false);
  const root = useRef<HTMLDivElement>(null);
  useEffect(() => { const close = (event: MouseEvent) => { if (!root.current?.contains(event.target as Node)) setOpen(false); }; addEventListener('mousedown', close); return () => removeEventListener('mousedown', close); }, []);
  return <div className="profile" ref={root}>
    {user.avatarUrl ? <img src={user.avatarUrl} alt="" /> : <span className="avatar-fallback">{(user.displayName || user.username)[0]}</span>}
    <div className="identity"><strong>{user.displayName || user.username}</strong><span>@{user.username}</span></div>
    <button className="icon-button" aria-label="Account menu" aria-expanded={open} onClick={() => setOpen(!open)}><Icon name="more" /></button>
    {open && <div className="menu"><button onClick={() => { setOpen(false); onSettings(); }}><Icon name="settings" />settings</button><button className="danger" onClick={onLogout}><Icon name="logout" />log out</button></div>}
  </div>;
}
