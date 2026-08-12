import type { ReactNode, SVGProps } from 'react';
const paths: Record<string, ReactNode> = {
  queue: <><rect x="4" y="5" width="16" height="4" rx="1.5"/><rect x="4" y="15" width="16" height="4" rx="1.5"/><path d="M8 9v6m8-6v6"/></>,
  activity: <><path d="M4 19V9m5 10V5m6 14v-7m5 7V3"/></>, settings: <><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06-2.83 2.83-.06-.06A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 .6 1.7 1.7 0 0 0-.4 1.1V21H9.6v-.08A1.7 1.7 0 0 0 8.5 19.4a1.7 1.7 0 0 0-1.88.34l-.06.06-2.83-2.83.06-.06A1.7 1.7 0 0 0 4.1 15a1.7 1.7 0 0 0-1.52-1H2V10h.58A1.7 1.7 0 0 0 4.1 9a1.7 1.7 0 0 0-.34-1.88l-.06-.06 2.83-2.83.06.06A1.7 1.7 0 0 0 8.5 4.6 1.7 1.7 0 0 0 9.5 3H14a1.7 1.7 0 0 0 1 1.6 1.7 1.7 0 0 0 1.88-.34l.06-.06 2.83 2.83-.06.06A1.7 1.7 0 0 0 19.4 9a1.7 1.7 0 0 0 1.52 1H22v4h-1.08A1.7 1.7 0 0 0 19.4 15Z"/></>,
  more: <><circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/></>,
  logout: <><path d="M10 17l5-5-5-5m5 5H3"/><path d="M14 4h5a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-5"/></>,
  retry: <><path d="M20 11a8 8 0 1 0-2.34 5.66"/><path d="M20 4v7h-7"/></>, close: <path d="M6 6l12 12M18 6 6 18"/>, minimize: <path d="M5 12h14"/>, check: <path d="m5 12 4 4L19 6"/>, plugin: <><path d="M8 3v4m8-4v4M5 7h14v5a7 7 0 0 1-14 0Z"/><path d="M9 21v-3m6 3v-3"/></>,
  discord: <path d="M19 5.5A16 16 0 0 0 15 4l-.5 1.1a14 14 0 0 0-5 0L9 4a16 16 0 0 0-4 1.5C2.5 9.2 1.8 12.8 2 16.3A16 16 0 0 0 7 19l1.2-1.7a9 9 0 0 1-1.8-.9l.5-.4a11.5 11.5 0 0 0 10.2 0l.5.4a9 9 0 0 1-1.8.9L17 19a16 16 0 0 0 5-2.7c.2-3.5-.5-7.1-3-10.8ZM8.8 14.5c-1 0-1.8-1-1.8-2.2s.8-2.2 1.8-2.2 1.8 1 1.8 2.2-.8 2.2-1.8 2.2Zm6.4 0c-1 0-1.8-1-1.8-2.2s.8-2.2 1.8-2.2 1.8 1 1.8 2.2-.8 2.2-1.8 2.2Z"/>,
  sad: <><path d="M5 5h14v14H5z"/><path d="M8 9h2m4 0h2m-7 6c1.7-1.5 4.3-1.5 6 0"/></>,
  arrowUpRight: <><path d="M7 17 17 7"/><path d="M8 7h9v9"/></>,
};
const filledPaths: Record<string, ReactNode> = {
  queue: <><rect x="3.5" y="4" width="17" height="6" rx="2"/><rect x="3.5" y="14" width="17" height="6" rx="2"/><rect x="7" y="8.5" width="2" height="7" rx="1"/><rect x="15" y="8.5" width="2" height="7" rx="1"/></>,
  activity: <path d="M3 12h4v9H3zm7-7h4v16h-4zm7 4h4v12h-4z" />,
};
export function Icon({ name, filled = false, ...props }: SVGProps<SVGSVGElement> & { name: string; filled?: boolean }) { return <svg viewBox="0 0 24 24" fill={filled ? 'currentColor' : 'none'} stroke="currentColor" strokeWidth="2.15" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>{filled ? filledPaths[name] ?? paths[name] : paths[name]}</svg>; }
