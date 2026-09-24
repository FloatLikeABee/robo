import { useEffect, useState } from 'react';
import { fetchAuthMe, patchAuthMe } from './auth';

type Props = {
  open: boolean;
  onClose: () => void;
};

export default function UserProfileModal({ open, onClose }: Props) {
  const [username, setUsername] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    if (!open) return;
    setError('');
    setSuccess('');
    setCurrentPassword('');
    setNewPassword('');
    setLoading(true);
    void fetchAuthMe()
      .then((u) => setUsername(u.username || ''))
      .catch((e) => setError(e instanceof Error ? e.message : 'Failed to load profile'))
      .finally(() => setLoading(false));
  }, [open]);

  if (!open) return null;

  async function onSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError('');
    setSuccess('');
    try {
      const body: { username?: string; password?: string; current_password?: string } = {};
      const trimmed = username.trim();
      if (trimmed) body.username = trimmed;
      if (newPassword) {
        body.password = newPassword;
        body.current_password = currentPassword;
      }
      const updated = await patchAuthMe(body);
      setUsername(updated.username || trimmed);
      setCurrentPassword('');
      setNewPassword('');
      setSuccess('Account updated.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Save failed');
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="mu-profile-backdrop" role="presentation" onClick={onClose}>
      <div
        className="mu-profile-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="mu-profile-title"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="mu-profile-head">
          <h2 id="mu-profile-title">Your account</h2>
          <button type="button" className="mu-profile-close" onClick={onClose} aria-label="Close">✕</button>
        </header>
        {error ? <p className="mu-profile-error">{error}</p> : null}
        {success ? <p className="mu-profile-success">{success}</p> : null}
        {loading ? (
          <p className="mu-profile-muted">Loading…</p>
        ) : (
          <form className="mu-profile-form" onSubmit={(e) => void onSave(e)}>
            <label>
              Username
              <input value={username} onChange={(e) => setUsername(e.target.value)} autoComplete="username" />
            </label>
            <label>
              New password
              <input type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} autoComplete="new-password" placeholder="Leave blank to keep" />
            </label>
            {newPassword ? (
              <label>
                Current password
                <input type="password" value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)} autoComplete="current-password" required />
              </label>
            ) : null}
            <div className="mu-profile-actions">
              <button type="button" className="mu-profile-secondary" onClick={onClose}>Cancel</button>
              <button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save'}</button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
