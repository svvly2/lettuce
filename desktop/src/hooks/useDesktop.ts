import { useCallback, useEffect, useState } from 'react';
import type { AppSnapshot, Settings } from '../domain/types';
import { desktop } from '../services/desktop';

export function useDesktop() {
  const [snapshot, setSnapshot] = useState<AppSnapshot | null>(null);
  useEffect(() => {
    let mounted = true;
    void desktop().snapshot().then((value) => mounted && setSnapshot(value));
    const unsubscribe = desktop().onEvent((event) => { if (event.type === 'snapshot') setSnapshot(event.payload); });
    return () => { mounted = false; unsubscribe(); };
  }, []);
  const updateSettings = useCallback(async (settings: Settings) => { await desktop().updateSettings(settings); setSnapshot((s) => s ? { ...s, settings } : s); }, []);
  return { snapshot, login: desktop().login, logout: desktop().logout, retry: desktop().retry, clearCompleted: desktop().clearCompleted, updateSettings };
}
