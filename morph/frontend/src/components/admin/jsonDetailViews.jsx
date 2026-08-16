import React from 'react';
import { Box, Typography, TextField, Button, Paper, Stack, useTheme } from '@mui/material';
import { JsonDetailTreeEditor } from './jsonDetailTreeEditor';
import { validateJsonDetailStructure, JSON_DETAIL_MAX_NESTING } from './jsonDetailValidate';
import { getDetailCardSx } from './detailViewTheme';

export { validateJsonDetailStructure, JSON_DETAIL_MAX_NESTING } from './jsonDetailValidate';

/** Normalize API `detail` (string or parsed object) for editors that expect JSON text. */
export function jsonDetailToString(value) {
  if (value == null) return '';
  if (typeof value === 'object' && !(value instanceof Date)) {
    try {
      return JSON.stringify(value, null, 2);
    } catch {
      return '{}';
    }
  }
  return String(value);
}

/** Labels for JSON keys (detail drawer): snake_case → readable title. */
function prettyJsonKeyLabel(key) {
  let s = String(key).trim().replace(/_/g, ' ');
  s = s.replace(/([a-z0-9])([A-Z])/g, '$1 $2');
  s = s.replace(/\s+/g, ' ').trim();
  s = s.replace(/\b\w/g, (c) => c.toUpperCase());
  s = s.replace(/\bId\b/g, 'ID');
  return s;
}

function formatMaybeDateString(value) {
  if (typeof value !== 'string') return null;
  const s = value.trim();
  if (!s) return null;

  // ISO datetime / timestamp shapes (e.g. 2026-05-06T19:17:26Z)
  const isoDateTimeRe = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2}(\.\d{1,3})?)?(Z|[+-]\d{2}:\d{2})?$/i;
  const isoDateOnlyRe = /^\d{4}-\d{2}-\d{2}$/;
  if (!isoDateTimeRe.test(s) && !isoDateOnlyRe.test(s)) return null;

  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return null;

  if (isoDateOnlyRe.test(s)) {
    return d.toLocaleDateString([], {
      year: 'numeric',
      month: 'short',
      day: '2-digit',
    });
  }

  return d.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  });
}

/** Recursive structured view for parsed JSON (no raw stringify for nested objects). */
function JsonStructuredValue({ value, depth }) {
  const borderAccent = depth > 0 ? 'divider' : 'transparent';

  if (value === null || value === undefined) {
    return (
      <Typography component="span" variant="body2" color="text.secondary">
        —
      </Typography>
    );
  }
  if (typeof value === 'boolean') {
    return (
      <Typography component="span" variant="body2" sx={{ fontWeight: 600, color: value ? 'success.main' : 'text.secondary' }}>
        {value ? 'Yes' : 'No'}
      </Typography>
    );
  }
  if (typeof value === 'number') {
    return (
      <Typography component="span" variant="body2" sx={{ fontVariantNumeric: 'tabular-nums' }}>
        {value}
      </Typography>
    );
  }
  if (typeof value === 'string') {
    const prettyDate = formatMaybeDateString(value);
    return (
      <Typography variant="body2" sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap', lineHeight: 1.45 }}>
        {prettyDate || value}
      </Typography>
    );
  }
  if (Array.isArray(value)) {
    if (value.length === 0) {
      return (
        <Typography variant="caption" color="text.secondary">
          Empty list
        </Typography>
      );
    }
    return (
      <Stack spacing={1} sx={{ pl: depth ? 1 : 0, borderLeft: depth ? '2px solid' : 'none', borderColor: borderAccent }}>
        {value.map((item, i) => (
          <Box key={i}>
            <Typography variant="caption" sx={{ fontWeight: 700, color: 'primary.main', display: 'block', mb: 0.25 }}>
              Item {i + 1}
            </Typography>
            <Box sx={{ pl: 0.5 }}>
              <JsonStructuredValue value={item} depth={depth + 1} />
            </Box>
          </Box>
        ))}
      </Stack>
    );
  }
  if (typeof value === 'object') {
    const keys = Object.keys(value).sort();
    if (keys.length === 0) {
      return (
        <Typography variant="caption" color="text.secondary">
          Empty
        </Typography>
      );
    }
    return (
      <Box
        sx={{
          m: 0,
          width: '100%',
          pl: depth ? 1.5 : 0,
          borderLeft: depth ? '2px solid' : 'none',
          borderColor: borderAccent,
        }}
      >
        <Stack spacing={1.25} sx={{ width: '100%' }}>
          {keys.map((k) => (
            <Box key={k} sx={{ width: '100%', minWidth: 0 }}>
              <Typography component="div" variant="caption" color="text.secondary" sx={{ fontWeight: 700, mb: 0.35, display: 'block' }}>
                {prettyJsonKeyLabel(k)}
              </Typography>
              <Box sx={{ width: '100%', minWidth: 0 }}>
                <JsonStructuredValue value={value[k]} depth={depth + 1} />
              </Box>
            </Box>
          ))}
        </Stack>
      </Box>
    );
  }
  return (
    <Typography variant="body2" sx={{ fontFamily: 'ui-monospace, monospace', fontSize: 12 }}>
      {String(value)}
    </Typography>
  );
}

