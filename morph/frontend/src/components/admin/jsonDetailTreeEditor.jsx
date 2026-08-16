import React, { useCallback, useState } from 'react';
import {
  Box,
  Typography,
  TextField,
  IconButton,
  Button,
  Collapse,
  Stack,
  MenuItem,
  Select,
  FormControl,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import { JSON_DETAIL_MAX_NESTING } from './jsonDetailValidate';
import { useDetailJsonSurfaceSx } from './detailViewTheme';

/** Shared column widths so property rows, values, and delete line up across the editor. */
const GRID_KEY = 'minmax(200px, 260px)';
const GRID_TYPE = '120px';
const GRID_VAL = 'minmax(0, 1fr)';
const GRID_DEL = '44px';
const gridObjectFieldRow = { xs: '1fr', sm: `${GRID_KEY} ${GRID_TYPE} ${GRID_VAL} ${GRID_DEL}` };
const gridPrimitiveRow = gridObjectFieldRow;
const gridContainerRow = { xs: '1fr', sm: `${GRID_KEY} minmax(0, 1fr)` };

/** Single-line control height; multiline text fields use exactly 2× this. */
const FIELD_UNIT = 40;

const compactFieldSx = {
  '& .MuiInputBase-root': {
    minHeight: FIELD_UNIT,
    height: FIELD_UNIT,
    boxSizing: 'border-box',
  },
};

const multilineFieldSx = {
  '& .MuiInputBase-root': {
    minHeight: FIELD_UNIT * 2,
    height: FIELD_UNIT * 2,
    alignItems: 'flex-start',
    boxSizing: 'border-box',
  },
  '& textarea.MuiInputBase-inputMultiline': {
    minHeight: '100% !important',
    height: '100% !important',
    overflow: 'auto',
    resize: 'vertical',
  },
};

function formatKeyLabel(key) {
  let s = String(key).trim().replace(/_/g, ' ');
  s = s.replace(/([a-z0-9])([A-Z])/g, '$1 $2');
  s = s.replace(/\s+/g, ' ').trim();
  s = s.replace(/\b\w/g, (c) => c.toUpperCase());
  return s.replace(/\bId\b/g, 'ID');
}

function uniqueObjectKey(o, base = 'field') {
  if (!Object.prototype.hasOwnProperty.call(o, base)) return base;
  let i = 1;
  while (Object.prototype.hasOwnProperty.call(o, `${base}_${i}`)) i += 1;
  return `${base}_${i}`;
}

/** @param {'string'|'number'|'boolean'|'null'} t */
function defaultForType(t) {
  if (t === 'number') return 0;
  if (t === 'boolean') return false;
  if (t === 'null') return null;
  return '';
}

function parseNumberInput(s) {
  const t = String(s).trim();
  if (t === '') return 0;
  const n = Number(t);
  return Number.isNaN(n) ? t : n;
}

function TreeRowChrome({ children, dense }) {
  return (
    <Box
      sx={{
        py: dense ? 0.25 : 0.5,
      }}
    >
      {children}
    </Box>
  );
}

function getPrimitiveType(value) {
  if (value === null) return 'null';
  if (typeof value === 'boolean') return 'boolean';
  if (typeof value === 'number') return 'number';
  return 'string';
}

function PrimitiveTypeSelect({ value, onCommit }) {
  const type = getPrimitiveType(value);
  return (
    <FormControl size="small" fullWidth sx={compactFieldSx}>
      <Select
        value={type}
        onChange={(e) => {
          onCommit(defaultForType(e.target.value));
        }}
      >
        <MenuItem value="string">Text</MenuItem>
        <MenuItem value="number">Number</MenuItem>
        <MenuItem value="boolean">Yes / No</MenuItem>
        <MenuItem value="null">Null</MenuItem>
      </Select>
    </FormControl>
  );
}

function PrimitiveValueControl({ value, onCommit }) {
  const type = getPrimitiveType(value);
  const [str, setStr] = useState(() => (typeof value === 'string' ? value : value == null ? '' : String(value)));

  React.useEffect(() => {
    if (typeof value === 'string') setStr(value);
    else if (value === null) setStr('');
    else setStr(String(value));
  }, [value]);

  if (type === 'string') {
    return (
      <TextField
        size="small"
        fullWidth
        multiline
        value={str}
        onChange={(e) => setStr(e.target.value)}
        onBlur={() => onCommit(str)}
        sx={multilineFieldSx}
      />
    );
  }
  if (type === 'number') {
    return (
      <TextField
        size="small"
        fullWidth
        type="text"
        inputProps={{ inputMode: 'decimal' }}
        value={str}
        onChange={(e) => setStr(e.target.value)}
        onBlur={() => onCommit(parseNumberInput(str))}
        sx={compactFieldSx}
      />
    );
  }
  if (type === 'boolean') {
    return (
      <FormControl size="small" fullWidth sx={compactFieldSx}>
        <Select value={value ? 'true' : 'false'} onChange={(e) => onCommit(e.target.value === 'true')}>
          <MenuItem value="true">Yes</MenuItem>
          <MenuItem value="false">No</MenuItem>
        </Select>
      </FormControl>
    );
  }
  return (
    <Typography
      variant="body2"
      color="text.secondary"
      sx={{ display: 'flex', alignItems: 'center', minHeight: FIELD_UNIT, height: FIELD_UNIT, px: 0.5 }}
    >
      null
    </Typography>
  );
}

function PrimitiveValueEditor({ value, onCommit }) {
  return (
    <Box
      sx={{
        display: 'grid',
        gridTemplateColumns: { xs: '1fr', sm: `${GRID_TYPE} minmax(0, 1fr)` },
        alignItems: 'start',
        columnGap: 1.5,
        rowGap: 0.75,
        width: '100%',
      }}
    >
      <PrimitiveTypeSelect value={value} onCommit={onCommit} />
      <PrimitiveValueControl value={value} onCommit={onCommit} />
    </Box>
  );
}

function DeleteFieldButton({ onClick, label = 'Remove field' }) {
  return (
    <IconButton
      size="small"
      color="error"
      aria-label={label}
      onClick={onClick}
      sx={{ justifySelf: 'center', alignSelf: 'center' }}
    >
      <DeleteOutlineIcon fontSize="small" />
    </IconButton>
  );
}

function TreeValue({
  value,
  depth,
  maxNesting,
  onReplace,
  label,
  onRemove,
  dense,
  /** When the key is already shown beside this control (object rows), only show counts here. */
  omitKeyFromHeader,
}) {
  const [expanded, setExpanded] = useState(true);
  const jsonRowSx = useDetailJsonSurfaceSx({ nested: true });

  if (value !== null && typeof value === 'object') {
    const isArr = Array.isArray(value);
    const Icon = expanded ? ExpandMoreIcon : ChevronRightIcon;
    return (
      <TreeRowChrome dense={dense}>
        <Stack direction="row" alignItems="flex-start" spacing={0.5} sx={{ width: '100%' }}>
          <IconButton size="small" onClick={() => setExpanded((e) => !e)} sx={{ mt: 0.125 }}>
            <Icon fontSize="small" />
          </IconButton>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Stack direction="row" alignItems="center" justifyContent="space-between" gap={1}>
              <Typography variant="caption" color="primary" fontWeight={700}>
                {omitKeyFromHeader
                  ? isArr
                    ? `List · ${value.length}`
                    : `Object · ${Object.keys(value).length}`
                  : `${label} ${isArr ? `· list (${value.length})` : `· object (${Object.keys(value).length})`}`}
              </Typography>
              {onRemove ? (
                <IconButton size="small" aria-label="Remove" onClick={onRemove} color="error">
                  <DeleteOutlineIcon fontSize="small" />
                </IconButton>
              ) : null}
            </Stack>
            <Collapse in={expanded}>
              <Box sx={{ pt: 0.75 }}>
                {isArr ? (
                  <TreeArrayInner value={value} depth={depth} maxNesting={maxNesting} onReplace={onReplace} />
                ) : (
                  <TreeObjectInner value={value} depth={depth} maxNesting={maxNesting} onReplace={onReplace} />
                )}
              </Box>
            </Collapse>
          </Box>
        </Stack>
      </TreeRowChrome>
    );
  }

  return (
    <Box
      sx={{
        display: 'grid',
        gridTemplateColumns: gridPrimitiveRow,
        columnGap: 1.5,
        rowGap: 0.5,
        alignItems: 'start',
        width: '100%',
        py: 1,
        px: 1.25,
        borderRadius: 1,
        ...jsonRowSx,
      }}
    >
      <Typography variant="caption" color="text.secondary" fontWeight={600} sx={{ alignSelf: 'start', pr: 0.5, pt: 0.75 }}>
        {label}
      </Typography>
      <PrimitiveTypeSelect value={value} onCommit={onReplace} />
      <Box sx={{ minWidth: 0, alignSelf: 'start' }}>
        <PrimitiveValueControl value={value} onCommit={onReplace} />
      </Box>
      {onRemove ? <DeleteFieldButton onClick={onRemove} label="Remove" /> : <Box />}
    </Box>
  );
}

function ObjectKeyField({ objKey, onRenameCommitted }) {
  const [local, setLocal] = useState(objKey);
  React.useEffect(() => setLocal(objKey), [objKey]);
  return (
    <TextField
      size="small"
      placeholder="Property name"
      value={local}
      onChange={(e) => setLocal(e.target.value)}
      onBlur={() => {
        const t = String(local || '').trim();
        if (!t) setLocal(objKey);
        else onRenameCommitted(t);
      }}
      fullWidth
      InputLabelProps={{ shrink: false }}
      sx={compactFieldSx}
    />
  );
}

function TreeObjectInner({ value, depth, maxNesting, onReplace }) {
  const canNest = depth < maxNesting;
  const keys = Object.keys(value);
  const jsonRowSx = useDetailJsonSurfaceSx({ nested: true });

  const renameKey = (oldKey, newKeyRaw) => {
    const newKey = String(newKeyRaw || '').trim();
    if (!newKey || newKey === oldKey) return;
    if (Object.prototype.hasOwnProperty.call(value, newKey)) return;
    const next = { ...value };
    next[newKey] = next[oldKey];
    delete next[oldKey];
    onReplace(next);
  };

  const setChild = (k, v) => {
    onReplace({ ...value, [k]: v });
  };

  const removeKey = (k) => {
    const next = { ...value };
    delete next[k];
    onReplace(next);
  };

  const addField = (kind) => {
    const k = uniqueObjectKey(value);
    if (kind === 'string') onReplace({ ...value, [k]: '' });
    else if (kind === 'number') onReplace({ ...value, [k]: 0 });
    else if (kind === 'boolean') onReplace({ ...value, [k]: false });
    else if (kind === 'null') onReplace({ ...value, [k]: null });
    else if (kind === 'object' && canNest) onReplace({ ...value, [k]: {} });
    else if (kind === 'array' && canNest) onReplace({ ...value, [k]: [] });
  };

  return (
    <Stack spacing={1.25}>
      {keys.length === 0 ? (
        <Typography variant="body2" color="text.secondary">
          No properties yet — add one below.
        </Typography>
      ) : (
        keys.map((k) => {
          const v = value[k];
          const isContainer = v !== null && typeof v === 'object';
          if (isContainer) {
            return (
              <Box
                key={k}
                sx={{
                  display: 'grid',
                  gridTemplateColumns: gridContainerRow,
                  columnGap: 1.5,
                  rowGap: 0,
                  alignItems: 'start',
                  py: 1,
                  px: 1.25,
                  borderRadius: 1,
                  ...jsonRowSx,
                }}
              >
                <ObjectKeyField objKey={k} onRenameCommitted={(nextRaw) => renameKey(k, nextRaw)} />
                <Box sx={{ minWidth: 0, alignSelf: 'stretch' }}>
                  <TreeValue
                    label={formatKeyLabel(k)}
                    value={v}
                    depth={depth + 1}
                    maxNesting={maxNesting}
                    onReplace={(next) => setChild(k, next)}
                    onRemove={() => removeKey(k)}
                    dense
                    omitKeyFromHeader
                  />
                </Box>
              </Box>
            );
          }
          return (
            <Box
              key={k}
              sx={{
                display: 'grid',
                gridTemplateColumns: gridObjectFieldRow,
                columnGap: 1.5,
                alignItems: 'start',
                py: 1,
                px: 1.25,
                borderRadius: 1,
                ...jsonRowSx,
              }}
            >
              <ObjectKeyField objKey={k} onRenameCommitted={(nextRaw) => renameKey(k, nextRaw)} />
              <PrimitiveTypeSelect value={v} onCommit={(next) => setChild(k, next)} />
              <Box sx={{ minWidth: 0, alignSelf: 'start' }}>
                <PrimitiveValueControl value={v} onCommit={(next) => setChild(k, next)} />
              </Box>
              <DeleteFieldButton onClick={() => removeKey(k)} />
            </Box>
          );
        })
      )}
      <Stack direction="row" flexWrap="wrap" gap={0.75}>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} onClick={() => addField('string')}>
          Text
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} onClick={() => addField('number')}>
          Number
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} onClick={() => addField('boolean')}>
          Yes/No
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} onClick={() => addField('null')}>
          Null
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} disabled={!canNest} onClick={() => addField('object')}>
          Object
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} disabled={!canNest} onClick={() => addField('array')}>
          List
        </Button>
      </Stack>
      {!canNest ? (
        <Typography variant="caption" color="text.secondary">
          Nesting limit reached ({maxNesting} layers). Add text or numbers only, or remove nesting.
        </Typography>
      ) : null}
    </Stack>
  );
}

