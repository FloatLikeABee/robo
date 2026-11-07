import React, { useEffect, useId, useState } from 'react';
import { tranApi } from '../../api/tranClient';
import { modelIdsFromScan, providerSaveReady } from '../../lib/chatProviderForm';

const empty = {
  provider: '',
  baseURL: '',
  apiKey: '',
  model: '',
  keySet: false,
};

export default function ChatProviderSettings({ open, onClose }) {
  const titleId = useId();
  const [providers, setProviders] = useState([]);
  const [form, setForm] = useState(empty);
  const [models, setModels] = useState([]);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState('');

  useEffect(() => {
    if (!open) return undefined;
    const onKey = (event) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, onClose]);

  useEffect(() => {
    if (!open) return undefined;
    let gone = false;
    setError('');
    Promise.all([
      tranApi.get('/api/chat/providers'),
      tranApi.get('/api/chat/provider'),
    ]).then(([list, saved]) => {
      if (gone) return;
      const rows = list.data?.providers || [];
      setProviders(rows);
      const current = saved.data || {};
      const provider = current.provider || rows[0]?.id || '';
      const info = rows.find((item) => item.id === provider);
      setForm({
        provider,
        baseURL: current.base_url || info?.default_base_url || '',
        apiKey: '',
        model: current.model || '',
        keySet: Boolean(current.key_set),
      });
      setModels(current.model ? [current.model] : []);
    }).catch((err) => {
      if (!gone) setError(err.response?.data?.error || 'Could not load settings');
    });
    return () => {
      gone = true;
    };
  }, [open]);

  if (!open) return null;

  const selected = providers.find((item) => item.id === form.provider);
  const canSave = providerSaveReady({ ...form, providers });

  const setProvider = (provider) => {
    const info = providers.find((item) => item.id === provider);
    setForm((current) => ({
      ...current,
      provider,
      baseURL: info?.default_base_url || current.baseURL,
      model: '',
      keySet: current.provider === provider ? current.keySet : false,
    }));
    setModels([]);
  };

  const scan = async () => {
    setBusy('scan');
    setError('');
    try {
      const { data } = await tranApi.post('/api/chat/provider/scan', {
        provider: form.provider,
        base_url: form.baseURL,
        api_key: form.apiKey,
      });
      const ids = modelIdsFromScan(data);
      setModels(ids);
      setForm((current) => ({
        ...current,
        model: ids.includes(current.model) ? current.model : (ids[0] || ''),
      }));
      if (!ids.length) setError('No models reported');
    } catch (err) {
      setError(err.response?.data?.error || 'Could not scan models');
    } finally {
      setBusy('');
    }
  };

  const save = async (event) => {
    event.preventDefault();
    setBusy('save');
    setError('');
    try {
      const { data } = await tranApi.put('/api/chat/provider', {
        provider: form.provider,
        base_url: form.baseURL,
        api_key: form.apiKey,
        model: form.model,
      });
      setForm((current) => ({
        ...current,
        apiKey: '',
        keySet: Boolean(data?.key_set),
        model: data?.model || current.model,
      }));
      onClose();
    } catch (err) {
      setError(err.response?.data?.error || 'Could not save settings');
    } finally {
      setBusy('');
    }
  };

  return (
    <div className="skills-modal-overlay" onClick={onClose}>
      <div
        className="skills-modal provider-settings-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        onClick={(event) => event.stopPropagation()}
      >
        <form className="skills-panel" onSubmit={save}>
          <div className="skills-panel-header">
            <h2 id={titleId} className="skills-panel-title">AI provider</h2>
            <button type="button" className="skills-panel-close" onClick={onClose} aria-label="Close">×</button>
          </div>
          {error ? <p className="skills-panel-alert" role="alert">{error}</p> : null}
          <label className="skills-panel-field">
            Provider
            <select value={form.provider} onChange={(event) => setProvider(event.target.value)}>
              {providers.map((item) => (
                <option key={item.id} value={item.id}>{item.display_name}</option>
              ))}
            </select>
          </label>
          <label className="skills-panel-field">
            API URL
            <input
              value={form.baseURL}
              onChange={(event) => setForm((current) => ({ ...current, baseURL: event.target.value }))}
              autoComplete="off"
              spellCheck="false"
            />
          </label>
          <label className="skills-panel-field">
            API key
            <input
              type="password"
              value={form.apiKey}
              placeholder={form.keySet ? 'Saved key stays if this is empty' : ''}
              onChange={(event) => setForm((current) => ({ ...current, apiKey: event.target.value }))}
              autoComplete="off"
            />
          </label>
          {selected && !selected.requires_key ? (
            <p className="skills-panel-empty">Local models do not need a key.</p>
          ) : null}
          <div className="skills-panel-form-actions">
            <button type="button" className="skills-panel-secondary" disabled={busy !== '' || !form.provider} onClick={scan}>
              {busy === 'scan' ? 'Scanning…' : 'Scan'}
            </button>
          </div>
          <label className="skills-panel-field">
            Model
            <select
              value={form.model}
              onChange={(event) => setForm((current) => ({ ...current, model: event.target.value }))}
            >
              <option value="">Choose a model</option>
              {models.map((id) => (
                <option key={id} value={id}>{id}</option>
              ))}
            </select>
          </label>
          <div className="skills-panel-form-actions">
            <button type="submit" className="skills-panel-primary" disabled={!canSave || busy !== ''}>
              {busy === 'save' ? 'Saving…' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
