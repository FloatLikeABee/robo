import React, { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Tab,
  Tabs,
  TextField,
  Typography,
} from '@mui/material';
import ReactQuill from 'react-quill';
import 'react-quill/dist/quill.snow.css';
import { tranApi, tranEndpoints } from '../../api/tranClient';

const QUILL_MODULES = {
  toolbar: [
    [{ header: [1, 2, 3, false] }],
    ['bold', 'italic', 'underline', 'strike'],
    [{ color: [] }, { background: [] }],
    [{ list: 'ordered' }, { list: 'bullet' }],
    [{ indent: '-1' }, { indent: '+1' }],
    ['blockquote', 'link'],
    ['clean'],
  ],
};

const QUILL_FORMATS = [
  'header',
  'bold',
  'italic',
  'underline',
  'strike',
  'color',
  'background',
  'list',
  'bullet',
  'indent',
  'blockquote',
  'link',
];

function recipientKey(r) {
  return `${r.kind}:${r.id}`;
}

function parseLocation(raw) {
  if (!raw) return { label: '', areaLen: 0 };
  const parsed = typeof raw === 'string' ? (() => {
    try {
      return JSON.parse(raw);
    } catch {
      return { label: String(raw || ''), area: [] };
    }
  })() : raw;
  const label =
    typeof parsed?.label === 'string'
      ? parsed.label
      : typeof parsed?.location === 'string'
        ? parsed.location
        : '';
  const arr = Array.isArray(parsed?.area) ? parsed.area : [];
  return { label, areaLen: arr.length };
}

function locationSummary(raw) {
  const p = parseLocation(raw);
  if (!p.label && p.areaLen === 0) return '';
  if (p.label && p.areaLen > 0) return `${p.label} (${p.areaLen} map points)`;
  if (p.label) return p.label;
  return `Area (${p.areaLen} map points)`;
}