function TreeArrayInner({ value, depth, maxNesting, onReplace }) {
  const canNest = depth < maxNesting;

  const setAt = (i, nextItem) => {
    const copy = [...value];
    copy[i] = nextItem;
    onReplace(copy);
  };

  const removeAt = (i) => {
    onReplace(value.filter((_, j) => j !== i));
  };

  const pushKind = (kind) => {
    if (kind === 'string') onReplace([...value, '']);
    else if (kind === 'number') onReplace([...value, 0]);
    else if (kind === 'boolean') onReplace([...value, false]);
    else if (kind === 'null') onReplace([...value, null]);
    else if (kind === 'object' && canNest) onReplace([...value, {}]);
    else if (kind === 'array' && canNest) onReplace([...value, []]);
  };

  return (
    <Stack spacing={1.25}>
      {value.length === 0 ? (
        <Typography variant="body2" color="text.secondary">
          Empty list — add an item below.
        </Typography>
      ) : (
        value.map((item, i) => (
          <Box key={i} sx={{ pl: 0.5, borderLeft: '2px solid', borderColor: 'divider' }}>
            <TreeValue
              label={`Item ${i + 1}`}
              value={item}
              depth={depth + 1}
              maxNesting={maxNesting}
              onReplace={(next) => setAt(i, next)}
              onRemove={() => removeAt(i)}
              dense
            />
          </Box>
        ))
      )}
      <Stack direction="row" flexWrap="wrap" gap={0.75}>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} onClick={() => pushKind('string')}>
          Text
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} onClick={() => pushKind('number')}>
          Number
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} onClick={() => pushKind('boolean')}>
          Yes/No
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} onClick={() => pushKind('null')}>
          Null
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} disabled={!canNest} onClick={() => pushKind('object')}>
          Object
        </Button>
        <Button size="small" variant="outlined" startIcon={<AddIcon />} disabled={!canNest} onClick={() => pushKind('array')}>
          List
        </Button>
      </Stack>
    </Stack>
  );
}

/**
 * Tree editor for parsed JSON (object, array, or primitive). Commits pretty-printed JSON string via onChange.
 */
export function JsonDetailTreeEditor({
  parsed,
  onChangeString,
  maxNesting = JSON_DETAIL_MAX_NESTING,
}) {
  const commit = useCallback(
    (next) => {
      onChangeString(JSON.stringify(next, null, 2));
    },
    [onChangeString]
  );

  if (parsed === undefined || parsed === null || typeof parsed !== 'object') {
    return (
      <Stack spacing={1}>
        <Typography variant="caption" color="text.secondary">
          Single value (not an object). Adjust type and content below.
        </Typography>
        <PrimitiveValueEditor
          value={parsed === undefined ? null : parsed}
          onCommit={(v) => commit(v)}
        />
      </Stack>
    );
  }

  if (Array.isArray(parsed)) {
    return (
      <Box sx={{ pt: 0.5 }}>
        <TreeArrayInner value={parsed} depth={0} maxNesting={maxNesting} onReplace={commit} />
      </Box>
    );
  }

  return (
    <Box sx={{ pt: 0.5 }}>
      <TreeObjectInner value={parsed} depth={0} maxNesting={maxNesting} onReplace={commit} />
    </Box>
  );
}
