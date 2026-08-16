import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
  Typography,
} from '@mui/material';
import { MapContainer, TileLayer, Marker, useMapEvents } from 'react-leaflet';
import L from 'leaflet';
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png';
import markerIcon from 'leaflet/dist/images/marker-icon.png';
import markerShadow from 'leaflet/dist/images/marker-shadow.png';
import 'leaflet/dist/leaflet.css';
import {
  mergeActivityLocationJson,
  parseActivityLocationJson,
} from '../../utils/activityLocationJson';

const defaultIcon = L.icon({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
});
L.Marker.prototype.options.icon = defaultIcon;

const US_CENTER = [39.8283, -98.5795];

function MapClickHandler({ onPick }) {
  useMapEvents({
    click(e) {
      onPick(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

function LocationMapFrame({ lat, lng, height, interactive, onPick, mapKey }) {
  const hasPin = lat != null && lng != null && Number.isFinite(lat) && Number.isFinite(lng);
  const center = hasPin ? [lat, lng] : US_CENTER;
  const zoom = hasPin ? 14 : 4;

  return (
    <MapContainer
      key={mapKey}
      center={center}
      zoom={zoom}
      style={{ height, width: '100%', borderRadius: 8 }}
      scrollWheelZoom
    >
      <TileLayer
        crossOrigin
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      {hasPin && <Marker position={[lat, lng]} />}
      {interactive && <MapClickHandler onPick={onPick} />}
    </MapContainer>
  );
}

export function RecordSheetActivityLocation({ raw }) {
  const { description, lat, lng } = parseActivityLocationJson(raw);
  const hasPin = lat != null && lng != null;

  if (!description && !hasPin) {
    return (
      <Typography variant="body2" color="text.secondary" sx={{ fontFamily: 'inherit' }}>
        —
      </Typography>
    );
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.25 }}>
      {description ? (
        <Typography variant="body2" sx={{ fontFamily: 'inherit', whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
          {description}
        </Typography>
      ) : null}
      {hasPin ? (
        <Box
          sx={{
            height: 220,
            width: '100%',
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 1,
            overflow: 'hidden',
          }}
        >
          <LocationMapFrame
            lat={lat}
            lng={lng}
            height={220}
            interactive={false}
            mapKey={`sheet-${lat}-${lng}`}
          />
        </Box>
      ) : null}
    </Box>
  );
}

function LocationEditModal({ open, initialValue, onClose, onApply }) {
  const parsed = useMemo(() => parseActivityLocationJson(initialValue), [initialValue]);
  const [desc, setDesc] = useState('');
  const [latStr, setLatStr] = useState('');
  const [lngStr, setLngStr] = useState('');
  const [mapNonce, setMapNonce] = useState(0);

  useEffect(() => {
    if (!open) return;
    setDesc(parsed.description || '');
    setLatStr(parsed.lat != null ? String(parsed.lat) : '');
    setLngStr(parsed.lng != null ? String(parsed.lng) : '');
    setMapNonce((n) => n + 1);
  }, [open, parsed.description, parsed.lat, parsed.lng]);

  const latNum = parseFloat(latStr, 10);
  const lngNum = parseFloat(lngStr, 10);

  const handlePick = useCallback((la, ln) => {
    setLatStr(String(Math.round(la * 1e6) / 1e6));
    setLngStr(String(Math.round(ln * 1e6) / 1e6));
  }, []);

  const apply = () => {
    const merged = mergeActivityLocationJson(initialValue, {
      description: desc,
      lat: latStr,
      lng: lngStr,
    });
    onApply(merged);
    onClose();
  };

  const clearAll = () => {
    onApply(null);
    onClose();
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth scroll="body">
      <DialogTitle>Edit location</DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
          Click the map to set coordinates (latitude / longitude). Optionally add a location description; it is stored in
          the JSON <code>location</code> field.
        </Typography>
        <Box sx={{ height: 280, width: '100%', mb: 2, borderRadius: 1, overflow: 'hidden' }}>
          <LocationMapFrame
            lat={Number.isFinite(latNum) ? latNum : null}
            lng={Number.isFinite(lngNum) ? lngNum : null}
            height={280}
            interactive
            onPick={handlePick}
            mapKey={`edit-${mapNonce}`}
          />
        </Box>
        <TextField
          size="small"
          fullWidth
          label="Location description"
          value={desc}
          onChange={(e) => setDesc(e.target.value)}
          sx={{ mb: 1.5 }}
        />
        <Box sx={{ display: 'flex', gap: 1 }}>
          <TextField
            size="small"
            fullWidth
            label="Latitude (Y)"
            value={latStr}
            onChange={(e) => setLatStr(e.target.value)}
            inputProps={{ inputMode: 'decimal' }}
          />
          <TextField
            size="small"
            fullWidth
            label="Longitude (X)"
            value={lngStr}
            onChange={(e) => setLngStr(e.target.value)}
            inputProps={{ inputMode: 'decimal' }}
          />
        </Box>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2, gap: 1 }}>
        <Button onClick={clearAll} color="inherit">
          Clear
        </Button>
        <Box sx={{ flex: 1 }} />
        <Button type="button" onClick={onClose}>
          Cancel
        </Button>
        <Button type="button" variant="contained" onClick={apply}>
          Apply
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export function ActivityLocationFieldEditor({ value, onChange, disabled }) {
  const [modalOpen, setModalOpen] = useState(false);
  const { description, lat, lng } = parseActivityLocationJson(value);
  const hasPin = lat != null && lng != null;

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1, width: '100%', minWidth: 0 }}>
      {description || hasPin ? (
        <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-word' }}>
          {description || '(No description)'}
          {hasPin ? (
            <Box component="span" sx={{ display: 'block', mt: 0.5, fontVariantNumeric: 'tabular-nums' }}>
              {lat.toFixed(5)}, {lng.toFixed(5)}
            </Box>
          ) : null}
        </Typography>
      ) : (
        <Typography variant="body2" color="text.secondary">
          No location set
        </Typography>
      )}
      {hasPin ? (
        <Box
          sx={{
            height: 180,
            width: '100%',
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 1,
            overflow: 'hidden',
          }}
        >
          <LocationMapFrame
            lat={lat}
            lng={lng}
            height={180}
            interactive={false}
            mapKey={`editor-${lat}-${lng}`}
          />
        </Box>
      ) : null}
      <Button variant="outlined" size="small" disabled={disabled} onClick={() => setModalOpen(true)} sx={{ alignSelf: 'flex-start' }}>
        Edit location
      </Button>
      <LocationEditModal
        open={modalOpen}
        initialValue={value}
        onClose={() => setModalOpen(false)}
        onApply={(merged) => onChange(merged)}
      />
    </Box>
  );
}
