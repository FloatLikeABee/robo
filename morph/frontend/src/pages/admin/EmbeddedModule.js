import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Box, Button, CircularProgress, Typography } from '@mui/material';
import { getMorphToken } from '../../auth/morphSession';
import { MORPH_NOTES_EMBEDS, resolveEmbedUrl, withSessionToken } from '../../lib/morphNotesEmbeds';
import { bindFrameLocale, getLocale, postLocaleToFrame, withLang } from '../../lib/locale';
import { useT } from '../../lib/localeReact';

const LABEL_KEY = {
  'event-logs': 'eventLogs',
  'content-maker': 'contentMaker',
  project: 'project',
};

function configuredOrigin(id) {
  if (id === 'event-logs') {
    return process.env.REACT_APP_SHEETX_URL || process.env.REACT_APP_FORMSX_URL || '';
  }
  if (id === 'content-maker') return process.env.REACT_APP_COMPOSERX_URL || '';
  return process.env.REACT_APP_PROJECTS_URL || process.env.REACT_APP_MORPH_ENGI_URL || '';
}

async function probeOrigin(url) {
  try {
    await fetch(new URL(url).origin, { method: 'GET', mode: 'no-cors', cache: 'no-store' });
    return true;
  } catch {
    return false;
  }
}

export default function EmbeddedModule({ id }) {
  const t = useT();
  const frameRef = useRef(null);
  const spec = MORPH_NOTES_EMBEDS.find((item) => item.id === id);
  const label = t(LABEL_KEY[id] || 'project');
  const url = resolveEmbedUrl({
    id,
    nodeEnv: process.env.NODE_ENV,
    configured: configuredOrigin(id),
  });
  const [up, setUp] = useState(null);
  const [tryId, setTryId] = useState(0);
  const src = useMemo(
    () => (url ? withLang(withSessionToken(url, getMorphToken()), getLocale()) : ''),
    [url],
  );

  useEffect(() => {
    if (up !== true || !frameRef.current) return undefined;
    return bindFrameLocale(frameRef.current);
  }, [up, src]);

  useEffect(() => {
    let cancel = false;
    if (!url) {
      setUp(false);
      return undefined;
    }
    setUp(null);
    probeOrigin(url).then((ok) => {
      if (!cancel) setUp(ok);
    });
    return () => {
      cancel = true;
    };
  }, [url, tryId]);

  if (!spec) return null;

  let body = null;
  if (!url) {
    body = <Alert severity="info">{t('notConfigured', { label })}</Alert>;
  } else if (up === null) {
    body = (
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, p: 2 }}>
        <CircularProgress size={18} />
        <Typography variant="body2">{t('checking', { label })}</Typography>
      </Box>
    );
  } else if (!up) {
    const local = /localhost|127\.0\.0\.1|\[::1\]/.test(url);
    body = (
      <Box sx={{ p: 2, display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: 1 }}>
        <Typography variant="body2">
          {local
            ? t('notRunning', { label, start: spec.start })
            : t('notReachable', { label })}
        </Typography>
        <Button variant="outlined" onClick={() => setTryId((n) => n + 1)}>{t('retry')}</Button>
      </Box>
    );
  } else {
    body = (
      <iframe
        ref={frameRef}
        title={label}
        src={src}
        className="notes-embed-frame"
        onLoad={() => postLocaleToFrame(frameRef.current, getLocale())}
      />
    );
  }

  return (
    <Box className="notes-embed" sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
      {body}
    </Box>
  );
}
