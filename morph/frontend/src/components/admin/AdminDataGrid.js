import React, { useMemo, useState, useEffect, useCallback, useRef } from 'react';
import {
  Box,
  Typography,
  Toolbar,
  TextField,
  IconButton,
  Select,
  MenuItem,
  Paper,
  FormControl,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  FormControlLabel,
  Checkbox,
  useTheme,
  Drawer,
  Badge,
  Tooltip,
  CircularProgress,
  Snackbar,
  Alert,
  useMediaQuery,
} from '@mui/material';
import { DataGrid } from '@mui/x-data-grid';
import SearchIcon from '@mui/icons-material/Search';
import CloseIcon from '@mui/icons-material/Close';
import FilterListIcon from '@mui/icons-material/FilterList';
import PaletteOutlinedIcon from '@mui/icons-material/PaletteOutlined';
import AddOutlinedIcon from '@mui/icons-material/AddOutlined';
import DeleteOutlineOutlinedIcon from '@mui/icons-material/DeleteOutlineOutlined';
import SaveOutlinedIcon from '@mui/icons-material/SaveOutlined';
import CommentOutlinedIcon from '@mui/icons-material/CommentOutlined';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import VisibilityOutlinedIcon from '@mui/icons-material/VisibilityOutlined';
import FileDownloadOutlinedIcon from '@mui/icons-material/FileDownloadOutlined';
import CommentsModal from './CommentsModal';
import { countCommentTree } from './commentCountUtils';
import { getAdminTableHeaderBg, getAdminTableHeaderColor } from '../../theme';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import AdvancedFilterModal from './AdvancedFilterModal';
import ColorTagsModal from './ColorTagsModal';
import { applyQuickSearchThenRules } from './gridFilterUtils';
import { normalizeColorConfig, buildRowColorTokens, buildRowColorCss } from './colorTagResolve';
import { usePlatformUi } from '../../PlatformUiContext';
import { sanitizeTextInput, sanitizeNumericInput } from '../../utils/inputValidation';
import { exportDomNodeToPdf, sanitizeRecordPdfFilename } from '../../utils/exportRecordSheetPdf';
import RecordAttachmentsPanel from './RecordAttachmentsPanel';
import { ActivityLocationFieldEditor, RecordSheetActivityLocation } from './ActivityLocationField';
import { activityLocationDescriptionOnly } from '../../utils/activityLocationJson';
import { JsonDetailEditor, RecordSheetJsonDetail, jsonDetailToString } from './jsonDetailViews';
import { validateJsonDetailStructure } from './jsonDetailValidate';
import {
  DETAIL_DRAWER_BG_DARK,
  DETAIL_PANEL_BG_DARK,
  DETAIL_CARD_SX_DARK,
} from './detailViewTheme';
import {
  ADMIN_RIGHT_PANEL_TOP,
  adminRightPanelHeightCalc,
} from './adminRightPanelStyle';

const STORAGE_PREFIX = 'morphdata-detail-order-';
const DEFAULT_HIDDEN_FIELDS = ['id', 'guid', 'attachments'];
const DEFAULT_EMPTY_LIST = [];
const DEFAULT_DETAIL_JSON_FIELD_KEYS = ['detail'];

/** Fields pinned below the draggable grid (not reorderable): long description, then JSON detail. */
const PINNED_BOTTOM_KEYS = new Set(['detail', 'description']);
function isPinnedBottomField(field) {
  return PINNED_BOTTOM_KEYS.has(String(field || '').toLowerCase());
}

/** Format a field value for the printable sheet (plain text / pretty JSON). */
function formatPrintFieldBody(field, raw, detailJsonFieldKeys, locationJsonFieldKeys) {
  if (raw === null || raw === undefined || raw === '') return '—';
  const name = String(field).toLowerCase();
  if (
    locationJsonFieldKeys &&
    locationJsonFieldKeys.length > 0 &&
    locationJsonFieldKeys.some((k) => String(k).toLowerCase() === name)
  ) {
    const d = activityLocationDescriptionOnly(raw);
    return d || '—';
  }
  const isJsonKey =
    detailJsonFieldKeys &&
    detailJsonFieldKeys.length > 0 &&
    detailJsonFieldKeys.some((k) => String(k).toLowerCase() === name);
  if (isJsonKey && typeof raw === 'string') {
    const t = raw.trim();
    if (t) {
      try {
        return JSON.stringify(JSON.parse(t), null, 2);
      } catch {
        return raw;
      }
    }
    return '—';
  }
  if (typeof raw === 'object' && raw !== null && !(raw instanceof Date)) {
    try {
      return JSON.stringify(raw, null, 2);
    } catch {
      return String(raw);
    }
  }
  if (typeof raw === 'boolean') return raw ? 'Yes' : 'No';
  return String(raw);
}
/** Prior named-layout blob; migrated once into {@link STORAGE_PREFIX} when absent. */
const LEGACY_NAMED_LAYOUTS_KEY = `${STORAGE_PREFIX}layouts-store-`;

function migrateLegacyNamedLayoutsToLastOrder(storageKey, keys) {
  try {
    const legacyRaw = localStorage.getItem(LEGACY_NAMED_LAYOUTS_KEY + storageKey);
    if (!legacyRaw) return;
    const p = JSON.parse(legacyRaw);
    const layouts = p.layouts && typeof p.layouts === 'object' ? p.layouts : {};
    const active = p.activeLayoutName;
    let saved =
      active && Array.isArray(layouts[active]) ? layouts[active] : null;
    if (!saved) {
      const firstName = Object.keys(layouts).sort()[0];
      if (firstName) saved = layouts[firstName];
    }
    if (!saved || !Array.isArray(saved)) {
      localStorage.removeItem(LEGACY_NAMED_LAYOUTS_KEY + storageKey);
      return;
    }
    const ordered = saved.filter((k) => keys.includes(k));
    const rest = keys.filter((k) => !ordered.includes(k));
    try {
      localStorage.setItem(STORAGE_PREFIX + storageKey, JSON.stringify([...ordered, ...rest]));
    } catch {
      /* ignore quota */
      return;
    }
    localStorage.removeItem(LEGACY_NAMED_LAYOUTS_KEY + storageKey);
  } catch {
    /* ignore */
  }
}

function isEmptyJsonDetailValue(v) {
  if (v == null) return true;
  if (typeof v === 'string' && v.trim() === '') return true;
  return false;
}

/** Split all-lowercase glued keys from SQL maps (e.g. licensenumber → license number). */
function splitGluedCompound(word) {
  const w = word.toLowerCase();
  const pairs = [
    ['license', 'number'],
    ['license', 'state'],
    ['license', 'expiration'],
    ['home', 'phone'],
    ['work', 'phone'],
    ['cell', 'phone'],
    ['last', 'name'],
    ['first', 'name'],
    ['middle', 'name'],
    ['active', 'flag'],
    ['last', 'updated'],
    ['deleted', 'flag'],
    ['deleted', 'date'],
    ['mail', 'street1'],
    ['mail', 'street2'],
    ['mail', 'city'],
    ['mail', 'zip'],
    ['entry', 'date'],
    ['load', 'time'],
    ['vehicle', 'detail'],
    ['vehicle', 'id'],
    ['trip', 'id'],
    ['staff', 'id'],
    ['stud', 'id'],
    ['student', 'id'],
    ['district', 'id'],
    ['dis', 'code'],
    ['ethnic', 'code'],
    ['feed', 'to'],
    ['school', 'code'],
    ['display', 'name'],
  ];
  for (const [a, b] of pairs) {
    if (w === a + b) return `${a} ${b}`;
  }
  const suffixes = [
    'number',
    'expiration',
    'updated',
    'phone',
    'name',
    'flag',
    'date',
    'time',
    'city',
    'zip',
    'state',
    'street1',
    'street2',
    'detail',
    'code',
    'id',
  ];
  for (const suf of suffixes) {
    if (w.endsWith(suf) && w.length > suf.length) {
      const pre = w.slice(0, -suf.length);
      const minPre = suf === 'id' ? 3 : 2;
      if (pre.length >= minPre) return `${pre} ${suf}`;
    }
  }
  return w;
}

/** Human-readable labels: snake_case, camelCase, glued lowercase (e.g. licensenumber → License Number). */
export function formatFieldLabel(field) {
  let s = String(field).trim();
  if (!s) return '';
  s = s.replace(/_/g, ' ');
  s = s.replace(/([a-z0-9])([A-Z])/g, '$1 $2');
  s = s.replace(/([A-Z])([A-Z][a-z])/g, '$1 $2');
  // Do not use bare "on" — it breaks words ending in ...expiration, ...ation, etc.
  s = s.replace(/([a-z])(detail|time|date|by|flag|id)$/i, '$1 $2');
  s = s.replace(/\s+/g, ' ').trim();
  if (!/\s/.test(s) && /^[a-z0-9]+$/i.test(s)) {
    s = splitGluedCompound(s);
  }
  s = s.replace(/\s+/g, ' ').trim();
  s = s.replace(/\b\w/g, (c) => c.toUpperCase());
  s = s.replace(/\bId\b/g, 'ID');
  return s;
}

/** Title for detail drawer / modals: prefers name, then person/school/vehicle/trip fallbacks. */
export function getRecordDisplayName(row, tableTitle, recordId) {
  if (!row || typeof row !== 'object') {
    if (tableTitle && recordId != null) return `${tableTitle} #${recordId}`;
    return tableTitle || 'Record';
  }
  const pick = (...keys) => {
    for (const k of keys) {
      const v = row[k];
      if (v !== undefined && v !== null && String(v).trim() !== '') return String(v).trim();
    }
    return '';
  };
  const n = pick('name', 'Name');
  if (n) return n;
  const ln = pick('last_name', 'Last_Name');
  const fn = pick('first_name', 'First_Name');
  if (ln || fn) return [ln, fn].filter(Boolean).join(', ');
  const dist = pick('district', 'District');
  if (dist) return dist;
  const sc = pick('facility_code', 'school_code', 'SchoolCode');
  if (sc) return sc;
  const fac = pick('facility', 'school');
  if (fac) return fac;
  const vid = row.vehicle_id ?? row.VehicleID ?? row.id;
  if (vid !== undefined && vid !== null && String(vid).trim() !== '') return `Vehicle ${vid}`;
  const tid = row.trip_id ?? row.TripID;
  if (tid !== undefined && tid !== null && String(tid).trim() !== '') return `Trip ${tid}`;
  const code = pick('code', 'Code');
  if (code) return code;
  const display = pick('display_name', 'DisplayName');
  if (display) return display;
  if (recordId != null) return `${tableTitle || 'Record'} #${recordId}`;
  return tableTitle || 'Record';
}

function getFieldOrder(storageKey, keys) {
  try {
    let raw = localStorage.getItem(STORAGE_PREFIX + storageKey);
    if (!raw) {
      migrateLegacyNamedLayoutsToLastOrder(storageKey, keys);
      raw = localStorage.getItem(STORAGE_PREFIX + storageKey);
    }
    if (!raw) return [...keys];
    const saved = JSON.parse(raw);
    if (!Array.isArray(saved)) return [...keys];
    const ordered = saved.filter((k) => keys.includes(k));
    const rest = keys.filter((k) => !ordered.includes(k));
    return [...ordered, ...rest];
  } catch {
    return [...keys];
  }
}

/** DB/API uses 1=M, 2=F (see seed_tran); MUI Select uses "", "M", "F". */
function genderApiToSelect(value) {
  if (value === null || value === undefined || value === '') return '';
  if (value === 0 || value === '0') return '';
  if (value === 1 || value === '1') return 'M';
  if (value === 2 || value === '2') return 'F';
  const s = String(value).trim();
  if (!s) return '';
  const u = s.toUpperCase();
  if (u === 'M' || u === 'MALE') return 'M';
  if (u === 'F' || u === 'FEMALE') return 'F';
  return '';
}

/** Persist as tinyint expected by handlers (intFromAny). */
function genderSelectToApi(value) {
  if (value === null || value === undefined || value === '') return null;
  if (value === 'M') return 1;
  if (value === 'F') return 2;
  if (value === 1 || value === 2) return value;
  if (value === '1') return 1;
  if (value === '2') return 2;
  return null;
}

function isDateOnlyFieldName(field) {
  const name = String(field).toLowerCase();
  if (name === 'dob') return true;
  if (
    (name.includes('date') && name.includes('time')) ||
    name.includes('datetime') ||
    name.includes('timestamp')
  ) {
    return false;
  }
  return name.includes('date') || name.endsWith('_on');
}

/** Build edit/create draft from a loaded detail row (shared by sync effects). */
function buildDraftFromDetailRow(
  detailRow,
  { detailJsonFieldKeys, locationJsonFieldKeys, isHiddenKey }
) {
  const filtered = Object.fromEntries(
    Object.entries(detailRow).filter(([k]) => !isHiddenKey(k))
  );
  for (const k of detailJsonFieldKeys) {
    const kk = String(k).toLowerCase();
    if (filtered[kk] == null) filtered[kk] = '{}';
    else if (typeof filtered[kk] === 'object' && filtered[kk] !== null && !(filtered[kk] instanceof Date)) {
      filtered[kk] = jsonDetailToString(filtered[kk]) || '{}';
    }
  }
  for (const k of locationJsonFieldKeys) {
    const kk = String(k).toLowerCase();
    const v = filtered[kk];
    if (v == null || v === '') {
      filtered[kk] = null;
    } else if (typeof v === 'object' && !Array.isArray(v)) {
      try {
        filtered[kk] = JSON.stringify(v);
      } catch {
        filtered[kk] = null;
      }
    }
  }
  if (Object.prototype.hasOwnProperty.call(filtered, 'gender')) {
    filtered.gender = genderApiToSelect(filtered.gender);
  }
  for (const k of Object.keys(filtered)) {
    if (isDateOnlyFieldName(k)) {
      const n = normalizeInputValue(k, filtered[k], 'date');
      filtered[k] = n === '' ? null : n;
    }
  }
  return filtered;
}