/** Parsed JSON: top-level sections (cards); nested objects as labeled grids — not compact JSON strings. */
function JsonDetailValueTable({ parsed, fallbackStr, parseErr }) {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const jsonCardSx = { p: 1.5, ...getDetailCardSx(isDark, { nested: true }) };

  if (parseErr) {
    return (
      <Typography component="pre" variant="body2" sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', fontFamily: 'monospace', fontSize: 12, m: 0 }}>
        {fallbackStr}
      </Typography>
    );
  }
  if (parsed !== null && typeof parsed === 'object' && !Array.isArray(parsed)) {
    const keys = Object.keys(parsed).sort();
    if (keys.length === 0) {
      return <Typography variant="body2" color="text.secondary">Empty object</Typography>;
    }
    return (
      <Stack spacing={1.25}>
        {keys.map((k) => (
          <Paper key={k} elevation={0} sx={jsonCardSx}>
            <Typography variant="subtitle2" fontWeight={700} sx={{ mb: 1, color: 'primary.main', letterSpacing: 0.02 }}>
              {prettyJsonKeyLabel(k)}
            </Typography>
            <JsonStructuredValue value={parsed[k]} depth={0} />
          </Paper>
        ))}
      </Stack>
    );
  }
  if (Array.isArray(parsed)) {
    if (parsed.length === 0) {
      return <Typography variant="body2" color="text.secondary">Empty list</Typography>;
    }
    return (
      <Stack spacing={1.25}>
        {parsed.map((item, i) => (
          <Paper key={i} elevation={0} sx={jsonCardSx}>
            <Typography variant="subtitle2" fontWeight={700} sx={{ mb: 1, color: 'primary.main' }}>
              Item {i + 1}
            </Typography>
            <JsonStructuredValue value={item} depth={0} />
          </Paper>
        ))}
      </Stack>
    );
  }
  if (parsed !== null && typeof parsed !== 'object') {
    return (
      <Typography variant="body2" sx={{ fontFamily: 'ui-monospace, monospace', fontSize: 13 }}>
        {String(parsed)}
      </Typography>
    );
  }
  return (
    <Typography component="pre" variant="body2" sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', fontFamily: 'monospace', fontSize: 12, m: 0 }}>
      {fallbackStr}
    </Typography>
  );
}

/** Record sheet / print: structured sections (not raw monospace JSON). */
export function RecordSheetJsonDetail({ raw }) {
  if (raw === null || raw === undefined || raw === '') {
    return (
      <Typography variant="body2" color="text.secondary" sx={{ fontFamily: 'inherit' }}>
        —
      </Typography>
    );
  }
  let parsed = null;
  let parseErr = null;
  let fallbackStr = '';
  if (typeof raw === 'object' && raw !== null && !(raw instanceof Date)) {
    parsed = raw;
  } else {
    const t = String(raw).trim();
    if (!t) {
      return (
        <Typography variant="body2" color="text.secondary" sx={{ fontFamily: 'inherit' }}>
          —
        </Typography>
      );
    }
    fallbackStr = t;
    try {
      parsed = JSON.parse(t);
    } catch (e) {
      parseErr = e?.message || 'Invalid JSON';
    }
  }
  return (
    <Box sx={{ fontFamily: 'inherit' }}>
      <JsonDetailValueTable parsed={parsed} fallbackStr={fallbackStr} parseErr={parseErr} />
    </Box>
  );
}

/**
 * Vehicle / facility-style JSON field:
 * - preview: structured cards (original layout)
 * - tree: tree editor (max 5 nesting layers)
 * - raw: JSON textarea
 * When externalEditControls is true the caller supplies the mode toolbar.
 */
