'use client';

import { useEffect } from 'react';
import { useAuthStore } from '@/stores/auth';

export function AuthHydration() {
  const hydrate = useAuthStore((s) => s.hydrate);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  return null;
}