function serializeLocationJsonPayload(payload, locationJsonFieldKeys) {
  for (const k of locationJsonFieldKeys) {
    const kk = String(k).toLowerCase();
    if (!Object.prototype.hasOwnProperty.call(payload, kk)) continue;
    const v = payload[kk];
    if (v == null || v === '') {
      payload[kk] = null;
      continue;
    }
    if (typeof v === 'object' && !Array.isArray(v)) {
      try {
        payload[kk] = JSON.stringify(v);
      } catch {
        payload[kk] = null;
      }
    }
  }
}

function normalizeInputValue(field, value, controlType) {
  if (value == null || value === '') return '';
  if (controlType === 'boolean') return Boolean(value);
  if (controlType === 'date') {
    if (value instanceof Date && !Number.isNaN(value.getTime())) {
      return value.toISOString().slice(0, 10);
    }
    if (typeof value === 'string') {
      const s = value.trim();
      if (s.length >= 10 && /^\d{4}-\d{2}-\d{2}/.test(s)) return s.slice(0, 10);
    }
    return '';
  }
  if (controlType === 'datetime' && typeof value === 'string') return value.slice(0, 16);
  if (controlType === 'time' && typeof value === 'string') return value.length >= 5 ? value.slice(0, 8) : value;
  if (controlType === 'gender') return genderApiToSelect(value);
  return value;
}

/** Map API keys (often lowercased PascalCase) to snake_case used by forms and backend. */
function canonicalizeDetailKeys(row) {
  if (!row || typeof row !== 'object') return row;
  const ALIASES = {
    lastname: 'last_name',
    firstname: 'first_name',
    middlename: 'middle_name',
    entrydate: 'entry_date',
    school_code: 'facility_code',
    school: 'facility',
    cell_phone: 'phone_number',
    modelinfo: 'model_info',
    tripdays: 'trip_days',
    comments: 'note',
  };
  const out = { ...row };
  for (const [from, to] of Object.entries(ALIASES)) {
    if (Object.prototype.hasOwnProperty.call(out, from) && out[to] === undefined) {
      out[to] = out[from];
      delete out[from];
    }
  }
  return out;
}

function numberOptionsForFieldName(field) {
  const name = String(field || '').toLowerCase();
  const allowDecimal =
    name.includes('distance') ||
    name.includes('rate') ||
    name.includes('amount') ||
    name.includes('price') ||
    name.includes('cost') ||
    name.includes('percent') ||
    name.includes('load_time');
  return { allowDecimal, allowNegative: false };
}

/**
 * AdminDataGrid (free @mui/x-data-grid)
 * - Full-width table; row click opens a right-side Drawer (modal) with detail fields
 * - Detail shows all fields; drag to reorder; order persisted per table
 */