export function JsonDetailEditor({ value, onChange, mode = 'preview', onModeChange, externalEditControls }) {
  const str = jsonDetailToString(value);
  const trimmed = str.trim();
  const empty = trimmed === '';

  let parsed = null;
  let parseErr = null;
  if (!empty) {
    try {
      parsed = JSON.parse(str);
    } catch (e) {
      parseErr = e?.message || 'Invalid JSON';
    }
  }

  const structureCheck = !parseErr && !empty ? validateJsonDetailStructure(parsed) : { ok: true };
  const nestingBlocked = structureCheck.ok === false && !parseErr;
  const rawEdit = mode === 'raw';
  const treeEdit = mode === 'tree';

  if (empty) {
    return (
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.25 }}>
        <Typography variant="caption" color="text.secondary">
          Start with a structured empty object or list, or paste JSON below.
        </Typography>
        <Stack direction="row" flexWrap="wrap" gap={1}>
          <Button size="small" variant="outlined" onClick={() => onChange('{}')}>
            Empty object
          </Button>
          <Button size="small" variant="outlined" onClick={() => onChange('[]')}>
            Empty list
          </Button>
        </Stack>
        <TextField
          size="small"
          fullWidth
          multiline
          minRows={6}
          placeholder='Paste JSON, e.g. {"make":"Thomas Built","model":"Saf-T-Liner C2","year":2020}'
          defaultValue=""
          onChange={(e) => onChange(e.target.value)}
          sx={{ '& .MuiInputBase-input': { fontFamily: 'ui-monospace, monospace', fontSize: 13 } }}
        />
      </Box>
    );
  }

  if (rawEdit) {
    return (
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
        <TextField
          size="small"
          fullWidth
          multiline
          minRows={12}
          value={str}
          onChange={(e) => onChange(e.target.value)}
          sx={{ '& .MuiInputBase-input': { fontFamily: 'ui-monospace, monospace', fontSize: 13 } }}
        />
        {!externalEditControls ? (
          <Stack direction="row" flexWrap="wrap" gap={1}>
            <Button size="small" variant="outlined" onClick={() => onModeChange?.('tree')}>
              Tree editor
            </Button>
            <Button size="small" variant="outlined" onClick={() => onModeChange?.('preview')}>
              Card view
            </Button>
          </Stack>
        ) : null}
      </Box>
    );
  }

  if (treeEdit) {
    return (
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.25 }}>
        {parseErr ? (
          <Typography variant="caption" color="error" display="block">
            {parseErr}
          </Typography>
        ) : null}
        {nestingBlocked ? (
          <Typography variant="caption" color="error" display="block">
            {structureCheck.error} Use &quot;Edit raw JSON&quot; to reduce nesting, then return here.
          </Typography>
        ) : null}
        {!parseErr && !nestingBlocked ? (
          <>
            <Typography variant="caption" color="text.secondary" display="block">
              Tree editor — max {JSON_DETAIL_MAX_NESTING} nesting layers. Saving is blocked if the structure exceeds this limit.
            </Typography>
            <JsonDetailTreeEditor
              parsed={parsed}
              onChangeString={(nextStr) => onChange(nextStr)}
              maxNesting={JSON_DETAIL_MAX_NESTING}
            />
          </>
        ) : (
          <JsonDetailValueTable parsed={parsed} fallbackStr={str} parseErr={parseErr} />
        )}
        {!externalEditControls ? (
          <Stack direction="row" flexWrap="wrap" gap={1}>
            <Button size="small" variant="outlined" onClick={() => onModeChange?.('preview')}>
              Card view
            </Button>
            <Button size="small" variant="outlined" onClick={() => onModeChange?.('raw')}>
              Edit raw JSON
            </Button>
          </Stack>
        ) : null}
      </Box>
    );
  }

  /* preview — original card layout */
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.25 }}>
      {parseErr ? (
        <Typography variant="caption" color="error" display="block">
          {parseErr}
        </Typography>
      ) : null}
      {nestingBlocked ? (
        <Typography variant="caption" color="error" display="block">
          {structureCheck.error} Open tree or raw JSON editor to fix this.
        </Typography>
      ) : null}
      <JsonDetailValueTable parsed={parsed} fallbackStr={str} parseErr={parseErr} />
      {!externalEditControls ? (
        <Box>
          <Button size="small" variant="contained" color="primary" onClick={() => onModeChange?.('tree')}>
            Edit JSON
          </Button>
        </Box>
      ) : null}
    </Box>
  );
}