function dateDisplay(value) {
  if (!value) return '—';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleString([], {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
}

function TabPanel({ hidden, children }) {
  if (hidden) return null;
  return <Box sx={{ pt: 2 }}>{children}</Box>;
}

export default function CaseTaskEmailDialog({
  open,
  onClose,
  caseTaskId,
  taskTitle,
  initialSubject,
  initialHtmlBody,
  recipientOptions,
  initialRecipients,
  onSent,
}) {
  const [tab, setTab] = useState(0);
  const [subject, setSubject] = useState('');
  const [htmlBody, setHtmlBody] = useState('');
  const [recipients, setRecipients] = useState([]);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState(null);
  const [summary, setSummary] = useState(null);
  const [summaryLoading, setSummaryLoading] = useState(false);
  const [summaryError, setSummaryError] = useState('');

  const optionByKey = useMemo(() => {
    const m = new Map();
    (recipientOptions || []).forEach((o) => m.set(recipientKey(o), o));
    return m;
  }, [recipientOptions]);

  useEffect(() => {
    if (!open) return;
    setTab(0);
    setError('');
    setResult(null);
    setSubject(initialSubject || '');
    setHtmlBody(initialHtmlBody || '<p></p>');
    const init = (initialRecipients || [])
      .map((r) => optionByKey.get(recipientKey(r)))
      .filter(Boolean);
    setRecipients(init);
  }, [open, initialSubject, initialHtmlBody, initialRecipients, optionByKey]);

  useEffect(() => {
    if (!open || !caseTaskId) {
      setSummary(null);
      setSummaryError('');
      return;
    }
    let cancelled = false;
    setSummaryLoading(true);
    setSummaryError('');
    tranApi
      .get(tranEndpoints.caseTaskFull(caseTaskId))
      .then((res) => {
        if (!cancelled) setSummary(res.data || {});
      })
      .catch((err) => {
        if (!cancelled) {
          setSummaryError(err.response?.data?.error || err.message || 'Could not load case/task summary.');
          setSummary(null);
        }
      })
      .finally(() => {
        if (!cancelled) setSummaryLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [open, caseTaskId]);

  const send = async () => {
    if (!caseTaskId) return;
    setError('');
    setResult(null);
    const sub = String(subject || '').trim();
    const html = String(htmlBody || '').replace(/<p><br><\/p>\s*$/i, '').trim();
    if (!sub) {
      setError('Subject is required.');
      return;
    }
    if (!html || html === '<p></p>' || html === '<p><br></p>') {
      setError('Message body is required.');
      return;
    }
    if (!recipients.length) {
      setError('Select at least one recipient with an email on file.');
      return;
    }
    const noEmail = recipients.filter((r) => !String(r.email || '').trim());
    if (noEmail.length) {
      setError('Every recipient must have an email. Remove entries without email or update the person in the roster.');
      return;
    }
    setSending(true);
    try {
      const res = await tranApi.post(tranEndpoints.caseTaskSendEmail(caseTaskId), {
        subject: sub,
        html_body: html,
        recipients: recipients.map((r) => ({ kind: r.kind, id: r.id })),
      });
      setResult(res.data);
      onSent?.(res.data);
    } catch (err) {
      const msg = err.response?.data?.error || err.message || 'Send failed.';
      const skipped = err.response?.data?.skipped;
      setError(skipped?.length ? `${msg} (${skipped.join(', ')})` : msg);
    } finally {
      setSending(false);
    }
  };

  return (
    <Dialog open={open} onClose={sending ? undefined : onClose} maxWidth="md" fullWidth scroll="paper">
      <DialogTitle>Send case/task by email</DialogTitle>
      <DialogContent dividers>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Edit the message, preview how it will look, and choose recipients. Each recipient gets the same HTML email
          (plus a plain-text fallback for clients that need it).
        </Typography>

        <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tab label="Compose" />
          <Tab label="Preview HTML" />
          <Tab label="Case/task summary" />
        </Tabs>

        <TabPanel hidden={tab !== 0}>
          <TextField
            label="Subject"
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            fullWidth
            size="small"
            sx={{ mb: 2 }}
          />
          <Typography variant="subtitle2" sx={{ mb: 0.5 }}>
            Message
          </Typography>
          <Box
            sx={{
              '& .ql-toolbar': {
                borderTopLeftRadius: 8,
                borderTopRightRadius: 8,
                borderColor: 'divider',
                bgcolor: 'action.hover',
              },
              '& .ql-container': {
                borderBottomLeftRadius: 8,
                borderBottomRightRadius: 8,
                borderColor: 'divider',
                minHeight: 220,
                fontSize: '0.95rem',
              },
              '& .ql-editor': { minHeight: 200 },
              mb: 2,
            }}
          >
            <ReactQuill theme="snow" value={htmlBody} onChange={setHtmlBody} modules={QUILL_MODULES} formats={QUILL_FORMATS} />
          </Box>
          <Autocomplete
            multiple
            options={recipientOptions || []}
            groupBy={(o) => o.group}
            getOptionLabel={(o) => o.label}
            isOptionEqualToValue={(a, b) => a.kind === b.kind && a.id === b.id}
            getOptionDisabled={(o) => !String(o.email || '').trim()}
            value={recipients}
            onChange={(_, v) => setRecipients(v)}
            filterSelectedOptions
            renderTags={(value, getTagProps) =>
              value.map((opt, index) => (
                <Chip
                  {...getTagProps({ index })}
                  key={recipientKey(opt)}
                  label={opt.email ? `${opt.label} · ${opt.email}` : opt.label}
                  size="small"
                />
              ))
            }
            renderInput={(params) => (
              <TextField
                {...params}
                label="Recipients"
                placeholder="Search people…"
                helperText="Only people with an email can be selected. Each chosen person is emailed separately."
              />
            )}
          />
        </TabPanel>

        <TabPanel hidden={tab !== 1}>
          <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1 }}>
            Approximate appearance in email clients (inline styles may vary).
          </Typography>
          <Box
            sx={{
              border: 1,
              borderColor: 'divider',
              borderRadius: 1,
              p: 2,
              bgcolor: 'background.default',
              maxHeight: 360,
              overflow: 'auto',
            }}
            dangerouslySetInnerHTML={{ __html: htmlBody || '' }}
          />
        </TabPanel>

        <TabPanel hidden={tab !== 2}>
          {summaryLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress size={32} />
            </Box>
          ) : null}
          {summaryError ? (
            <Alert severity="warning" sx={{ mb: 1 }}>
              {summaryError}
            </Alert>
          ) : null}
          {!summaryLoading && summary ? (
            <Box
              sx={{
                border: 1,
                borderColor: 'divider',
                borderRadius: 1,
                p: 2,
                maxHeight: 480,
                overflow: 'auto',
                bgcolor: 'background.default',
              }}
            >
              <Typography variant="subtitle1" fontWeight={700}>
                {summary.title || 'Untitled'}
              </Typography>
              <Divider sx={{ my: 1.5 }} />
              {summary.description ? (
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary" fontWeight={600} display="block">
                    Description
                  </Typography>
                  <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap', mt: 0.5 }}>
                    {summary.description}
                  </Typography>
                </Box>
              ) : null}
              <Typography variant="caption" color="text.secondary" fontWeight={600} display="block">
                Assignees
              </Typography>
              <Box sx={{ mt: 0.75, mb: 2, display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                {(summary.assignees || []).length ? (
                  (summary.assignees || []).map((a) => (
                    <Chip
                      key={`${a.assignee_kind || a.kind}-${a.assignee_id ?? a.id}`}
                      size="small"
                      label={
                        a.email
                          ? `${a.name || '?'} (${a.email})`
                          : a.name || `${a.assignee_kind || a.kind} #${a.assignee_id ?? a.id}`
                      }
                      variant="outlined"
                    />
                  ))
                ) : (
                  <Typography variant="body2" color="text.secondary">
                    {summary.assignees_label || '—'}
                  </Typography>
                )}
              </Box>
              <Typography variant="caption" color="text.secondary" fontWeight={600} display="block">
                Schedule
              </Typography>
              <Typography variant="body2" sx={{ mt: 0.5, mb: 1.5 }}>
                Start: {dateDisplay(summary.start_at)} · End: {dateDisplay(summary.end_at)}
              </Typography>
              {locationSummary(summary.location) ? (
                <>
                  <Typography variant="caption" color="text.secondary" fontWeight={600} display="block">
                    Location / area
                  </Typography>
                  <Typography variant="body2" sx={{ mt: 0.5, mb: 1.5 }}>
                    {locationSummary(summary.location)}
                  </Typography>
                </>
              ) : null}
              {summary.attachments?.length ? (
                <>
                  <Typography variant="caption" color="text.secondary" fontWeight={600} display="block">
                    Attachments
                  </Typography>
                  <Box component="ul" sx={{ mt: 0.5, pl: 2.5, mb: 0 }}>
                    {summary.attachments.map((at) => (
                      <Typography key={at.id} component="li" variant="body2">
                        {at.original_name || `File #${at.id}`}
                      </Typography>
                    ))}
                  </Box>
                </>
              ) : null}
            </Box>
          ) : null}
          {!summaryLoading && !summary && !summaryError && caseTaskId ? (
            <Typography color="text.secondary">No summary loaded.</Typography>
          ) : null}
        </TabPanel>

        {taskTitle ? (
          <Typography variant="caption" color="text.secondary" sx={{ mt: 2, display: 'block' }}>
            Task: {taskTitle}
          </Typography>
        ) : null}
        {error ? (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        ) : null}
        {result ? (
          <Alert severity="success" sx={{ mt: 2 }}>
            Sent {result.sent} message(s).
            {result.skipped?.length ? ` Skipped: ${result.skipped.join(', ')}.` : ''}
          </Alert>
        ) : null}
      </DialogContent>
      <DialogActions sx={{ px: 3, py: 2 }}>
        <Button onClick={onClose} disabled={sending}>
          {result ? 'Close' : 'Cancel'}
        </Button>
        <Box sx={{ flex: 1 }} />
        <Button variant="contained" onClick={send} disabled={sending || !!result}>
          {sending ? 'Sending…' : 'Send email'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
