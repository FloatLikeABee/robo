import { useEffect, useRef } from 'react';
import { drawPixelGrid, parsePixelGrid } from './pixelGrid';
import { enlargeBind, useVisualLightbox } from './visualLightbox';

export default function PixelGridBlock({ source }) {
  const canvasRef = useRef(null);
  const grid = parsePixelGrid(source);
  const open = useVisualLightbox();

  useEffect(() => {
    if (grid && canvasRef.current) drawPixelGrid(canvasRef.current, grid);
  }, [source, grid]);

  if (!grid) {
    return <pre className="chat-pixel-fallback">{String(source || '').trim()}</pre>;
  }
  const bind = enlargeBind(open, () => {
    const src = canvasRef.current?.toDataURL?.();
    return src ? { src, alt: 'Pixel art' } : null;
  });
  return (
    <canvas
      ref={canvasRef}
      className={open ? 'chat-pixel-grid chat-visual-enlarge' : 'chat-pixel-grid'}
      style={{ imageRendering: 'pixelated', width: grid.width * 8, height: grid.height * 8 }}
      aria-label={open ? 'Enlarge pixel art' : 'Pixel art'}
      {...bind}
    />
  );
}
