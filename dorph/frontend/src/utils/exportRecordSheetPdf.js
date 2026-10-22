import html2canvas from 'html2canvas';
import { jsPDF } from 'jspdf';

/** Safe filename for user download (no path chars). */
export function sanitizeRecordPdfFilename(displayName, recordId) {
  const base =
    displayName && String(displayName).trim()
      ? String(displayName)
          .replace(/[/\\?%*:|"<>]/g, '-')
          .replace(/\s+/g, ' ')
          .trim()
          .slice(0, 120)
      : `record-${recordId != null ? recordId : 'export'}`;
  return `${base}.pdf`;
}

/**
 * Rasterize a DOM subtree to a multi-page A4 PDF (full width, preserves on-screen colors).
 */
export async function exportDomNodeToPdf(node, filename) {
  if (!node || typeof window === 'undefined') return;

  const canvas = await html2canvas(node, {
    scale: 2,
    useCORS: true,
    logging: false,
    allowTaint: true,
    width: Math.max(node.scrollWidth, node.offsetWidth),
    height: Math.max(node.scrollHeight, node.offsetHeight),
  });

  const pdf = new jsPDF({ orientation: 'portrait', unit: 'mm', format: 'a4' });
  const pageW = pdf.internal.pageSize.getWidth();
  const pageH = pdf.internal.pageSize.getHeight();
  const imgW = canvas.width;
  const imgH = canvas.height;
  if (imgW < 1 || imgH < 1) {
    pdf.save(filename);
    return;
  }

  /** Source rows that fit on one PDF page when the image is scaled to full page width. */
  const pxPerPageVertical = Math.max(1, Math.floor((pageH * imgW) / pageW));

  let y = 0;
  let first = true;
  while (y < imgH) {
    const sliceH = Math.min(pxPerPageVertical, imgH - y);
    const sliceCanvas = document.createElement('canvas');
    sliceCanvas.width = imgW;
    sliceCanvas.height = sliceH;
    const ctx = sliceCanvas.getContext('2d');
    ctx.drawImage(canvas, 0, y, imgW, sliceH, 0, 0, imgW, sliceH);
    const sliceData = sliceCanvas.toDataURL('image/png');
    const mmH = (sliceH * pageW) / imgW;

    if (!first) pdf.addPage();
    pdf.addImage(sliceData, 'PNG', 0, 0, pageW, mmH);
    first = false;
    y += sliceH;
  }

  pdf.save(filename);
}
