const releases = [
  {
    version: '0.1.14',
    label: 'latest',
    changes: [
      'roblox sign-in with oauth 2.0 and pkce',
      'secure sessions with automatic token refresh',
      'automatic animation and sound upload support',
      'rebuilt studio plugin with discovery and replacement',
      'new desktop design, icons, and automatic updates',
    ],
  },
  {
    version: '0.1.0',
    changes: ['first lettuce desktop and studio release'],
  },
] as const;

export function ReleaseNotes({ currentVersion }: { currentVersion: string }) {
  return <>
    <div className="page-heading release-heading"><div><h1>updates</h1><p>new stuff, without the essay.</p></div><span>v{currentVersion}</span></div>
    <section className="release-list">
      {releases.map((release) => <article className="release-card" key={release.version}>
        <header><strong>v{release.version}</strong>{'label' in release && <span>{release.label}</span>}</header>
        <ul>{release.changes.map((change) => <li key={change}>{change}</li>)}</ul>
      </article>)}
    </section>
  </>;
}
