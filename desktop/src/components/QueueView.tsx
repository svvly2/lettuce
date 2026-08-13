import type { UploadJob } from '../domain/types';
import { Icon } from './Icon';

export function QueueView({ jobs, retry, clear }: { jobs: UploadJob[]; retry(id: string): void; clear(): void }) {
  const active = jobs.filter((job) => job.status !== 'complete' && job.status !== 'failed').length;
  const done = jobs.filter((job) => job.status === 'complete').length;
  return <div className="queue-page">
    <header className="home-title"><div><h1>lettuce</h1><p>{active ? `${active} animation${active === 1 ? '' : 's'} moving` : 'ready for studio'}</p></div>{done > 0 && <button className="text-button" onClick={clear}>clear</button>}</header>
    {jobs.length === 0 ? <div className="home-empty"><span className="pixel-face"><Icon name="sad" /></span><h2>nothing here yet</h2><p>run the plugin in studio. uploads show up here.</p></div> : <div className="clean-list scroll-area" aria-label="Animation uploads">{jobs.map((job) => <article className={`clean-job ${job.status}`} key={job.id}>
      <span className={`status ${job.status}`}><Icon name={job.status === 'complete' ? 'check' : 'activity'} /></span>
      <div><strong>{job.name}</strong><small>{job.status}{job.resultAssetId ? ` · ${job.resultAssetId}` : ''}</small><div className="progress"><i style={{ width: `${job.progress}%` }} /></div></div>
      {job.status === 'failed' && <button className="icon-button" aria-label={`Retry ${job.name}`} onClick={() => retry(job.id)}><Icon name="retry" /></button>}
    </article>)}</div>}
  </div>;
}
