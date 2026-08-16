import React, { createContext, useCallback, useContext, useMemo, useState } from 'react';
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
} from '@mui/material';

const ConfirmContext = createContext(null);

/**
 * In-app confirm dialog (replaces window.confirm / window.alert for destructive actions).
 */
export function ConfirmProvider({ children }) {
  const [state, setState] = useState(null);

  const confirm = useCallback((opts) => {
    const options = typeof opts === 'string' ? { message: opts } : opts || {};
    return new Promise((resolve) => {
      setState({
        title: options.title || 'Confirm',
        message: options.message || 'Are you sure?',
        confirmLabel: options.confirmLabel || 'Confirm',
        cancelLabel: options.cancelLabel || 'Cancel',
        danger: Boolean(options.danger),
        resolve,
      });
    });
  }, []);

  const alert = useCallback((opts) => {
    const options = typeof opts === 'string' ? { message: opts } : opts || {};
    return new Promise((resolve) => {
      setState({
        title: options.title || 'Notice',
        message: options.message || '',
        confirmLabel: options.confirmLabel || 'OK',
        cancelLabel: null,
        danger: false,
        resolve: () => resolve(true),
      });
    });
  }, []);

  const close = (result) => {
    if (state?.resolve) state.resolve(result);
    setState(null);
  };

  const value = useMemo(() => ({ confirm, alert }), [confirm, alert]);

  return (
    <ConfirmContext.Provider value={value}>
      {children}
      <Dialog open={Boolean(state)} onClose={() => close(false)} maxWidth="xs" fullWidth>
        {state ? (
          <>
            <DialogTitle>{state.title}</DialogTitle>
            <DialogContent>
              <DialogContentText sx={{ whiteSpace: 'pre-wrap' }}>{state.message}</DialogContentText>
            </DialogContent>
            <DialogActions sx={{ px: 2, pb: 1.5 }}>
              {state.cancelLabel ? (
                <Button onClick={() => close(false)} color="inherit">
                  {state.cancelLabel}
                </Button>
              ) : null}
              <Button
                onClick={() => close(true)}
                variant="contained"
                color={state.danger ? 'error' : 'primary'}
                autoFocus
              >
                {state.confirmLabel}
              </Button>
            </DialogActions>
          </>
        ) : null}
      </Dialog>
    </ConfirmContext.Provider>
  );
}

export function useConfirm() {
  const ctx = useContext(ConfirmContext);
  if (!ctx) {
    return {
      confirm: async (opts) => {
        const message = typeof opts === 'string' ? opts : opts?.message || 'Are you sure?';
        return window.confirm(message);
      },
      alert: async (opts) => {
        const message = typeof opts === 'string' ? opts : opts?.message || '';
        window.alert(message);
        return true;
      },
    };
  }
  return ctx;
}
