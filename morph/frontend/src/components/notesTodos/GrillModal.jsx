import React, { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, Typography } from '@mui/material';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { createGrillHost, grillErrorMessage } from './grillHost';

const font = 'Arial, "Helvetica Neue", Helvetica, sans-serif';

export default function GrillModal({
  kind,
  title,
  description,
  onClose,
  onApply,
  createHost = createGrillHost,
}) {
  const frameRef = useRef(null);
  const onApplyRef = useRef(onApply);
  onApplyRef.current = onApply;
  const [html, setHtml] = useState('');
  const [draft, setDraft] = useState('');
  const [error, setError] = useState('');
  const [frameHeight, setFrameHeight] = useState(420);

  useEffect(() => {
    let dead = false;
    tranApi
      .get(tranEndpoints.grillUI, {
        responseType: 'text',
        transformResponse: [(data) => data],
      })
      .then((res) => {
        const text = typeof res.data === 'string' ? res.data : '';
        if (dead) return;
        if (!text.includes('<html')) setError('AI assist page did not load');
        else setHtml(text);
      })
      .catch((err) => {
        if (!dead) setError(grillErrorMessage(err));
      });
    return () => {
      dead = true;
    };
  }, []);

  useLayoutEffect(() => {
    if (!html) return undefined;
    const frame = frameRef.current;
    if (!frame) return undefined;
    const host = createHost({
      kind,
      title,
      description,
      callGrill: async (args) => {
        const { data } = await tranApi.post(tranEndpoints.grill, {
          action: args.action,
          id: args.id,
          kind: args.kind || kind,
          title: args.title || title,
          description: args.description != null ? args.description : description,
          answer: args.answer,
        });
        return data;
      },
      onApply: (text) => {
        if (onApplyRef.current) onApplyRef.current(text);
      },
      onReview: (text) => setDraft(text),
      onError: (message) => setError(message),
      onResize: (height) => {
        if (height > 80 && height < 900) setFrameHeight(height);
      },
      post: (msg) => {
        if (frame.contentWindow) frame.contentWindow.postMessage(msg, '*');
      },
    });
    function onMessage(event) {
      if (event.source !== frame.contentWindow) return;
      host(event);
    }
    window.addEventListener('message', onMessage);
    // Assign after the listener. srcdoc scripts run as the document loads.
    frame.srcdoc = html;
    return () => {
      window.removeEventListener('message', onMessage);
    };
  }, [html, kind, title, description, createHost]);

  return (
    <Dialog
      open
      onClose={onClose}
      fullWidth
      maxWidth="sm"
      aria-labelledby="grill-title"
      PaperProps={{
        sx: {
          bgcolor: '#070b14',
          color: '#e8eef7',
          fontFamily: font,
          backgroundImage: 'none',
        },
      }}
    >
      <DialogTitle id="grill-title" sx={{ fontFamily: font, color: '#a78bfa' }}>
        AI assist
      </DialogTitle>
      <DialogContent sx={{ fontFamily: font }}>
        {error ? (
          <Typography color="error" variant="body2" sx={{ mb: 1 }}>
            {error}
          </Typography>
        ) : null}
        {html ? (
          <Box
            component="iframe"
            ref={frameRef}
            title="AI assist"
            sandbox="allow-scripts"
            sx={{
              width: '100%',
              height: frameHeight,
              border: 0,
              borderRadius: 1,
              bgcolor: '#070b14',
            }}
          />
        ) : (
          <Typography variant="body2" sx={{ color: '#94a3b8' }}>
            Opening AI assist…
          </Typography>
        )}
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} sx={{ color: '#38bdf8', fontFamily: font }}>
          Cancel
        </Button>
        <Button
          variant="contained"
          disabled={!draft}
          onClick={() => {
            if (draft && onApplyRef.current) onApplyRef.current(draft);
          }}
          sx={{ fontFamily: font, bgcolor: '#38bdf8', color: '#070b14' }}
        >
          Use this
        </Button>
      </DialogActions>
    </Dialog>
  );
}
