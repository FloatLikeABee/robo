/** Shared labels for question type dropdowns (forms). */
export const QUESTION_TYPES = [
  { value: 'text', label: 'Text' },
  { value: 'select', label: 'Single select' },
  { value: 'multiselect', label: 'Multi-select' },
  { value: 'boolean', label: 'Boolean' },
  { value: 'document', label: 'File upload' },
  { value: 'image', label: 'Image' },
] as const;

/** Legacy types still rendered on existing forms; not offered in the picker. */
export const LEGACY_QUESTION_TYPES = [
  { value: 'integer', label: 'Integer' },
  { value: 'float', label: 'Float' },
  { value: 'date', label: 'Date' },
  { value: 'datetime', label: 'Date & Time' },
  { value: 'qrcode', label: 'QR Code' },
] as const;
