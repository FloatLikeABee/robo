export function grillErrorMessage(err) {
  const data = err && err.response && err.response.data;
  if (data && typeof data.error === 'string' && data.error) return data.error;
  if (err && err.message) return err.message;
  return 'AI assist failed';
}

export function createGrillHost({
  kind,
  title,
  description,
  callGrill,
  onApply,
  onReview,
  onError,
  post,
  onResize,
}) {
  let draft = '';
  return function onMessage(event) {
    const data = event && event.data;
    if (!data || data.jsonrpc !== '2.0') return undefined;
    if (data.method === 'ui/initialize' && data.id != null) {
      post({
        jsonrpc: '2.0',
        id: data.id,
        result: {
          protocolVersion: '2026-01-26',
          hostInfo: { name: 'Dorph', version: '0.3.0' },
          hostCapabilities: {},
        },
      });
      return undefined;
    }
    if (data.method === 'ui/notifications/initialized') {
      post({
        jsonrpc: '2.0',
        method: 'ui/notifications/host-context',
        params: {
          surface: 'spa',
          kind: kind || 'note',
          title: title || '',
          description: description || '',
        },
      });
      return undefined;
    }
    if (data.method === 'ui/notifications/size-changed') {
      const height = data.params && data.params.height;
      if (onResize && typeof height === 'number') onResize(height);
      return undefined;
    }
    if (data.method === 'ui/notifications/grill-apply') {
      if (draft && onApply) onApply(draft);
      return undefined;
    }
    if (data.method !== 'tools/call' || data.id == null) return undefined;
    const name = data.params && data.params.name;
    const args = (data.params && data.params.arguments) || {};
    if (name !== 'note_grill') {
      post({
        jsonrpc: '2.0',
        id: data.id,
        result: { isError: true, content: [{ type: 'text', text: 'unsupported tool' }] },
      });
      return undefined;
    }
    return Promise.resolve()
      .then(() => callGrill(args))
      .then((view) => {
        if (view && view.phase === 'review' && typeof view.draft === 'string' && view.draft) {
          draft = view.draft;
          if (onReview) onReview(view.draft);
        }
        post({
          jsonrpc: '2.0',
          id: data.id,
          result: {
            structuredContent: view,
            content: [{ type: 'text', text: (view && (view.draft || view.question)) || '' }],
          },
        });
      })
      .catch((err) => {
        const message = grillErrorMessage(err);
        if (onError) onError(message);
        post({
          jsonrpc: '2.0',
          id: data.id,
          result: { isError: true, content: [{ type: 'text', text: message }] },
        });
      });
  };
}