export default function AdminDataGrid({
  title,
  rows,
  columns,
  loading,
  rowHeight = 40,
  getRowClassName,
  requiredFields = DEFAULT_EMPTY_LIST,
  hiddenFields = DEFAULT_HIDDEN_FIELDS,
  fetchDetail, // optional async (id) => fullRowObject
  createFields = DEFAULT_EMPTY_LIST, // optional array of field names for create modal
  onCreate,
  onUpdate, // optional (id, payload) => Promise<void> for saving detail edits
  onDelete,
  entityTypeForComments, // optional 'student' | 'staff' | 'school' | 'vehicle' | 'trip' to show Comments button
  /** Field keys (lowercase) that use vehicle-style JSON table + Edit JSON. Default: detail only. */
  detailJsonFieldKeys = DEFAULT_DETAIL_JSON_FIELD_KEYS,
  /** Optional (row, field) => value for advanced filter when column is computed */
  filterGetValue = undefined,
  /** When true, the `facility` field uses a facility code selector (from /api/tran/facilities). */
  facilityAsSelect = false,
  /** When true, `facility_id` is a numeric FK selector (facility primary keys from /api/tran/facilities). */
  facilityIdAsSelect = false,
  /** When true, `district_id` is a numeric FK selector (district primary keys from /api/tran/districts). */
  districtIdAsSelect = false,
  /** When true, the `detail` field is pinned at the bottom (not draggable; large JSON editor). */
  pinNoteSection = true,
  /** Field keys (lowercase) ending in `_id` that stay visible in grid/detail/print; others like `district_id` are hidden by default. */
  underscoreIdFieldsVisible = DEFAULT_EMPTY_LIST,
  /** Field keys that use map-backed JSON editing (OpenStreetMap / Leaflet) instead of a plain text control. */
  locationJsonFieldKeys = DEFAULT_EMPTY_LIST,
  /** API route segment for file attachments, e.g. facilities, members, employees, assets, activities, case-tasks */
  entityRouteForAttachments = null,
}) {
  const { dictionaries: entityDictionaries, labels: platformLabels } = usePlatformUi();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const isPhone = useMediaQuery(theme.breakpoints.down('sm'));
  const isCompact = useMediaQuery(theme.breakpoints.down('md'));
  const headerAccentBg = getAdminTableHeaderBg(theme.palette.mode);
  const headerAccentColor = getAdminTableHeaderColor(theme.palette.mode);
  const storageKey = useMemo(() => (title || 'table').replace(/\s+/g, '-').toLowerCase(), [title]);
  const detailJsonKeysSignature = useMemo(
    () =>
      (Array.isArray(detailJsonFieldKeys) ? detailJsonFieldKeys : [])
        .map((k) => String(k || '').trim().toLowerCase())
        .filter(Boolean)
        .sort()
        .join('|'),
    [detailJsonFieldKeys]
  );
  const locationJsonKeysSignature = useMemo(
    () =>
      (Array.isArray(locationJsonFieldKeys) ? locationJsonFieldKeys : [])
        .map((k) => String(k || '').trim().toLowerCase())
        .filter(Boolean)
        .sort()
        .join('|'),
    [locationJsonFieldKeys]
  );
  const normalizedDetailJsonFieldKeys = useMemo(
    () => (detailJsonKeysSignature ? detailJsonKeysSignature.split('|') : []),
    [detailJsonKeysSignature]
  );
  const normalizedLocationJsonFieldKeys = useMemo(
    () => (locationJsonKeysSignature ? locationJsonKeysSignature.split('|') : []),
    [locationJsonKeysSignature]
  );
  const [searchText, setSearchText] = useState('');
  const [filterRules, setFilterRules] = useState([]);
  const [filterModalOpen, setFilterModalOpen] = useState(false);
  const [colorConfig, setColorConfig] = useState(() => normalizeColorConfig(null));
  const [colorTagsOpen, setColorTagsOpen] = useState(false);
  const [pageSize, setPageSize] = useState(50);
  const [paginationPage, setPaginationPage] = useState(0);
  const [expandedId, setExpandedId] = useState(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [createDraft, setCreateDraft] = useState({});
  const [createErr, setCreateErr] = useState(null);
  const [createBusy, setCreateBusy] = useState(false);
  const [deleteBusy, setDeleteBusy] = useState(false);
  const [deleteErr, setDeleteErr] = useState(null);
  const [detailRow, setDetailRow] = useState(null);
  const [detailBusy, setDetailBusy] = useState(false);
  /** Bumped when server detail should replace editDraft (row open / refresh). Not bumped on local edits. */
  const [detailDraftSeed, setDetailDraftSeed] = useState(0);
  const [editDraft, setEditDraft] = useState({});
  const [saveBusy, setSaveBusy] = useState(false);
  const [saveErr, setSaveErr] = useState(null);
  const [saveSuccessOpen, setSaveSuccessOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [batchDeleteConfirmOpen, setBatchDeleteConfirmOpen] = useState(false);
  const [rowSelectionModel, setRowSelectionModel] = useState([]);
  const [saveConfirmOpen, setSaveConfirmOpen] = useState(false);
  const [commentsOpen, setCommentsOpen] = useState(false);
  const [threadCommentCount, setThreadCommentCount] = useState(0);
  const [detailDrawerOpen, setDetailDrawerOpen] = useState(false);
  /** Per-field: 'preview' | 'tree' | 'raw' for json_detail fields */
  const [jsonDetailEditorMode, setJsonDetailEditorMode] = useState({});
  /** Current detail card order for the open record; persists to localStorage on drag. */
  const [detailFieldOrder, setDetailFieldOrder] = useState([]);
  const [dragOverDetailIndex, setDragOverDetailIndex] = useState(null);
  /** { code: facility_code, label: display } for facility selector */
  const [facilityOptions, setFacilityOptions] = useState([]);
  /** { id: number, label: display } for facility_id FK selector */
  const [facilityIdOptions, setFacilityIdOptions] = useState([]);
  /** { id: number, label: display } for district_id FK selector */
  const [districtIdOptions, setDistrictIdOptions] = useState([]);

  useEffect(() => {
    if (!facilityAsSelect && !facilityIdAsSelect) {
      setFacilityOptions([]);
      setFacilityIdOptions([]);
      return;
    }
    let alive = true;
    tranApi
      .get(tranEndpoints.facilities)
      .then((res) => {
        if (!alive) return;
        const list = Array.isArray(res.data) ? res.data : [];
        const rows = list
          .map((s) => {
            const idRaw = s.id ?? s.ID;
            const id = Number(idRaw);
            const code = s.facility_code ?? s.FacilityCode ?? '';
            const name = s.name ?? s.Name ?? '';
            const codeStr = String(code).trim();
            const nameStr = String(name).trim();
            const placeNoun = platformLabels?.term_facility || 'Place';
            const label =
              codeStr && nameStr
                ? `${codeStr} — ${nameStr}`
                : codeStr || nameStr || (Number.isFinite(id) ? `${placeNoun} #${id}` : '');
            return { id, code: codeStr, label: label || String(idRaw ?? '') };
          })
          .filter((r) => Number.isFinite(r.id));
        if (facilityAsSelect) {
          setFacilityOptions(rows.filter((r) => r.code).map((r) => ({ code: r.code, label: r.label })));
        } else {
          setFacilityOptions([]);
        }
        if (facilityIdAsSelect) {
          const placeNoun = platformLabels?.term_facility || 'Place';
          setFacilityIdOptions(rows.map((r) => ({ id: r.id, label: r.label || `${placeNoun} #${r.id}` })));
        } else {
          setFacilityIdOptions([]);
        }
      })
      .catch(() => {
        if (!alive) return;
        if (facilityAsSelect) setFacilityOptions([]);
        if (facilityIdAsSelect) setFacilityIdOptions([]);
      });
    return () => {
      alive = false;
    };
  }, [facilityAsSelect, facilityIdAsSelect]);

  useEffect(() => {
    if (!districtIdAsSelect) {
      setDistrictIdOptions([]);
      return;
    }
    let alive = true;
    tranApi
      .get(tranEndpoints.districts)
      .then((res) => {
        if (!alive) return;
        const list = Array.isArray(res.data) ? res.data : [];
        const rows = list
          .map((d) => {
            const idRaw = d.id ?? d.ID;
            const id = Number(idRaw);
            const code = String(d.district ?? d.District ?? '').trim();
            const name = String(d.name ?? d.Name ?? '').trim();
            const label =
              code && name ? `${code} — ${name}` : code || name || (Number.isFinite(id) ? `District #${id}` : '');
            return { id, label: label || String(idRaw ?? '') };
          })
          .filter((r) => Number.isFinite(r.id));
        setDistrictIdOptions(rows);
      })
      .catch(() => {
        if (!alive) return;
        setDistrictIdOptions([]);
      });
    return () => {
      alive = false;
    };
  }, [districtIdAsSelect]);

  useEffect(() => {
    let alive = true;
    tranApi
      .get(`${tranEndpoints.gridColorConfig}?grid_key=${encodeURIComponent(storageKey)}`)
      .then((res) => {
        if (!alive) return;
        setColorConfig(normalizeColorConfig(res.data));
      })
      .catch(() => {
        if (!alive) return;
        setColorConfig(normalizeColorConfig(null));
      });
    return () => {
      alive = false;
    };
  }, [storageKey]);

  const activeFilterCount = useMemo(
    () => (filterRules || []).filter((r) => r && r.field).length,
    [filterRules]
  );

  const filteredRows = useMemo(() => {
    return applyQuickSearchThenRules(rows, searchText, filterRules, filterGetValue);
  }, [rows, searchText, filterRules, filterGetValue]);

  useEffect(() => {
    setPaginationPage(0);
  }, [searchText, filterRules]);

  useEffect(() => {
    const maxPage = Math.max(0, Math.ceil(filteredRows.length / pageSize) - 1);
    setPaginationPage((p) => Math.min(p, maxPage));
  }, [filteredRows.length, pageSize]);

  const colorRowMeta = useMemo(
    () => buildRowColorTokens(filteredRows, colorConfig, filterGetValue),
    [filteredRows, colorConfig, filterGetValue]
  );

  const colorRowCss = useMemo(
    () => buildRowColorCss(colorRowMeta.uniqueColors, isDark),
    [colorRowMeta.uniqueColors, isDark]
  );

  // Keep selection only if still visible; do not auto-select first row (detail opens from row click)
  useEffect(() => {
    if (!filteredRows.length) {
      setExpandedId(null);
      return;
    }
    setExpandedId((prev) => {
      const stillVisible = filteredRows.some((r) => r.id === prev);
      if (prev != null && stillVisible) return prev;
      return null;
    });
  }, [filteredRows]);

  useEffect(() => {
    if (expandedId == null) setDetailDrawerOpen(false);
  }, [expandedId]);

  useEffect(() => {
    setJsonDetailEditorMode({});
  }, [expandedId]);

  useEffect(() => {
    setRowSelectionModel((prev) => prev.filter((id) => filteredRows.some((r) => r.id === id)));
  }, [filteredRows]);

  const selectedRow = useMemo(
    () => (expandedId != null ? filteredRows.find((r) => r.id === expandedId) : null),
    [filteredRows, expandedId]
  );

  const deleteTargets = useMemo(() => {
    if (rowSelectionModel.length > 0) {
      const idSet = new Set(rowSelectionModel);
      return filteredRows.filter((r) => idSet.has(r.id));
    }
    if (selectedRow) return [selectedRow];
    return [];
  }, [rowSelectionModel, filteredRows, selectedRow]);

  const detailPanelTitle = useMemo(
    () => getRecordDisplayName(detailRow || selectedRow, title, expandedId),
    [detailRow, selectedRow, title, expandedId]
  );

  const isHiddenKey = useCallback((key) => {
    const lk = String(key).toLowerCase();
    const hiddenLower = hiddenFields.map((x) => String(x).toLowerCase());
    if (hiddenLower.includes(lk)) return true;
    // Attachments are edited in the dedicated Documents&media section, not as a generic field card.
    if (lk === 'attachments') return true;
    if (lk === 'id' || lk === 'db_id' || lk === 'dbid') return true;
    if (lk === 'created_on' || lk === 'created_by' || lk === 'createdon' || lk === 'createdby') return true;
    // Hide relational/technical *_id fields unless whitelisted (e.g. district_id for facilities).
    const showUnderscoreIds = new Set(
      (underscoreIdFieldsVisible || []).map((x) => String(x).toLowerCase())
    );
    if (lk.endsWith('_id') && !showUnderscoreIds.has(lk)) return true;
    // Hide most *id suffixes (foreign keys + technical ids).
    if (lk.endsWith('id') && !showUnderscoreIds.has(lk)) return true;
    return false;
  }, [hiddenFields, underscoreIdFieldsVisible]);

  const baseGridColumns = useMemo(() => {
    const cols = columns || [];
    // Hide any columns that look like technical/relational ids
    return cols
      .filter((c) => {
        const f = c.field;
        if (!f) return true;
        return !isHiddenKey(f);
      })
      .map((c) => ({
        ...c,
        headerName: c.headerName != null && c.headerName !== '' ? c.headerName : formatFieldLabel(c.field),
      }));
  }, [columns, hiddenFields, underscoreIdFieldsVisible]); // eslint-disable-line react-hooks/exhaustive-deps -- isHiddenKey uses hiddenFields + underscoreIdFieldsVisible

  const [printModalOpen, setPrintModalOpen] = useState(false);
  const [printRecordId, setPrintRecordId] = useState(null);
  const [printRow, setPrintRow] = useState(null);
  const [printBusy, setPrintBusy] = useState(false);
  const [sheetExportBusy, setSheetExportBusy] = useState(false);
  const recordSheetPdfRef = useRef(null);
  const fetchDetailRef = useRef(fetchDetail);

  useEffect(() => {
    fetchDetailRef.current = fetchDetail;
  }, [fetchDetail]);

  const openPrintSheet = useCallback(
    async (gridRow) => {
      const id = gridRow?.id;
      if (id == null) return;
      setPrintModalOpen(true);
      setPrintBusy(true);
      setPrintRecordId(id);
      setPrintRow(null);
      try {
        let merged = canonicalizeDetailKeys({ ...gridRow });
        if (fetchDetail) {
          try {
            const full = await fetchDetail(id);
            const lowered = Object.fromEntries(
              Object.entries(full || {}).map(([k, v]) => [String(k).toLowerCase(), v])
            );
            merged = canonicalizeDetailKeys(lowered);
          } catch {
            merged = canonicalizeDetailKeys({ ...gridRow });
          }
        }
        setPrintRow(merged);
      } finally {
        setPrintBusy(false);
      }
    },
    [fetchDetail]
  );

  const closePrintSheet = useCallback(() => {
    setPrintModalOpen(false);
    setPrintRecordId(null);
    setPrintRow(null);
    setPrintBusy(false);
    setSheetExportBusy(false);
  }, []);

  const exportRecordSheetPdf = useCallback(async () => {
    const el = recordSheetPdfRef.current;
    if (!el || sheetExportBusy || printBusy || !printRow) return;
    setSheetExportBusy(true);
    try {
      const display = getRecordDisplayName(printRow, title, printRecordId);
      await exportDomNodeToPdf(el, sanitizeRecordPdfFilename(display, printRecordId));
    } catch (err) {
      console.error('Record PDF export failed', err);
    } finally {
      setSheetExportBusy(false);
    }
  }, [sheetExportBusy, printBusy, printRow, title, printRecordId]);

  const gridColumns = useMemo(() => {
    const viewCol = {
      field: '__print_view',
      headerName: '',
      width: 56,
      minWidth: 56,
      maxWidth: 56,
      sortable: false,
      filterable: false,
      hideable: false,
      disableColumnMenu: true,
      align: 'center',
      renderCell: (params) => (
        <IconButton
          size="small"
          onClick={(e) => {
            e.stopPropagation();
            e.preventDefault();
            void openPrintSheet(params.row);
          }}
          aria-label="Open record sheet"
          title="Open record sheet"
        >
          <VisibilityOutlinedIcon fontSize="small" />
        </IconButton>
      ),
    };
    return [viewCol, ...baseGridColumns];
  }, [baseGridColumns, openPrintSheet]);

  const printFieldKeys = useMemo(() => {
    if (!printRow || typeof printRow !== 'object') return [];
    const colFields = (columns || [])
      .map((c) => c.field)
      .filter((f) => f && !String(f).startsWith('__'));
    const rowKeys = Object.keys(printRow).filter((k) => !isHiddenKey(k));
    const seen = new Set();
    const out = [];
    for (const f of colFields) {
      if (Object.prototype.hasOwnProperty.call(printRow, f) && !isHiddenKey(f)) {
        out.push(f);
        seen.add(f);
      }
    }
    for (const k of [...rowKeys].sort()) {
      if (!seen.has(k)) out.push(k);
    }
    return out;
  // eslint-disable-next-line react-hooks/exhaustive-deps -- isHiddenKey closes over hiddenFields + underscoreIdFieldsVisible
  }, [printRow, columns, hiddenFields, underscoreIdFieldsVisible]);

  // Fetch full detail row for the right panel (e.g. Students: show all nullable content columns)
  useEffect(() => {
    let alive = true;
    async function run() {
      if (expandedId == null) {
        setDetailRow(null);
        setDetailBusy(false);
        setEditDraft({});
        return;
      }
      const listRow = selectedRow ?? null;
      setDetailRow(listRow);
      if (listRow) {
        setEditDraft(
          buildDraftFromDetailRow(canonicalizeDetailKeys({ ...listRow }), {
            detailJsonFieldKeys: normalizedDetailJsonFieldKeys,
            locationJsonFieldKeys: normalizedLocationJsonFieldKeys,
            isHiddenKey,
          })
        );
      } else {
        setEditDraft({});
      }
      const fetchDetailFn = fetchDetailRef.current;
      if (!fetchDetailFn) {
        setDetailBusy(false);
        setDetailDraftSeed((s) => s + 1);
        return;
      }
      setDetailBusy(true);
      try {
        const full = await fetchDetailFn(expandedId);
        if (!alive) return;
        const lowered = Object.fromEntries(
          Object.entries(full || {}).map(([k, v]) => [String(k).toLowerCase(), v])
        );
        setDetailRow(canonicalizeDetailKeys(lowered));
      } catch {
        if (!alive) return;
        const row = listRow ? { ...listRow } : {};
        for (const k of normalizedDetailJsonFieldKeys) {
          const key = String(k).toLowerCase();
          if (row[key] == null) row[key] = '{}';
        }
        setDetailRow(canonicalizeDetailKeys(row));
      } finally {
        if (!alive) return;
        setDetailBusy(false);
        setDetailDraftSeed((s) => s + 1);
      }
    }
    run();
    return () => {
      alive = false;
    };
  }, [
    expandedId,
    selectedRow,
    normalizedDetailJsonFieldKeys,
    normalizedLocationJsonFieldKeys,
    isHiddenKey,
  ]);

  // Seed editDraft from server detail only when detailDraftSeed bumps (not on every local edit).
  useEffect(() => {
    if (!detailRow || typeof detailRow !== 'object' || detailBusy) {
      if (!detailRow) setEditDraft({});
      return;
    }
    setEditDraft(
      buildDraftFromDetailRow(detailRow, {
        detailJsonFieldKeys: normalizedDetailJsonFieldKeys,
        locationJsonFieldKeys: normalizedLocationJsonFieldKeys,
        isHiddenKey,
      })
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [detailDraftSeed, detailBusy, underscoreIdFieldsVisible]);

  const handleCommentThreadChanged = useCallback((tree) => {
    setThreadCommentCount(countCommentTree(Array.isArray(tree) ? tree : []));
  }, []);

  useEffect(() => {
    if (!entityTypeForComments || expandedId == null) {
      setThreadCommentCount(0);
      return;
    }
    let cancelled = false;
    tranApi
      .get(tranEndpoints.comments, { params: { entity_type: entityTypeForComments, record_id: expandedId } })
      .then((res) => {
        if (cancelled) return;
        setThreadCommentCount(countCommentTree(Array.isArray(res.data) ? res.data : []));
      })
      .catch(() => {
        if (!cancelled) setThreadCommentCount(0);
      });
    return () => {
      cancelled = true;
    };
  }, [entityTypeForComments, expandedId]);

  const hasInlineCommentsField = useMemo(
    () => !entityTypeForComments && String(editDraft?.comments ?? '').trim() !== '',
    [entityTypeForComments, editDraft?.comments]
  );

  const createFieldListFull = useMemo(() => {
    const base =
      createFields && createFields.length
        ? createFields
        : detailRow && typeof detailRow === 'object'
          ? Object.keys(detailRow)
          : (columns || []).map((c) => c.field).filter(Boolean);
    const fields = Array.from(new Set([...requiredFields, ...base]))
      .filter((f) => f && !isHiddenKey(f));
    return fields.sort();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [columns, requiredFields, createFields, hiddenFields, detailRow, underscoreIdFieldsVisible]);

  const createFieldList = useMemo(
    () =>
      pinNoteSection ? createFieldListFull.filter((f) => !isPinnedBottomField(f)) : createFieldListFull,
    [createFieldListFull, pinNoteSection]
  );

  const createPinnedBottomFields = useMemo(() => {
    if (!pinNoteSection) return [];
    const keys = createFieldListFull.filter((f) => isPinnedBottomField(f));
    return ['description', 'detail'].filter((k) => keys.includes(k));
  }, [createFieldListFull, pinNoteSection]);

  const openCreate = () => {
    setCreateErr(null);
    const defaults = {};
    for (const f of createFieldListFull) defaults[f] = '';
    setCreateDraft(defaults);
    setCreateOpen(true);
  };

  const submitCreate = async () => {
    if (!onCreate) return;
    setCreateErr(null);
    for (const fk of normalizedDetailJsonFieldKeys) {
      const raw = createDraft[fk];
      if (raw == null || String(raw).trim() === '') continue;
      const r = validateJsonDetailStructure(raw);
      if (!r.ok) {
        setCreateErr(r.error);
        return;
      }
    }
    setCreateBusy(true);
    try {
      let payload = canonicalizeDetailKeys({ ...createDraft });
      // Best-effort numeric coercion for common numeric fields
      for (const k of Object.keys(payload)) {
        if (payload[k] === '') payload[k] = null;
        if (isDateOnlyFieldName(k)) {
          const n = normalizeInputValue(k, payload[k], 'date');
          payload[k] = n === '' ? null : n;
        }
      }
      if (Object.prototype.hasOwnProperty.call(payload, 'gender')) {
        payload.gender = genderSelectToApi(payload.gender);
      }
      serializeLocationJsonPayload(payload, normalizedLocationJsonFieldKeys);
      const createdId = await onCreate(payload);
      if (createdId != null) setExpandedId(createdId);
      setCreateOpen(false);
      setSaveSuccessOpen(true);
    } catch (e) {
      setCreateErr(e?.message || 'Create failed');
    } finally {
      setCreateBusy(false);
    }
  };

  const submitDelete = async () => {
    setDeleteConfirmOpen(false);
    setBatchDeleteConfirmOpen(false);
    if (!onDelete || deleteTargets.length === 0) return;
    setDeleteErr(null);
    setDeleteBusy(true);
    try {
      for (const row of deleteTargets) {
        await onDelete(row);
      }
      if (deleteTargets.some((r) => r.id === expandedId)) {
        setExpandedId(null);
        setDetailDrawerOpen(false);
        setDetailRow(null);
      }
      setRowSelectionModel([]);
    } catch (e) {
      setDeleteErr(e?.message || 'Delete failed');
    } finally {
      setDeleteBusy(false);
    }
  };

  const openDeleteConfirm = () => {
    if (deleteTargets.length === 0) return;
    if (deleteTargets.length > 1) {
      setBatchDeleteConfirmOpen(true);
      return;
    }
    setDeleteConfirmOpen(true);
  };

  const submitSave = async () => {
    setSaveConfirmOpen(false);
    if (expandedId == null) return;
    if (!onUpdate) return;
    setSaveErr(null);
    for (const fk of normalizedDetailJsonFieldKeys) {
      const raw = editDraft[fk];
      if (raw == null || String(raw).trim() === '') continue;
      const r = validateJsonDetailStructure(raw);
      if (!r.ok) {
        setSaveErr(r.error);
        return;
      }
    }
    setSaveBusy(true);
    try {
      let payload = canonicalizeDetailKeys({ ...editDraft });
      for (const k of Object.keys(payload)) {
        if (payload[k] === '') payload[k] = null;
        if (isDateOnlyFieldName(k)) {
          const n = normalizeInputValue(k, payload[k], 'date');
          payload[k] = n === '' ? null : n;
        }
      }
      if (Object.prototype.hasOwnProperty.call(payload, 'gender')) {
        payload.gender = genderSelectToApi(payload.gender);
      }
      serializeLocationJsonPayload(payload, normalizedLocationJsonFieldKeys);
      if (onUpdate) {
        await onUpdate(expandedId, payload);
      }
      if (fetchDetail) {
        try {
          const full = await fetchDetail(expandedId);
          const lowered = Object.fromEntries(
            Object.entries(full || {}).map(([k, v]) => [String(k).toLowerCase(), v])
          );
          setDetailRow(canonicalizeDetailKeys(lowered));
          setDetailDraftSeed((s) => s + 1);
        } catch (refreshErr) {
          // Save already succeeded — do not surface refresh errors as failed saves (avoids bogus "status 500" after OK PUT).
          console.warn('[AdminDataGrid] detail refresh after save failed:', refreshErr);
        }
      }
      setSaveSuccessOpen(true);
    } catch (e) {
      const msg =
        e?.response?.data?.error ||
        (typeof e?.response?.data === 'string' ? e.response.data : null) ||
        e?.message ||
        'Save failed';
      setSaveErr(String(msg));
    } finally {
      setSaveBusy(false);
    }
  };

  const fieldLabel = (field) => {
    const key = String(field || '').toLowerCase();
    if (key === 'facility_code') return platformLabels?.col_facility_code || 'Code';
    if (key === 'facility_type') return platformLabels?.col_facility_type || 'Type';
    if (key === 'facility' || key === 'facility_display') {
      return platformLabels?.term_facility || 'Place';
    }
    if (key === 'facility_id') {
      return `${platformLabels?.term_facility || 'Place'} ID`;
    }
    return formatFieldLabel(field);
  };

  const controlTypeForField = (field) => {
    const name = String(field).toLowerCase();

    if (facilityAsSelect && name === 'facility') {
      return 'facility_select';
    }

    if (facilityIdAsSelect && name === 'facility_id') {
      return 'facility_id_select';
    }

    if (districtIdAsSelect && name === 'district_id') {
      return 'district_id_select';
    }

    if (
      normalizedLocationJsonFieldKeys &&
      normalizedLocationJsonFieldKeys.length > 0 &&
      normalizedLocationJsonFieldKeys.some((k) => String(k).toLowerCase() === name)
    ) {
      return 'location_map';
    }

    if (
      ['employ_type', 'asset_type', 'activity_type', 'facility_type', 'participant_type'].includes(name)
    ) {
      return 'type_dictionary';
    }

    if (
      normalizedDetailJsonFieldKeys &&
      normalizedDetailJsonFieldKeys.length > 0 &&
      normalizedDetailJsonFieldKeys.some((k) => String(k).toLowerCase() === name)
    ) {
      return 'json_detail';
    }

    if (name === 'gender') return 'gender';

    if (name === 'dob') return 'date';

    if (name.includes('email')) return 'email';
    if (name.includes('phone')) return 'phone';

    if ((name.includes('date') && name.includes('time')) || name.includes('datetime') || name.includes('timestamp')) {
      return 'datetime';
    }
    if (name.includes('date') || name.endsWith('_on')) return 'date';
    if (name.includes('time')) return 'time';

    if (name === 'description') {
      return 'long_text';
    }

    if (['comments', 'memo'].some((k) => name.includes(k))) {
      return 'multiline';
    }

    if (
      name.startsWith('is_') ||
      name.endsWith('_flag') ||
      name === 'administrator' ||
      name === 'deactivated' ||
      [
        'disabled',
        'monday',
        'tuesday',
        'wednesday',
        'thursday',
        'friday',
        'saturday',
        'sunday',
        'bus_aide',
        'home_schl',
        'home_trans',
        'shuttle',
        'activitytrip',
        'inactive',
        'in_active',
      ].includes(name)
    ) {
      return 'boolean';
    }

    if (
      name === 'grade' ||
      (name.endsWith('_id') && name !== 'guid') ||
      name.includes('capacity') ||
      name.endsWith('_count') ||
      name.includes('distance') ||
      name.includes('num') ||
      name.includes('rate') ||
      name.includes('time') ||
      name.includes('load_time')
    ) {
      return 'number';
    }

    return 'text';
  };

  const controlTypeForCreateField = (field) => {
    return controlTypeForField(field);
  };

  const dictionaryKeyForField = (field) => {
    const n = String(field || '').toLowerCase();
    if (['employ_type', 'asset_type', 'activity_type', 'facility_type', 'participant_type'].includes(n)) {
      return n;
    }
    return null;
  };

  const renderCreateFieldControl = (field, controlType) => {
    const value = createDraft[field] ?? '';
    const commonTextProps = {
      size: 'small',
      fullWidth: true,
    };

    switch (controlType) {
      case 'location_map':
        return (
          <ActivityLocationFieldEditor
            value={createDraft[field]}
            onChange={(v) => setCreateDraft((prev) => ({ ...prev, [field]: v }))}
          />
        );
      case 'type_dictionary': {
        const dk = dictionaryKeyForField(field);
        const opts = (dk && entityDictionaries[dk]) || [];
        const selVal = value === '' || value == null ? '' : String(value);
        return (
          <FormControl size="small" fullWidth>
            <Select
              value={selVal}
              onChange={(e) =>
                setCreateDraft((prev) => ({
                  ...prev,
                  [field]: e.target.value === '' ? null : e.target.value,
                }))
              }
              displayEmpty
            >
              <MenuItem value="">
                <em>Not set</em>
              </MenuItem>
              {opts.map((o) => (
                <MenuItem key={o.code} value={o.code}>
                  {o.label || o.code}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        );
      }
      case 'facility_select': {
        const selVal = value === '' || value == null ? '' : String(value);
        const has = facilityOptions.some((o) => o.code === selVal);
        const opts = has || !selVal ? facilityOptions : [{ code: selVal, label: selVal }, ...facilityOptions];
        const renderSel = (sel) => {
          if (sel == null || sel === '') {
            return <em>Not set</em>;
          }
          const o = opts.find((x) => x.code === sel);
          const text = o ? o.label : String(sel);
          return (
            <Box component="span" title={text} sx={{ display: 'block', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              {text}
            </Box>
          );
        };
        return (
          <FormControl size="small" fullWidth sx={{ minWidth: 0, maxWidth: '100%' }}>
            <Select
              value={selVal}
              onChange={(e) =>
                setCreateDraft((prev) => ({
                  ...prev,
                  [field]: e.target.value === '' ? null : e.target.value,
                }))
              }
              displayEmpty
              renderValue={renderSel}
              sx={{
                maxWidth: '100%',
                '& .MuiSelect-select': { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' },
              }}
              MenuProps={{
                PaperProps: { sx: { maxWidth: 'min(100vw, 480px)' } },
              }}
            >
              <MenuItem value="">
                <em>Not set</em>
              </MenuItem>
              {opts.map((o) => (
                <MenuItem key={o.code} value={o.code} title={o.label}>
                  {o.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        );
      }
      case 'facility_id_select': {
        const numVal = value === '' || value == null ? '' : String(Number(value));
        const badNum = numVal !== '' && !Number.isFinite(Number(numVal));
        const selVal = badNum ? '' : numVal;
        const has = selVal !== '' && facilityIdOptions.some((o) => String(o.id) === selVal);
        const opts =
          has || selVal === ''
            ? facilityIdOptions
            : [
                {
                  id: Number(selVal),
                  label: `${platformLabels?.term_facility || 'Place'} #${selVal}`,
                },
                ...facilityIdOptions,
              ];
        const renderSel = (sel) => {
          if (sel == null || sel === '') {
            return <em>Not set</em>;
          }
          const o = opts.find((x) => String(x.id) === String(sel));
          const text = o ? o.label : String(sel);
          return (
            <Box component="span" title={text} sx={{ display: 'block', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              {text}
            </Box>
          );
        };
        return (
          <FormControl size="small" fullWidth sx={{ minWidth: 0, maxWidth: '100%' }}>
            <Select
              value={selVal}
              onChange={(e) => {
                const v = e.target.value;
                setCreateDraft((prev) => ({
                  ...prev,
                  [field]: v === '' ? null : Number(v),
                }));
              }}
              displayEmpty
              renderValue={renderSel}
              sx={{
                maxWidth: '100%',
                '& .MuiSelect-select': { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' },
              }}
              MenuProps={{
                PaperProps: { sx: { maxWidth: 'min(100vw, 480px)' } },
              }}
            >
              <MenuItem value="">
                <em>Not set</em>
              </MenuItem>
              {opts.map((o) => (
                <MenuItem key={o.id} value={String(o.id)} title={o.label}>
                  {o.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        );
      }
      case 'district_id_select': {
        const numVal = value === '' || value == null ? '' : String(Number(value));
        const badNum = numVal !== '' && !Number.isFinite(Number(numVal));
        const selVal = badNum ? '' : numVal;
        const has = selVal !== '' && districtIdOptions.some((o) => String(o.id) === selVal);
        const opts =
          has || selVal === ''
            ? districtIdOptions
            : [{ id: Number(selVal), label: `District #${selVal}` }, ...districtIdOptions];
        const renderSel = (sel) => {
          if (sel == null || sel === '') {
            return <em>Not set</em>;
          }
          const o = opts.find((x) => String(x.id) === String(sel));
          const text = o ? o.label : String(sel);
          return (
            <Box component="span" title={text} sx={{ display: 'block', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              {text}
            </Box>
          );
        };
        return (
          <FormControl size="small" fullWidth sx={{ minWidth: 0, maxWidth: '100%' }}>
            <Select
              value={selVal}
              onChange={(e) => {
                const v = e.target.value;
                setCreateDraft((prev) => ({
                  ...prev,
                  [field]: v === '' ? null : Number(v),
                }));
              }}
              displayEmpty
              renderValue={renderSel}
              sx={{
                maxWidth: '100%',
                '& .MuiSelect-select': { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' },
              }}
              MenuProps={{
                PaperProps: { sx: { maxWidth: 'min(100vw, 480px)' } },
              }}
            >
              <MenuItem value="">
                <em>Not set</em>
              </MenuItem>
              {opts.map((o) => (
                <MenuItem key={o.id} value={String(o.id)} title={o.label}>
                  {o.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        );
      }
      case 'json_detail':
        return (
          <TextField
            {...commonTextProps}
            multiline
            minRows={10}
            value={value ?? ''}
            onChange={(e) => setCreateDraft((prev) => ({ ...prev, [field]: e.target.value }))}
            placeholder='{"example": true}'
            sx={{ '& .MuiInputBase-input': { fontFamily: 'ui-monospace, monospace', fontSize: 13 } }}
          />
        );
      case 'date':
        return (
          <TextField
            {...commonTextProps}
            type="date"
            value={value ?? ''}
            onChange={(e) =>
              setCreateDraft((prev) => ({
                ...prev,
                [field]: e.target.value,
              }))
            }
          />
        );
      case 'datetime':
        return (
          <TextField
            {...commonTextProps}
            type="datetime-local"
            value={value ?? ''}
            onChange={(e) =>
              setCreateDraft((prev) => ({
                ...prev,
                [field]: e.target.value,
              }))
            }
          />
        );
      case 'time':
        return (
          <TextField
            {...commonTextProps}
            type="time"
            value={value ?? ''}
            onChange={(e) =>
              setCreateDraft((prev) => ({
                ...prev,
                [field]: e.target.value,
              }))
            }
          />
        );
      case 'boolean':
        return (
          <FormControlLabel
            control={
              <Checkbox
                checked={Boolean(value)}
                onChange={(e) =>
                  setCreateDraft((prev) => ({
                    ...prev,
                    [field]: e.target.checked,
                  }))
                }
              />
            }
            label=""
          />
        );
      case 'gender': {
        const gv = genderApiToSelect(value);
        return (
          <FormControl size="small" fullWidth>
            <Select
              value={gv}
              onChange={(e) =>
                setCreateDraft((prev) => ({
                  ...prev,
                  [field]: e.target.value,
                }))
              }
              displayEmpty
            >
              <MenuItem value="">
                <em>Not set</em>
              </MenuItem>
              <MenuItem value="M">Male</MenuItem>
              <MenuItem value="F">Female</MenuItem>
            </Select>
          </FormControl>
        );
      }
      case 'number':
        return (
          <TextField
            {...commonTextProps}
            type="number"
            value={value ?? ''}
            onChange={(e) =>
              setCreateDraft((prev) => ({
                ...prev,
                [field]: sanitizeNumericInput(e.target.value, numberOptionsForFieldName(field)),
              }))
            }
            inputProps={{ inputMode: 'decimal', step: 'any' }}
          />
        );
      case 'email':
        return (
          <TextField
            {...commonTextProps}
            type="email"
            value={value ?? ''}
            onChange={(e) =>
              setCreateDraft((prev) => ({
                ...prev,
                [field]: sanitizeTextInput(e.target.value),
              }))
            }
          />
        );
      case 'phone':
        return (
          <TextField
            {...commonTextProps}
            type="tel"
            value={value ?? ''}
            onChange={(e) =>
              setCreateDraft((prev) => ({
                ...prev,
                [field]: sanitizeTextInput(e.target.value),
              }))
            }
          />
        );
      case 'long_text':
        return (
          <TextField
            {...commonTextProps}
            multiline
            minRows={3}
            maxRows={10}
            value={value ?? ''}
            onChange={(e) =>
              setCreateDraft((prev) => ({
                ...prev,
                [field]: sanitizeTextInput(e.target.value),
              }))
            }
            sx={{
              '& .MuiInputBase-root': { alignItems: 'flex-start' },
              '& .MuiInputBase-input': { overflow: 'auto', lineHeight: 1.5 },
            }}
          />
        );
      case 'multiline':
        return (
          <TextField
            {...commonTextProps}
            multiline
            minRows={String(field).toLowerCase().endsWith('detail') ? 4 : 2}
            maxRows={12}
            value={value ?? ''}
            onChange={(e) =>
              setCreateDraft((prev) => ({
                ...prev,
                [field]: sanitizeTextInput(e.target.value),
              }))
            }
          />
        );
      default:
        return (
          <TextField
            {...commonTextProps}
            value={value ?? ''}
            onChange={(e) =>
              setCreateDraft((prev) => ({
                ...prev,
                [field]: sanitizeTextInput(e.target.value),
              }))
            }
          />
        );
    }
  };

  const renderDetailFieldControl = (field, controlType) => {
    if (controlType === 'location_map') {
      return (
        <ActivityLocationFieldEditor
          value={editDraft[field]}
          onChange={(v) => setEditDraft((prev) => ({ ...prev, [field]: v }))}
        />
      );
    }
    if (controlType === 'json_detail') {
      const rawValue = editDraft[field];
      const jsonMode = jsonDetailEditorMode[field] || 'preview';
      return (
        <JsonDetailEditor
          value={rawValue}
          onChange={(v) => setEditDraft((prev) => ({ ...prev, [field]: v }))}
          mode={jsonMode}
          onModeChange={(m) => setJsonDetailEditorMode((prev) => ({ ...prev, [field]: m }))}
        />
      );
    }
    const rawValue = editDraft[field];

    if (controlType === 'type_dictionary') {
      const dk = dictionaryKeyForField(field);
      const opts = (dk && entityDictionaries[dk]) || [];
      const selVal = rawValue == null || rawValue === '' ? '' : String(rawValue);
      return (
        <FormControl size="small" fullWidth>
          <Select
            value={selVal}
            onChange={(e) => {
              const v = e.target.value === '' ? null : e.target.value;
              setEditDraft((prev) => ({ ...prev, [field]: v }));
            }}
            displayEmpty
          >
            <MenuItem value="">
              <em>Not set</em>
            </MenuItem>
            {opts.map((o) => (
              <MenuItem key={o.code} value={o.code}>
                {o.label || o.code}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      );
    }

    if (controlType === 'facility_select') {
      const selVal = rawValue == null || rawValue === '' ? '' : String(rawValue);
      const has = facilityOptions.some((o) => o.code === selVal);
      const opts = has || !selVal ? facilityOptions : [{ code: selVal, label: selVal }, ...facilityOptions];
      const renderSel = (sel) => {
        if (sel == null || sel === '') {
          return <em>Not set</em>;
        }
        const o = opts.find((x) => x.code === sel);
        const text = o ? o.label : String(sel);
        return (
          <Box component="span" title={text} sx={{ display: 'block', overflow: 'hidden', textOverflow: 'ellipsis' }}>
            {text}
          </Box>
        );
      };
      return (
        <FormControl size="small" fullWidth sx={{ minWidth: 0, maxWidth: '100%' }}>
          <Select
            value={selVal}
            onChange={(e) => {
              const v = e.target.value === '' ? null : e.target.value;
              setEditDraft((prev) => ({ ...prev, [field]: v }));
            }}
            displayEmpty
            renderValue={renderSel}
            sx={{
              maxWidth: '100%',
              '& .MuiSelect-select': { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' },
            }}
            MenuProps={{
              PaperProps: { sx: { maxWidth: 'min(100vw, 480px)' } },
            }}
          >
            <MenuItem value="">
              <em>Not set</em>
            </MenuItem>
            {opts.map((o) => (
              <MenuItem key={o.code} value={o.code} title={o.label}>
                {o.label}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      );
    }

    if (controlType === 'facility_id_select') {
      const numVal = rawValue == null || rawValue === '' ? '' : String(Number(rawValue));
      const badNum = numVal !== '' && !Number.isFinite(Number(numVal));
      const selVal = badNum ? '' : numVal;
      const has = selVal !== '' && facilityIdOptions.some((o) => String(o.id) === selVal);
      const opts =
        has || selVal === ''
          ? facilityIdOptions
          : [
              {
                id: Number(selVal),
                label: `${platformLabels?.term_facility || 'Place'} #${selVal}`,
              },
              ...facilityIdOptions,
            ];
      const renderSel = (sel) => {
        if (sel == null || sel === '') {
          return <em>Not set</em>;
        }
        const o = opts.find((x) => String(x.id) === String(sel));
        const text = o ? o.label : String(sel);
        return (
          <Box component="span" title={text} sx={{ display: 'block', overflow: 'hidden', textOverflow: 'ellipsis' }}>
            {text}
          </Box>
        );
      };
      return (
        <FormControl size="small" fullWidth sx={{ minWidth: 0, maxWidth: '100%' }}>
          <Select
            value={selVal}
            onChange={(e) => {
              const v = e.target.value;
              setEditDraft((prev) => ({
                ...prev,
                [field]: v === '' ? null : Number(v),
              }));
            }}
            displayEmpty
            renderValue={renderSel}
            sx={{
              maxWidth: '100%',
              '& .MuiSelect-select': { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' },
            }}
            MenuProps={{
              PaperProps: { sx: { maxWidth: 'min(100vw, 480px)' } },
            }}
          >
            <MenuItem value="">
              <em>Not set</em>
            </MenuItem>
            {opts.map((o) => (
              <MenuItem key={o.id} value={String(o.id)} title={o.label}>
                {o.label}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      );
    }

    if (controlType === 'district_id_select') {
      const numVal = rawValue == null || rawValue === '' ? '' : String(Number(rawValue));
      const badNum = numVal !== '' && !Number.isFinite(Number(numVal));
      const selVal = badNum ? '' : numVal;
      const has = selVal !== '' && districtIdOptions.some((o) => String(o.id) === selVal);
      const opts =
        has || selVal === ''
          ? districtIdOptions
          : [{ id: Number(selVal), label: `District #${selVal}` }, ...districtIdOptions];
      const renderSel = (sel) => {
        if (sel == null || sel === '') {
          return <em>Not set</em>;
        }
        const o = opts.find((x) => String(x.id) === String(sel));
        const text = o ? o.label : String(sel);
        return (
          <Box component="span" title={text} sx={{ display: 'block', overflow: 'hidden', textOverflow: 'ellipsis' }}>
            {text}
          </Box>
        );
      };
      return (
        <FormControl size="small" fullWidth sx={{ minWidth: 0, maxWidth: '100%' }}>
          <Select
            value={selVal}
            onChange={(e) => {
              const v = e.target.value;
              setEditDraft((prev) => ({
                ...prev,
                [field]: v === '' ? null : Number(v),
              }));
            }}
            displayEmpty
            renderValue={renderSel}
            sx={{
              maxWidth: '100%',
              '& .MuiSelect-select': { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' },
            }}
            MenuProps={{
              PaperProps: { sx: { maxWidth: 'min(100vw, 480px)' } },
            }}
          >
            <MenuItem value="">
              <em>Not set</em>
            </MenuItem>
            {opts.map((o) => (
              <MenuItem key={o.id} value={String(o.id)} title={o.label}>
                {o.label}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      );
    }

    const value = normalizeInputValue(field, rawValue, controlType);
    const commonTextProps = { size: 'small', fullWidth: true };

    const update = (v) => setEditDraft((prev) => ({ ...prev, [field]: v }));

    switch (controlType) {
      case 'date':
        return (
          <TextField
            {...commonTextProps}
            type="date"
            value={value ?? ''}
            onChange={(e) => update(e.target.value || null)}
          />
        );
      case 'datetime':
        return (
          <TextField
            {...commonTextProps}
            type="datetime-local"
            value={value ?? ''}
            onChange={(e) => update(e.target.value || null)}
          />
        );
      case 'time':
        return (
          <TextField
            {...commonTextProps}
            type="time"
            value={value ?? ''}
            onChange={(e) => update(e.target.value || null)}
          />
        );
      case 'boolean':
        return (
          <FormControlLabel
            control={<Checkbox checked={Boolean(value)} onChange={(e) => update(e.target.checked)} />}
            label=""
          />
        );
      case 'gender':
        return (
          <FormControl size="small" fullWidth>
            <Select value={value ?? ''} onChange={(e) => update(e.target.value)} displayEmpty>
              <MenuItem value=""><em>Not set</em></MenuItem>
              <MenuItem value="M">Male</MenuItem>
              <MenuItem value="F">Female</MenuItem>
            </Select>
          </FormControl>
        );
      case 'number':
        return (
          <TextField
            {...commonTextProps}
            type="number"
            value={value ?? ''}
            onChange={(e) => {
              const clean = sanitizeNumericInput(e.target.value, numberOptionsForFieldName(field));
              update(clean === '' ? null : clean);
            }}
            inputProps={{ inputMode: 'decimal', step: 'any' }}
          />
        );
      case 'email':
        return (
          <TextField
            {...commonTextProps}
            type="email"
            value={value ?? ''}
            onChange={(e) => update(sanitizeTextInput(e.target.value))}
          />
        );
      case 'phone':
        return (
          <TextField
            {...commonTextProps}
            type="tel"
            value={value ?? ''}
            onChange={(e) => update(sanitizeTextInput(e.target.value))}
          />
        );
      case 'long_text':
        return (
          <TextField
            {...commonTextProps}
            multiline
            minRows={3}
            maxRows={10}
            value={value ?? ''}
            onChange={(e) => update(sanitizeTextInput(e.target.value))}
            sx={{
              '& .MuiInputBase-root': { alignItems: 'flex-start' },
              '& .MuiInputBase-input': { overflow: 'auto', lineHeight: 1.5 },
            }}
          />
        );
      case 'multiline':
        return (
          <TextField
            {...commonTextProps}
            multiline
            minRows={String(field).toLowerCase().endsWith('detail') ? 4 : 2}
            maxRows={12}
            value={value ?? ''}
            onChange={(e) => update(sanitizeTextInput(e.target.value))}
          />
        );
      default:
        return (
          <TextField
            {...commonTextProps}
            value={value ?? ''}
            onChange={(e) => update(sanitizeTextInput(e.target.value))}
          />
        );
    }
  };

  const keysSignature = useMemo(() => Object.keys(editDraft).sort().join('|'), [editDraft]);

  // Intentionally omit editDraft from deps; keysSignature tracks key set. Including editDraft would reset order every keystroke.
  useEffect(() => {
    if (expandedId == null || detailBusy) return;
    const keys = Object.keys(editDraft).filter((k) => !pinNoteSection || !isPinnedBottomField(k));
    if (keys.length === 0) {
      setDetailFieldOrder([]);
      return;
    }
    setDetailFieldOrder(getFieldOrder(storageKey, keys));
    // eslint-disable-next-line react-hooks/exhaustive-deps -- see comment above
  }, [expandedId, detailBusy, keysSignature, storageKey, pinNoteSection]);

  const persistLayoutOrder = useCallback((order) => {
    try {
      localStorage.setItem(STORAGE_PREFIX + storageKey, JSON.stringify(order));
    } catch {
      /* ignore */
    }
  }, [storageKey]);

  const handleReorderDetailFields = (fromIndex, toIndex) => {
    if (fromIndex === toIndex) return;
    setDetailFieldOrder((prev) => {
      const keys = Object.keys(editDraft).filter((k) => !pinNoteSection || !isPinnedBottomField(k));
      const keySet = new Set(keys);
      let current = prev.filter((f) => keySet.has(f));
      const missing = keys.filter((k) => !current.includes(k));
      current = [...current, ...missing];
      if (fromIndex < 0 || toIndex < 0 || fromIndex >= current.length || toIndex >= current.length) return prev;
      const arr = [...current];
      const [item] = arr.splice(fromIndex, 1);
      arr.splice(toIndex, 0, item);
      persistLayoutOrder(arr);
      return arr;
    });
    setDragOverDetailIndex(null);
  };

  const detailEditFieldList = useMemo(() => {
    const keys = Object.keys(editDraft).filter((k) => !pinNoteSection || !isPinnedBottomField(k));
    if (keys.length === 0) return [];
    const keySet = new Set(keys);
    if (detailFieldOrder.length === 0) {
      return getFieldOrder(storageKey, keys);
    }
    const ordered = detailFieldOrder.filter((f) => keySet.has(f));
    const missing = keys.filter((k) => !ordered.includes(k));
    return [...ordered, ...missing];
  }, [editDraft, detailFieldOrder, storageKey, pinNoteSection]);

  const pinnedBottomFields = useMemo(() => {
    if (!pinNoteSection) return [];
    const keys = Object.keys(editDraft);
    return ['description', 'detail'].filter((k) => keys.includes(k));
  }, [editDraft, pinNoteSection]);

  const renderPinnedNoteBlock = (field, isCreate) => {
    const ct = isCreate ? controlTypeForCreateField(field) : controlTypeForField(field);
    const value = isCreate ? createDraft[field] ?? '' : editDraft[field] ?? '';
    const setDraft = isCreate ? setCreateDraft : setEditDraft;

    const onImportFile = (e) => {
      const file = e.target.files?.[0];
      e.target.value = '';
      if (!file) return;
      const reader = new FileReader();
      reader.onload = () => {
        setDraft((prev) => ({ ...prev, [field]: String(reader.result ?? '') }));
      };
      reader.readAsText(file);
    };

    const importBtn = (
      <Button component="label" size="small" variant="outlined" startIcon={<UploadFileIcon />} sx={{ flexShrink: 0 }}>
        Import JSON file
        <input type="file" hidden accept="application/json,.json,text/plain" onChange={onImportFile} />
      </Button>
    );

    if (String(field).toLowerCase() === 'description') {
      return (
        <Box key={field} sx={{ width: '100%' }}>
          <Paper
            variant="outlined"
            sx={{
              p: 2,
              ...(isDark ? DETAIL_CARD_SX_DARK : { bgcolor: 'background.default', border: '1px solid rgba(0,0,0,0.1)' }),
            }}
          >
            <TextField
              size="small"
              fullWidth
              multiline
              minRows={3}
              maxRows={10}
              placeholder="Description…"
              value={value ?? ''}
              onChange={(e) => setDraft((prev) => ({ ...prev, [field]: sanitizeTextInput(e.target.value) }))}
              sx={{
                '& .MuiInputBase-root': { alignItems: 'flex-start' },
                '& .MuiInputBase-input': {
                  overflow: 'auto',
                  lineHeight: 1.5,
                  fontSize: '0.9375rem',
                },
              }}
            />
          </Paper>
        </Box>
      );
    }

    const jsonMode = jsonDetailEditorMode[field] || 'preview';
    const editJsonToolbarBtn =
      !isCreate && ct === 'json_detail' ? (
        jsonMode === 'preview' ? (
          <Button
            size="small"
            variant="contained"
            color="primary"
            sx={{ flexShrink: 0 }}
            onClick={() => setJsonDetailEditorMode((prev) => ({ ...prev, [field]: 'tree' }))}
          >
            Edit JSON
          </Button>
        ) : jsonMode === 'tree' ? (
          <>
            <Button
              size="small"
              variant="outlined"
              sx={{ flexShrink: 0 }}
              onClick={() => setJsonDetailEditorMode((prev) => ({ ...prev, [field]: 'preview' }))}
            >
              Card view
            </Button>
            <Button
              size="small"
              variant="outlined"
              sx={{ flexShrink: 0 }}
              onClick={() => setJsonDetailEditorMode((prev) => ({ ...prev, [field]: 'raw' }))}
            >
              Edit raw JSON
            </Button>
          </>
        ) : (
          <>
            <Button
              size="small"
              variant="outlined"
              sx={{ flexShrink: 0 }}
              onClick={() => setJsonDetailEditorMode((prev) => ({ ...prev, [field]: 'tree' }))}
            >
              Tree editor
            </Button>
            <Button
              size="small"
              variant="outlined"
              sx={{ flexShrink: 0 }}
              onClick={() => setJsonDetailEditorMode((prev) => ({ ...prev, [field]: 'preview' }))}
            >
              Card view
            </Button>
          </>
        )
      ) : null;

    if (ct === 'json_detail' && !isCreate) {
      return (
        <Box key={field} sx={{ width: '100%' }}>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 1, mb: 1 }}>
            {importBtn}
            {editJsonToolbarBtn}
          </Box>
          <JsonDetailEditor
            value={editDraft[field]}
            onChange={(v) => setEditDraft((prev) => ({ ...prev, [field]: v }))}
            mode={jsonMode}
            onModeChange={(m) => setJsonDetailEditorMode((prev) => ({ ...prev, [field]: m }))}
            externalEditControls
          />
        </Box>
      );
    }

    if (ct === 'json_detail' && isCreate) {
      return (
        <Box key={field} sx={{ width: '100%' }}>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 1, mb: 1 }}>
            {importBtn}
          </Box>
          <TextField
            size="small"
            fullWidth
            multiline
            minRows={14}
            value={value ?? ''}
            onChange={(e) => setDraft((prev) => ({ ...prev, [field]: e.target.value }))}
            placeholder="Paste or import JSON…"
            sx={{ '& .MuiInputBase-input': { fontFamily: 'ui-monospace, monospace', fontSize: 13 } }}
          />
        </Box>
      );
    }

    return (
      <Box key={field} sx={{ width: '100%' }}>
        <Typography variant="body2" color="text.secondary">
          Unsupported pinned field
        </Typography>
      </Box>
    );
  };

  const canSaveDetail =
    expandedId != null &&
    !saveBusy &&
    onUpdate &&
    (detailEditFieldList.length > 0 || pinnedBottomFields.length > 0);

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        minHeight: 0,
        width: '100%',
        height: '100%',
        overflow: 'hidden',
      }}
    >
      <Box sx={{ display: 'flex', flexDirection: 'row', flex: 1, minHeight: 0, gap: 0 }}>
        <Box
          sx={{
            flex: '1 1 100%',
            minWidth: 0,
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
          }}
        >
          <Toolbar
            sx={{
              minHeight: { xs: 'auto', sm: 48 },
              px: { xs: 0.75, sm: 1 },
              py: 0.5,
              borderBottom: isDark ? '1px solid rgba(255,255,255,0.08)' : '1px solid rgba(0,0,0,0.12)',
              bgcolor: 'background.paper',
              flexWrap: 'wrap',
              gap: 0.5,
              alignItems: 'center',
            }}
          >
            <Box sx={{ display: 'flex', flexDirection: 'column', minWidth: 0, flex: { xs: '1 1 100%', sm: 1 } }}>
              <Typography variant="subtitle1" fontWeight={700} noWrap sx={{ color: 'text.primary' }}>
                {title}
              </Typography>
              <Typography
                variant="caption"
                color="text.secondary"
                noWrap
                sx={{ display: { xs: 'none', sm: 'block' } }}
              >
                Select a row to open the detail panel
              </Typography>
            </Box>
            {onCreate && (
              <IconButton size="small" color="inherit" onClick={openCreate} title="Add record" sx={{ width: 40, height: 40 }}>
                <AddOutlinedIcon />
              </IconButton>
            )}
            <IconButton
              size="small"
              color="inherit"
              onClick={openDeleteConfirm}
              disabled={!onDelete || deleteTargets.length === 0 || deleteBusy}
              sx={{ width: 40, height: 40 }}
              title={
                deleteTargets.length > 1
                  ? `Delete ${deleteTargets.length} records`
                  : deleteTargets.length === 1
                    ? 'Delete selected record'
                    : 'Delete selected'
              }
            >
              <DeleteOutlineOutlinedIcon />
            </IconButton>
            <TextField
              size="small"
              placeholder="Search..."
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              InputProps={{ startAdornment: <SearchIcon sx={{ mr: 0.5, color: 'text.secondary' }} /> }}
              sx={{
                width: { xs: '100%', sm: 220 },
                flex: { xs: '1 1 100%', sm: '0 0 auto' },
                ml: { xs: 0, sm: 1 },
                order: { xs: 10, sm: 0 },
                '& .MuiInputBase-input': { fontSize: 16 },
              }}
            />
            <Badge
              badgeContent={activeFilterCount}
              color="primary"
              invisible={activeFilterCount === 0}
              sx={{ ml: { xs: 0, sm: 0.5 } }}
            >
              <IconButton
                size="small"
                color={activeFilterCount ? 'primary' : 'inherit'}
                onClick={() => setFilterModalOpen(true)}
                title="Advanced filters & saved filters"
                sx={{ width: 40, height: 40 }}
              >
                <FilterListIcon fontSize="small" />
              </IconButton>
            </Badge>
            <IconButton
              size="small"
              color="inherit"
              onClick={() => setColorTagsOpen(true)}
              title="Color tags — row colors (display only)"
              sx={{ ml: { xs: 0, sm: 0.5 }, width: 40, height: 40 }}
            >
              <PaletteOutlinedIcon fontSize="small" />
            </IconButton>
            <FormControl size="small" sx={{ minWidth: { xs: 96, sm: 110 }, ml: { xs: 'auto', sm: 1 } }}>
              <Select
                value={pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value));
                  setPaginationPage(0);
                }}
                displayEmpty
              >
                <MenuItem value={50}>50 / page</MenuItem>
                <MenuItem value={100}>100 / page</MenuItem>
              </Select>
            </FormControl>
          </Toolbar>

          {/* Grid wrapper: only this area scrolls horizontally (DataGrid pins header/footer) */}
          <Box sx={{ flex: 1, minHeight: 0, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
            {colorRowCss ? <style>{colorRowCss}</style> : null}
            <DataGrid
              rows={filteredRows}
              columns={gridColumns}
              loading={loading}
              checkboxSelection={Boolean(onDelete)}
              checkboxSelectionVisibleOnly
              rowSelectionModel={rowSelectionModel}
              onRowSelectionModelChange={(model) => setRowSelectionModel(model)}
              disableMultipleRowSelection={false}
              scrollbarSize={8}
              disableRowSelectionOnClick
              onRowClick={(params) => {
                setExpandedId(params.id);
                setDetailDrawerOpen(true);
              }}
              sortingOrder={['asc', 'desc']}
              pageSizeOptions={[50, 100]}
              paginationModel={{ page: paginationPage, pageSize }}
              onPaginationModelChange={(model) => {
                setPaginationPage(model.page);
                if (model.pageSize !== pageSize) setPageSize(model.pageSize);
              }}
              rowHeight={isPhone ? Math.min(rowHeight, 44) : rowHeight}
              columnHeaderHeight={isPhone ? 44 : 56}
              density={isCompact ? 'compact' : 'standard'}
              getRowClassName={(params) => {
                let base = getRowClassName ? getRowClassName(params) : '';
                const ct = colorRowMeta.meta.get(params.id);
                if (ct?.token) base = `${base} ${ct.token}`.trim();
                const selected = params.id === expandedId ? ' row-selected' : '';
                return (base + selected).trim() || undefined;
              }}
              sx={{
                border: 'none',
                flex: 1,
                minHeight: 0,
                bgcolor: 'background.paper',
                '& .MuiDataGrid-root': {
                  height: '100%',
                  minHeight: 0,
                  '--DataGrid-scrollbarSize': '8px',
                },
                '& .MuiDataGrid-row.row-selected': {
                  backgroundColor: isDark ? 'rgba(230, 81, 0, 0.28)' : 'rgba(245, 124, 0, 0.18)',
                  '&:hover': {
                    backgroundColor: isDark ? 'rgba(230, 81, 0, 0.35)' : 'rgba(245, 124, 0, 0.25)',
                  },
                },
                '& .MuiDataGrid-columnHeaders': {
                  position: 'relative',
                  zIndex: 2,
                  backgroundColor: headerAccentBg,
                  color: headerAccentColor,
                  borderBottom: isDark
                    ? '1px solid rgba(6, 78, 59, 0.6)'
                    : '1px solid rgba(0, 0, 0, 0.08)',
                },
                '& .MuiDataGrid-columnHeader': {
                  backgroundColor: headerAccentBg,
                  color: headerAccentColor,
                  fontWeight: 600,
                  fontSize: 13,
                },
                '& .MuiDataGrid-columnHeader .MuiIconButton-root': {
                  color: headerAccentColor,
                },
                '& .MuiDataGrid-columnSeparator': {
                  color: isDark ? 'rgba(255,255,255,0.25)' : 'rgba(0,0,0,0.12)',
                },
                '& .MuiDataGrid-row:hover': {
                  bgcolor: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.04)',
                },
                '& .MuiDataGrid-cell:focus, & .MuiDataGrid-columnHeader:focus, & .MuiDataGrid-cell:focus-within, & .MuiDataGrid-columnHeader:focus-within': {
                  outline: 'none',
                },
                '& .MuiDataGrid-footerContainer': {
                  borderTop: isDark ? '1px solid rgba(255,255,255,0.08)' : '1px solid rgba(0,0,0,0.12)',
                  bgcolor: 'background.paper',
                  minHeight: isPhone ? 48 : 56,
                  flexWrap: 'wrap',
                },
                '& .MuiTablePagination-selectLabel, & .MuiTablePagination-displayedRows': {
                  fontSize: isPhone ? 12 : undefined,
                },
                '& .MuiDataGrid-cell': {
                  borderBottom: isDark ? '1px solid rgba(255,255,255,0.06)' : '1px solid rgba(0,0,0,0.08)',
                  fontSize: isPhone ? 13 : undefined,
                },
                '& .MuiDataGrid-main': {
                  minHeight: 0,
                  overflow: 'hidden',
                },
                /* v7 hides native scrollbars on virtualScroller; the visible bar is .MuiDataGrid-scrollbar--vertical */
                '& .MuiDataGrid-scrollbar--vertical': {
                  zIndex: 1,
                  scrollbarWidth: 'thin',
                  scrollbarColor: isDark ? 'rgba(255, 255, 255, 0.35) #14161a' : 'rgba(0, 0, 0, 0.25) rgba(0, 0, 0, 0.06)',
                },
                '& .MuiDataGrid-scrollbar--vertical::-webkit-scrollbar': {
                  width: 8,
                },
                '& .MuiDataGrid-scrollbar--vertical::-webkit-scrollbar-thumb': {
                  backgroundColor: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.25)',
                  borderRadius: 4,
                },
                '& .MuiDataGrid-virtualScroller': {
                  bgcolor: 'background.paper',
                },
              }}
            />
          </Box>
        </Box>
      </Box>

      <Drawer
        anchor="right"
        open={detailDrawerOpen && expandedId != null}
        onClose={() => setDetailDrawerOpen(false)}
        PaperProps={{
          sx: {
            width: { xs: '100%', sm: '66.666vw' },
            maxWidth: '100%',
            top: { xs: 0, sm: `calc(${ADMIN_RIGHT_PANEL_TOP}px + env(safe-area-inset-top, 0px))` },
            height: {
              xs: '100%',
              sm: adminRightPanelHeightCalc,
            },
            maxHeight: {
              xs: '100%',
              sm: adminRightPanelHeightCalc,
            },
            boxSizing: 'border-box',
            display: 'flex',
            flexDirection: 'column',
            bgcolor: isDark ? DETAIL_DRAWER_BG_DARK : 'background.default',
            pt: { xs: 'env(safe-area-inset-top)', sm: 0 },
            pb: { xs: 'env(safe-area-inset-bottom)', sm: 0 },
          },
        }}
        ModalProps={{ keepMounted: false, disableScrollLock: true }}
      >
        <Toolbar
          sx={{
            minHeight: 48,
            px: 1.5,
            py: 0.5,
            borderBottom: isDark ? '1px solid rgba(255,255,255,0.08)' : '1px solid rgba(0,0,0,0.12)',
            bgcolor: isDark ? DETAIL_PANEL_BG_DARK : 'background.paper',
            gap: 1,
          }}
        >
          <Typography
            variant="subtitle1"
            fontWeight={700}
            sx={{
              color: 'text.primary',
              flex: 1,
              minWidth: 0,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
            title={detailPanelTitle}
          >
            {detailPanelTitle}
          </Typography>
          {hasInlineCommentsField && (
            <Tooltip title="This record has comments">
              <CommentOutlinedIcon fontSize="small" sx={{ color: 'secondary.main', flexShrink: 0 }} aria-label="Has comments" />
            </Tooltip>
          )}
          <IconButton size="small" onClick={() => setDetailDrawerOpen(false)} aria-label="Close detail panel">
            <CloseIcon />
          </IconButton>
        </Toolbar>
        <Box
          sx={{
            flex: 1,
            minHeight: 0,
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
          }}
        >
          <Box
            sx={{
              flex: 1,
              minHeight: 0,
              overflowY: 'auto',
              overflowX: 'hidden',
              scrollbarWidth: 'thin',
              ...(isDark && { bgcolor: DETAIL_DRAWER_BG_DARK }),
            }}
          >
            {deleteErr && (
              <Typography color="error" variant="caption" sx={{ px: 2, pt: 1, display: 'block' }}>
                {deleteErr}
              </Typography>
            )}
            {saveErr && (
              <Typography color="error" variant="caption" sx={{ px: 2, pt: 1, display: 'block' }}>
                {saveErr}
              </Typography>
            )}
            {detailBusy && (
              <Typography color="text.secondary" variant="caption" sx={{ px: 2, pt: 1, display: 'block' }}>
                Loading full details…
              </Typography>
            )}
            {!detailRow && !detailBusy && (
              <Typography color="text.secondary" sx={{ p: 2 }}>
                Select a row to view and edit details.
              </Typography>
            )}
            {detailRow && !detailBusy && detailEditFieldList.length > 0 && (
              <Box
                sx={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr',
                  gap: 1.5,
                  p: 1.5,
                  alignContent: 'start',
                  minWidth: 0,
                }}
              >
                {detailEditFieldList.map((field, i) => {
                    const label = fieldLabel(field);
                    const controlType = controlTypeForField(field);
                    const hideEmptyJsonLabel =
                      controlType === 'json_detail' && isEmptyJsonDetailValue(editDraft[field]);
                    return (
                      <Paper
                        key={field}
                        elevation={0}
                        onDragOver={(e) => {
                          e.preventDefault();
                          e.dataTransfer.dropEffect = 'move';
                          setDragOverDetailIndex(i);
                        }}
                        onDragLeave={() => setDragOverDetailIndex(null)}
                        onDrop={(e) => {
                          e.preventDefault();
                          const from = parseInt(
                            e.dataTransfer.getData('application/x-detail-field-index'),
                            10
                          );
                          if (!Number.isNaN(from)) handleReorderDetailFields(from, i);
                          setDragOverDetailIndex(null);
                        }}
                        sx={{
                          p: 1.5,
                          ...(isDark
                            ? DETAIL_CARD_SX_DARK
                            : { bgcolor: 'action.hover', border: '1px solid rgba(0,0,0,0.12)' }),
                          borderRadius: 1,
                          display: 'flex',
                          flexDirection: 'column',
                          gap: 0.75,
                          minHeight: 56,
                          minWidth: 0,
                          ...(controlType === 'json_detail' || controlType === 'long_text' || controlType === 'location_map'
                            ? { gridColumn: '1 / -1' }
                            : {}),
                          opacity: dragOverDetailIndex === i ? 0.92 : 1,
                          outline:
                            dragOverDetailIndex === i ? `2px dashed ${theme.palette.primary.main}` : 'none',
                          outlineOffset: 0,
                          transition: 'opacity 0.15s ease',
                        }}
                      >
                        <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 0.75 }}>
                          <IconButton
                            size="small"
                            draggable
                            onDragStart={(e) => {
                              e.dataTransfer.setData('application/x-detail-field-index', String(i));
                              e.dataTransfer.effectAllowed = 'move';
                            }}
                            onDragEnd={() => setDragOverDetailIndex(null)}
                            sx={{
                              cursor: 'grab',
                              color: 'text.secondary',
                              mt: -0.25,
                              flexShrink: 0,
                              '&:active': { cursor: 'grabbing' },
                            }}
                            title="Drag to reorder"
                            aria-label="Drag to reorder"
                          >
                            <DragIndicatorIcon fontSize="small" />
                          </IconButton>
                          <Box sx={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 0.75 }}>
                            {!hideEmptyJsonLabel && (
                              <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600 }}>
                                {label}:
                              </Typography>
                            )}
                            {renderDetailFieldControl(field, controlType)}
                          </Box>
                        </Box>
                      </Paper>
                    );
                  })}
                </Box>
            )}
            {detailRow && !detailBusy && pinNoteSection && pinnedBottomFields.length > 0 && (
              <Box
                sx={{
                  px: 1.5,
                  pb: 2,
                  pt: 2,
                  borderTop: detailEditFieldList.length > 0 ? 1 : 0,
                  borderColor: 'divider',
                }}
              >
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2.5 }}>
                  {pinnedBottomFields.map((f) => (
                    <Box key={f}>
                      <Typography variant="subtitle2" fontWeight={700} sx={{ mb: 1.25 }}>
                        {f === 'detail' ? 'Detail (JSON)' : 'Description'}
                      </Typography>
                      {renderPinnedNoteBlock(f, false)}
                    </Box>
                  ))}
                </Box>
              </Box>
            )}
            {detailRow && !detailBusy && entityRouteForAttachments && expandedId != null && (
              <Box
                sx={{
                  px: 1.5,
                  pb: 2,
                  pt:
                    detailEditFieldList.length > 0 || (pinNoteSection && pinnedBottomFields.length > 0)
                      ? 2
                      : 0,
                  borderTop:
                    detailEditFieldList.length > 0 || (pinNoteSection && pinnedBottomFields.length > 0)
                      ? 1
                      : 0,
                  borderColor: 'divider',
                }}
              >
                <RecordAttachmentsPanel
                  entityRoute={entityRouteForAttachments}
                  recordId={expandedId}
                  attachments={detailRow.attachments}
                  onChange={(next) => setDetailRow((prev) => (prev ? { ...prev, attachments: next } : prev))}
                />
              </Box>
            )}
          </Box>
          <Box
            sx={{
              minHeight: 56,
              flexShrink: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              px: 2,
              borderTop: isDark ? '1px solid rgba(255,255,255,0.08)' : '1px solid rgba(0,0,0,0.12)',
              bgcolor: isDark ? DETAIL_PANEL_BG_DARK : 'background.paper',
            }}
          >
            <Box>
              {entityTypeForComments && expandedId != null && (
                <Badge
                  color="secondary"
                  variant="dot"
                  invisible={threadCommentCount === 0}
                  overlap="rectangular"
                  sx={{
                    '& .MuiBadge-badge': {
                      right: 10,
                      top: 10,
                      boxShadow: (t) => `0 0 0 2px ${t.palette.background.paper}`,
                    },
                  }}
                >
                  <Button
                    color="primary"
                    size="small"
                    startIcon={<CommentOutlinedIcon />}
                    onClick={() => setCommentsOpen(true)}
                    sx={{
                      fontWeight: 600,
                      ...(isDark && {
                        color: '#a5b4fc',
                        '& .MuiSvgIcon-root': { color: '#a5b4fc' },
                      }),
                    }}
                  >
                    Comments
                  </Button>
                </Badge>
              )}
            </Box>
            <Button
              color="primary"
              variant="contained"
              size="small"
              startIcon={<SaveOutlinedIcon />}
              onClick={() => setSaveConfirmOpen(true)}
              disabled={!canSaveDetail}
              sx={{
                fontWeight: 600,
                ...(isDark && {
                  color: '#060a12',
                  background: 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)',
                  '&:hover': {
                    background: 'linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%)',
                  },
                  '& .MuiSvgIcon-root': { color: '#060a12' },
                  '&.Mui-disabled': {
                    color: 'rgba(6, 10, 18, 0.5)',
                    background: 'rgba(34, 211, 238, 0.35)',
                  },
                }),
              }}
            >
              {saveBusy ? 'Saving…' : 'Save'}
            </Button>
          </Box>
        </Box>
      </Drawer>



      {/* Create record — wide right drawer so controls are not squeezed */}
      <Drawer
        anchor="right"
        open={createOpen}
        onClose={() => {
          if (createBusy) return;
          setCreateOpen(false);
        }}
        PaperProps={{
          sx: {
            width: { xs: '100%', sm: '66.666vw' },
            maxWidth: '100%',
            top: { xs: 0, sm: `calc(${ADMIN_RIGHT_PANEL_TOP}px + env(safe-area-inset-top, 0px))` },
            height: { xs: '100%', sm: adminRightPanelHeightCalc },
            maxHeight: { xs: '100%', sm: adminRightPanelHeightCalc },
            boxSizing: 'border-box',
            display: 'flex',
            flexDirection: 'column',
            bgcolor: 'background.default',
            pt: { xs: 'env(safe-area-inset-top)', sm: 0 },
            pb: { xs: 'env(safe-area-inset-bottom)', sm: 0 },
          },
        }}
        ModalProps={{ keepMounted: false, disableScrollLock: true }}
      >
        <Toolbar
          sx={{
            minHeight: 48,
            px: 1.5,
            py: 0.5,
            flexShrink: 0,
            borderBottom: isDark ? '1px solid rgba(255,255,255,0.08)' : '1px solid rgba(0,0,0,0.12)',
            bgcolor: 'background.paper',
            gap: 1,
          }}
        >
          <Typography variant="subtitle1" fontWeight={700} sx={{ flex: 1, minWidth: 0 }}>
            Create {title || 'Record'}
          </Typography>
          <IconButton
            size="small"
            onClick={() => !createBusy && setCreateOpen(false)}
            aria-label="Close create panel"
            disabled={createBusy}
          >
            <CloseIcon />
          </IconButton>
        </Toolbar>
        <Box
          sx={{
            flex: 1,
            minHeight: 0,
            overflow: 'auto',
            px: { xs: 1.5, sm: 2.5 },
            py: 2,
            borderBottom: isDark ? '1px solid rgba(255,255,255,0.08)' : '1px solid rgba(0,0,0,0.12)',
          }}
        >
          {createErr && (
            <Typography color="error" variant="body2" sx={{ mb: 1 }}>
              {createErr}
            </Typography>
          )}
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' },
              columnGap: { xs: 2, sm: 3 },
              rowGap: 2,
            }}
          >
            {createFieldList.map((field) => {
              const isReq = requiredFields.includes(field);
              const label = fieldLabel(field);
              const controlType = controlTypeForCreateField(field);
              return (
                <Box
                  key={field}
                  sx={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'stretch',
                    gap: 0.75,
                    minWidth: 0,
                  }}
                >
                  <Typography variant="body2" sx={{ fontWeight: 600, wordBreak: 'break-word' }}>
                    {label}
                    {isReq ? ' *' : ''}
                  </Typography>
                  {renderCreateFieldControl(field, controlType)}
                </Box>
              );
            })}
          </Box>
          {pinNoteSection && createPinnedBottomFields.length > 0 && (
            <Box sx={{ mt: 3, display: 'flex', flexDirection: 'column', gap: 2.5 }}>
              {createPinnedBottomFields.map((f) => (
                <Box key={f}>
                  <Typography variant="subtitle2" fontWeight={700} sx={{ mb: 1.25 }}>
                    {f === 'detail' ? 'Detail (JSON)' : 'Description'}
                  </Typography>
                  {renderPinnedNoteBlock(f, true)}
                </Box>
              ))}
            </Box>
          )}
        </Box>
        <Box
          sx={{
            flexShrink: 0,
            px: { xs: 1.5, sm: 2.5 },
            py: 1.5,
            display: 'flex',
            justifyContent: 'flex-end',
            gap: 1,
            bgcolor: 'background.paper',
          }}
        >
          <Button onClick={() => setCreateOpen(false)} disabled={createBusy}>
            Cancel
          </Button>
          <Button
            color="primary"
            variant="contained"
            onClick={submitCreate}
            disabled={!onCreate || createBusy}
            sx={{
              fontWeight: 600,
              ...(isDark && {
                color: '#060a12',
                background: 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)',
                '&:hover': {
                  background: 'linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%)',
                },
                '&.Mui-disabled': {
                  color: 'rgba(6, 10, 18, 0.5)',
                  background: 'rgba(34, 211, 238, 0.35)',
                },
              }),
            }}
          >
            Create
          </Button>
        </Box>
      </Drawer>

      {/* Delete confirmation */}
      <Dialog open={deleteConfirmOpen} onClose={() => setDeleteConfirmOpen(false)}>
        <DialogTitle>Delete record</DialogTitle>
        <DialogContent>
          <Typography>Are you sure you want to delete this record? This cannot be undone.</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteConfirmOpen(false)}>Cancel</Button>
          <Button
            color="error"
            variant="contained"
            onClick={submitDelete}
            disabled={deleteBusy}
            sx={{
              fontWeight: 600,
              ...(isDark && {
                color: '#ffebee',
                background: 'linear-gradient(135deg, #e53935 0%, #c62828 100%)',
                '&:hover': { background: 'linear-gradient(135deg, #ef5350 0%, #e53935 100%)' },
                '&.Mui-disabled': {
                  color: 'rgba(255, 235, 238, 0.5)',
                  background: 'rgba(229, 57, 53, 0.35)',
                },
              }),
            }}
          >
            OK
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={batchDeleteConfirmOpen} onClose={() => setBatchDeleteConfirmOpen(false)}>
        <DialogTitle>Delete {deleteTargets.length} records</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete {deleteTargets.length} selected records? This cannot be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setBatchDeleteConfirmOpen(false)}>Cancel</Button>
          <Button
            color="error"
            variant="contained"
            onClick={submitDelete}
            disabled={deleteBusy}
            sx={{
              fontWeight: 600,
              ...(isDark && {
                color: '#ffebee',
                background: 'linear-gradient(135deg, #e53935 0%, #c62828 100%)',
                '&:hover': { background: 'linear-gradient(135deg, #ef5350 0%, #e53935 100%)' },
                '&.Mui-disabled': {
                  color: 'rgba(255, 235, 238, 0.5)',
                  background: 'rgba(229, 57, 53, 0.35)',
                },
              }),
            }}
          >
            Delete {deleteTargets.length}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Save confirmation */}
      <Dialog open={saveConfirmOpen} onClose={() => setSaveConfirmOpen(false)}>
        <DialogTitle>Save changes</DialogTitle>
        <DialogContent>
          <Typography>Save your changes to this record?</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSaveConfirmOpen(false)}>Cancel</Button>
          <Button
            color="primary"
            variant="contained"
            onClick={submitSave}
            disabled={saveBusy}
            sx={{
              fontWeight: 600,
              ...(isDark && {
                color: '#060a12',
                background: 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)',
                '&:hover': {
                  background: 'linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%)',
                },
                '&.Mui-disabled': {
                  color: 'rgba(6, 10, 18, 0.5)',
                  background: 'rgba(34, 211, 238, 0.35)',
                },
              }),
            }}
          >
            OK
          </Button>
        </DialogActions>
      </Dialog>

      <AdvancedFilterModal
        open={filterModalOpen}
        onClose={() => setFilterModalOpen(false)}
        gridKey={storageKey}
        columns={baseGridColumns}
        rules={filterRules}
        setRules={setFilterRules}
      />

      <ColorTagsModal
        open={colorTagsOpen}
        onClose={() => setColorTagsOpen(false)}
        gridKey={storageKey}
        columns={baseGridColumns}
        config={colorConfig}
        onSaved={(cfg) => setColorConfig(normalizeColorConfig(cfg))}
      />

      {/* Record sheet (read-only snapshot; Export downloads PDF matching this view) */}
      <Dialog
        open={printModalOpen}
        onClose={closePrintSheet}
        maxWidth="md"
        fullWidth
        scroll="paper"
      >
        <DialogTitle sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1, pr: 1 }}>
          <Typography component="span" variant="h6" fontWeight={700}>
            Record sheet
          </Typography>
          <IconButton size="small" onClick={closePrintSheet} aria-label="Close">
            <CloseIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          {printBusy && (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress size={32} />
            </Box>
          )}
          {!printBusy && printRow && (
            <Box
              ref={recordSheetPdfRef}
              className="record-print-sheet"
              sx={{
                fontFamily: '"Palatino Linotype", Palatino, "Iowan Old Style", "Book Antiqua", Georgia, serif',
                fontSize: '11pt',
                lineHeight: 1.5,
                color: 'text.primary',
                px: { xs: 0, sm: 1 },
              }}
            >
              <Box
                sx={{
                  borderBottom: '2px solid',
                  borderColor: 'primary.main',
                  pb: 1.5,
                  mb: 2.5,
                }}
              >
                <Typography
                  variant="h5"
                  sx={{
                    fontFamily: 'inherit',
                    fontSize: '1.35rem',
                    fontWeight: 700,
                    letterSpacing: '0.02em',
                    m: 0,
                  }}
                >
                  {getRecordDisplayName(printRow, title, printRecordId)}
                </Typography>
                <Typography variant="body2" sx={{ mt: 0.5, opacity: 0.85, fontFamily: 'inherit' }}>
                  {title || 'Table'}
                  {printRecordId != null ? `  ·  ID ${printRecordId}` : ''}
                </Typography>
                <Typography variant="caption" sx={{ display: 'block', mt: 0.75, opacity: 0.7, fontFamily: 'inherit' }}>
                  Generated {new Date().toLocaleString()}
                </Typography>
              </Box>

              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                {printFieldKeys.map((field) => {
                  const label = fieldLabel(field);
                  const ct = controlTypeForField(field);
                  const raw = printRow[field];
                  const isLong =
                    ct === 'multiline' ||
                    ct === 'long_text' ||
                    ct === 'json_detail' ||
                    ct === 'location_map' ||
                    isPinnedBottomField(field);

                  if (ct === 'json_detail') {
                    const labelEl = (
                      <Typography
                        variant="overline"
                        sx={{
                          display: 'block',
                          fontWeight: 700,
                          letterSpacing: '0.08em',
                          mb: isPinnedBottomField(field) ? 1 : 0.75,
                          fontFamily: 'inherit',
                        }}
                      >
                        {label}
                      </Typography>
                    );
                    const inner = (
                      <>
                        {labelEl}
                        <RecordSheetJsonDetail raw={raw} />
                      </>
                    );
                    if (isPinnedBottomField(field)) {
                      return (
                        <Box
                          key={field}
                          sx={{
                            breakInside: 'avoid',
                            pageBreakInside: 'avoid',
                            border: '1px solid',
                            borderColor: 'divider',
                            borderRadius: 1,
                            p: 2,
                            bgcolor: 'action.hover',
                          }}
                        >
                          {inner}
                        </Box>
                      );
                    }
                    return (
                      <Box key={field} sx={{ breakInside: 'avoid', pageBreakInside: 'avoid' }}>
                        {inner}
                      </Box>
                    );
                  }

                  if (ct === 'location_map') {
                    return (
                      <Box key={field} sx={{ breakInside: 'avoid', pageBreakInside: 'avoid' }}>
                        <Typography
                          variant="overline"
                          sx={{
                            display: 'block',
                            fontWeight: 700,
                            letterSpacing: '0.08em',
                            mb: 0.75,
                            fontFamily: 'inherit',
                          }}
                        >
                          {label}
                        </Typography>
                        <RecordSheetActivityLocation raw={raw} />
                      </Box>
                    );
                  }

                  if (isPinnedBottomField(field) && typeof raw === 'string') {
                    return (
                      <Box
                        key={field}
                        sx={{
                          breakInside: 'avoid',
                          pageBreakInside: 'avoid',
                          border: '1px solid',
                          borderColor: 'divider',
                          borderRadius: 1,
                          p: 2,
                          bgcolor: 'action.hover',
                        }}
                      >
                        <Typography
                          variant="overline"
                          sx={{ display: 'block', fontWeight: 700, letterSpacing: '0.08em', mb: 1, fontFamily: 'inherit' }}
                        >
                          {label}
                        </Typography>
                        <Typography component="pre" variant="body2" sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', m: 0, fontFamily: 'ui-monospace, monospace', fontSize: 12 }}>
                          {raw || '—'}
                        </Typography>
                      </Box>
                    );
                  }

                  const textVal = formatPrintFieldBody(
                    field,
                    raw,
                    normalizedDetailJsonFieldKeys,
                    normalizedLocationJsonFieldKeys
                  );
                  if (isLong && String(textVal).length > 80) {
                    return (
                      <Box key={field} sx={{ breakInside: 'avoid', pageBreakInside: 'avoid' }}>
                        <Typography
                          variant="overline"
                          sx={{ display: 'block', fontWeight: 700, letterSpacing: '0.08em', mb: 0.75, fontFamily: 'inherit' }}
                        >
                          {label}
                        </Typography>
                        <Typography
                          component="div"
                          variant="body2"
                          sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', fontFamily: 'inherit' }}
                        >
                          {textVal}
                        </Typography>
                      </Box>
                    );
                  }

                  return (
                    <Box
                      key={field}
                      sx={{
                        display: 'grid',
                        gridTemplateColumns: { xs: '1fr', sm: 'minmax(130px, 190px) 1fr' },
                        gap: { xs: 0.5, sm: 2 },
                        alignItems: 'baseline',
                        breakInside: 'avoid',
                        pageBreakInside: 'avoid',
                        py: 0.5,
                        borderBottom: '1px dotted',
                        borderColor: 'divider',
                      }}
                    >
                      <Typography
                        variant="body2"
                        sx={{
                          fontWeight: 700,
                          color: 'text.secondary',
                          textAlign: { xs: 'left', sm: 'right' },
                          pr: { sm: 1 },
                          fontFamily: 'inherit',
                        }}
                      >
                        {label}
                      </Typography>
                      <Typography variant="body2" sx={{ fontFamily: 'inherit', wordBreak: 'break-word' }}>
                        {textVal}
                      </Typography>
                    </Box>
                  );
                })}
              </Box>

            </Box>
          )}
        </DialogContent>
        <DialogActions sx={{ px: 3, py: 2, gap: 1 }}>
          <Button onClick={closePrintSheet} color="inherit">
            Close
          </Button>
          <Button
            variant="contained"
            startIcon={
              sheetExportBusy ? <CircularProgress size={18} thickness={5} color="inherit" /> : <FileDownloadOutlinedIcon />
            }
            onClick={() => void exportRecordSheetPdf()}
            disabled={printBusy || sheetExportBusy || !printRow}
          >
            {sheetExportBusy ? 'Exporting…' : 'Export'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Comments modal */}
      {entityTypeForComments && (
        <CommentsModal
          open={commentsOpen}
          onClose={() => setCommentsOpen(false)}
          entityType={entityTypeForComments}
          recordId={expandedId}
          recordLabel={detailRow ? getRecordDisplayName(detailRow, title, expandedId) : null}
          onThreadChanged={handleCommentThreadChanged}
        />
      )}

      <Snackbar
        open={saveSuccessOpen}
        autoHideDuration={4000}
        onClose={(_, reason) => {
          if (reason === 'clickaway') return;
          setSaveSuccessOpen(false);
        }}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={() => setSaveSuccessOpen(false)}
          severity="success"
          variant="filled"
          sx={{ width: '100%' }}
        >
          Saved
        </Alert>
      </Snackbar>
    </Box>
  );
}
