import type { LogEntry } from '../domain/types';

export function ActivityView({ logs }: { logs: LogEntry[] }) {
  return <div className="activity-page">
    <div className="page-heading"><div><h1>activity</h1><p>uploads, retries, and studio stuff.</p></div></div>
    <section className="panel logs scroll-area" aria-label="Activity log">
      {logs.length > 0 ? logs.map((log) => <div className={`log-row ${log.level}`} key={log.id}>
        <time>{log.at}</time><span aria-hidden="true" /><p>{log.message}</p>
      </div>) : <div className="empty"><h3>quiet in here</h3><p>events show up when studio connects.</p></div>}
    </section>
  </div>;
}
