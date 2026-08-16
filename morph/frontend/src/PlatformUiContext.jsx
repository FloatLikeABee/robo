import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { tranApi, tranEndpoints } from './api/tranClient';
import { defaultPlatformUILabels, defaultEntityDictionaries } from './platformUiDefaults';

const PlatformUiContext = createContext(null);

export function PlatformUiProvider({ children }) {
  const [labels, setLabels] = useState(defaultPlatformUILabels);
  const [dictionaries, setDictionaries] = useState(defaultEntityDictionaries);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const refresh = useCallback(() => {
    setError(null);
    // Display labels are fixed product names — do not apply editable overrides from the API.
    setLabels(defaultPlatformUILabels);
    return tranApi
      .get(tranEndpoints.platformUiConfig)
      .then((res) => {
        const d = res.data?.dictionaries && typeof res.data.dictionaries === 'object' ? res.data.dictionaries : null;
        if (d) {
          setDictionaries((prev) => ({ ...defaultEntityDictionaries, ...d }));
        } else {
          setDictionaries(defaultEntityDictionaries);
        }
      })
      .catch((err) => {
        setDictionaries(defaultEntityDictionaries);
        setError(err.response?.data?.error || err.message || 'Failed to load platform config');
      })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const value = useMemo(
    () => ({
      labels,
      dictionaries,
      loading,
      error,
      refresh,
    }),
    [labels, dictionaries, loading, error, refresh],
  );

  return <PlatformUiContext.Provider value={value}>{children}</PlatformUiContext.Provider>;
}

export function usePlatformUi() {
  const ctx = useContext(PlatformUiContext);
  if (!ctx) {
    return {
      labels: defaultPlatformUILabels,
      dictionaries: defaultEntityDictionaries,
      loading: false,
      error: null,
      refresh: async () => {},
    };
  }
  return ctx;
}
