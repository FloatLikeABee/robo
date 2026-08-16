import React, { useCallback, useEffect, useState } from 'react';
import { API_BASE_URL } from '../apiBase';
import { getMorphToken } from '../auth/morphSession';

async function skillsFetch(path, opts = {}) {
  const token = getMorphToken();
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...opts,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(opts.headers || {}),
    },
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || `Request failed (${res.status})`);
  return data;
}

/** Skills catalog body — used inside the modal and the /skills fallback page. */
export function SkillsPanel({ onClose, embedded = false }) {
  const [skills, setSkills] = useState([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [instructions, setInstructions] = useState('');
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await skillsFetch('/api/skills');
      setSkills(Array.isArray(data.skills) ? data.skills : Array.isArray(data) ? data : []);
    } catch (e) {
      setError(e.message || 'Failed to load skills');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function onUpload(e) {
    e.preventDefault();
    setSaving(true);
    setError('');
    try {
      await skillsFetch('/api/skills', {
        method: 'POST',
        body: JSON.stringify({
          name: name.trim(),
          description: description.trim(),
          instructions: instructions.trim(),
          body: instructions.trim(),
        }),
      });
      setName('');
      setDescription('');
      setInstructions('');
      await load();
    } catch (err) {
      setError(err.message || 'Upload failed');
    } finally {
      setSaving(false);
    }
  }

  async function toggleEnabled(skill) {
    try {
      await skillsFetch(`/api/skills/${encodeURIComponent(skill.id)}`, {
        method: 'PATCH',
        body: JSON.stringify({ enabled: !skill.enabled }),
      });
      await load();
    } catch (err) {
      setError(err.message || 'Update failed');
    }
  }

  async function removeSkill(skill) {
    if (!window.confirm(`Delete skill “${skill.name}”?`)) return;
    try {
      await skillsFetch(`/api/skills/${encodeURIComponent(skill.id)}`, { method: 'DELETE' });
      await load();
    } catch (err) {
      setError(err.message || 'Delete failed');
    }
  }

  return (
    <div className={`skills-panel${embedded ? ' skills-panel--embedded' : ''}`}>
      <header className="skills-panel-header">
        <div>
          <h2 id="skills-modal-title" className="skills-panel-title">
            Skills
          </h2>
          <p className="skills-panel-hint">
            Upload skills for Morph AI assistants. Syncs to Neo4j asynchronously when configured.
          </p>
        </div>
        {onClose ? (
          <button type="button" className="skills-panel-close" onClick={onClose} aria-label="Close skills">
            ×
          </button>
        ) : null}
      </header>

      {error ? (
        <div className="skills-panel-alert" role="alert">
          {error}
        </div>
      ) : null}

      <form className="skills-panel-form" onSubmit={onUpload}>
        <strong className="skills-panel-form-title">Upload skill</strong>
        <label className="skills-panel-field">
          Name
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </label>
        <label className="skills-panel-field">
          Description
          <input value={description} onChange={(e) => setDescription(e.target.value)} />
        </label>
        <label className="skills-panel-field">
          Instructions
          <textarea
            value={instructions}
            onChange={(e) => setInstructions(e.target.value)}
            required
            rows={5}
          />
        </label>
        <button type="submit" className="skills-panel-primary" disabled={saving}>
          {saving ? 'Uploading…' : 'Upload'}
        </button>
      </form>

      <section className="skills-panel-catalog">
        <h3 className="skills-panel-catalog-title">Catalog</h3>
        {loading ? (
          <p className="skills-panel-empty">Loading…</p>
        ) : skills.length === 0 ? (
          <p className="skills-panel-empty">No skills yet.</p>
        ) : (
          <ul className="skills-panel-list">
            {skills.map((s) => (
              <li key={s.id} className="skills-panel-item">
                <div className="skills-panel-item-main">
                  <strong>{s.name}</strong>
                  {s.description ? <p className="skills-panel-item-desc">{s.description}</p> : null}
                  <p className="skills-panel-item-meta">
                    {s.enabled ? 'Enabled' : 'Disabled'}
                    {s.builtin || s.is_builtin ? ' · built-in' : ''}
                  </p>
                </div>
                <div className="skills-panel-item-actions">
                  <button type="button" className="skills-panel-ghost" onClick={() => void toggleEnabled(s)}>
                    {s.enabled ? 'Disable' : 'Enable'}
                  </button>
                  {!s.builtin && !s.is_builtin ? (
                    <button type="button" className="skills-panel-ghost" onClick={() => void removeSkill(s)}>
                      Delete
                    </button>
                  ) : null}
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}

/** Overlay modal for the Morph AI chat header. */
export default function SkillsModal({ open, onClose }) {
  useEffect(() => {
    if (!open) return undefined;
    const onKey = (e) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      window.removeEventListener('keydown', onKey);
      document.body.style.overflow = prev;
    };
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div className="skills-modal-overlay" role="presentation" onClick={onClose}>
      <div
        className="skills-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="skills-modal-title"
        onClick={(e) => e.stopPropagation()}
      >
        <SkillsPanel onClose={onClose} embedded />
      </div>
    </div>
  );
}
