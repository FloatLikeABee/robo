export function providerSaveReady({ provider, model, apiKey, keySet, providers }) {
  const id = String(provider || '').trim();
  const chosen = String(model || '').trim();
  const info = (providers || []).find((item) => item.id === id);
  if (!info || !chosen) return false;
  if (info.requires_key && !String(apiKey || '').trim() && !keySet) return false;
  return true;
}

export function modelIdsFromScan(payload) {
  const rows = Array.isArray(payload?.models) ? payload.models : [];
  const ids = [];
  rows.forEach((row) => {
    const id = typeof row === 'string' ? row : row?.id;
    const name = String(id || '').trim();
    if (name && !ids.includes(name)) ids.push(name);
  });
  return ids;
}
