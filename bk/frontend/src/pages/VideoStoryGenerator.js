import React, { useState, useCallback } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Grid,
  IconButton,
  Alert,
  Chip,
  Stepper,
  Step,
  StepLabel,
  CircularProgress,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  FormControlLabel,
  Switch,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Download as DownloadIcon,
  ExpandMore as ExpandMoreIcon,
  Movie as MovieIcon,
  Image as ImageIcon,
  AutoFixHigh as PolishIcon,
  Visibility as PreviewIcon,
  Groups as CastIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import api, {
  getVideoStoryFileUrl,
  polishVideoStoryScenes,
  polishVideoStoryContent,
  generateVideoStoryImages,
  generateVideoStoryVideos,
} from '../services/api';
import ModuleShell from '../components/ModuleShell';

const PIPELINE_STEPS = ['Scenes', 'Polish', 'Cast & world', 'Videos', 'Review'];

const SCENE_STATUS_COLOR = {
  draft: 'default',
  prompt_polished: 'info',
  images_ready: 'warning',
  video_ready: 'success',
  failed: 'error',
};

const emptyScene = (order) => ({
  id: `new-${Date.now()}-${order}`,
  order,
  title: `Scene ${order}`,
  user_prompt: '',
  polished_prompt: '',
  character_descriptions: [],
  scenery_description: '',
  images: [],
  video_filename: null,
  status: 'draft',
  notes: '',
  error: null,
  duration_seconds: 5,
  continue_from_previous: order > 1,
  cut_note: '',
});

const formatApiError = (err, fallback = 'Request failed') => {
  const detail = err?.response?.data?.detail;
  if (typeof detail === 'string' && detail.trim()) return detail;
  if (Array.isArray(detail)) {
    const parts = detail.map((item) => {
      if (typeof item === 'string') return item;
      if (item && typeof item === 'object') {
        const loc = Array.isArray(item.loc) ? item.loc.filter((x) => x !== 'body').join('.') : '';
        const msg = item.msg || item.message || JSON.stringify(item);
        return loc ? `${loc}: ${msg}` : msg;
      }
      return String(item);
    });
    return parts.filter(Boolean).join('; ') || fallback;
  }
  if (detail && typeof detail === 'object') {
    return detail.message || detail.msg || JSON.stringify(detail);
  }
  return err?.message || fallback;
};

const castWorldCardsFromProject = (project) => {
  const cast = (project?.cast || []).map((m) => ({
    id: m.id,
    name: m.name || 'Character',
    filename: m?.image?.filename || null,
    description: m.canonical_description || '',
    is_primary: Boolean(m.is_primary),
    kind: 'character',
  }));
  const hasWorldMeta = Boolean(
    project?.world
    && ((project.world.canonical_description || '').trim() || project.world?.image?.filename),
  );
  const world = hasWorldMeta
    ? [{
      id: 'world',
      name: 'World',
      filename: project.world?.image?.filename || null,
      description: project.world?.canonical_description || '',
      is_primary: false,
      kind: 'scenery',
    }]
    : [];
  return [...cast, ...world];
};

const sharedAssetsFromProject = (project) => (
  castWorldCardsFromProject(project).filter((a) => a.filename)
);

const sceneHasPreview = (scene) => Boolean(scene?.video_filename);

const i2vReferenceHint = (scene) => {
  if (!scene) return '';
  if (scene.order > 1 && scene.continue_from_previous !== false) {
    return 'I2V: previous last frame (PNG), cast fallback if rejected';
  }
  return 'I2V starts from cast/world sheet';
};

const defaultImagePrompt = (kind, name, description) => {
  const desc = (description || '').trim();
  if (kind === 'scenery') {
    return desc
      ? `Environment / world establishing plate, no readable text, consistent series location. ${desc}`
      : '';
  }
  return desc
    ? `Character reference sheet, full body and clear face, neutral pose, consistent design bible, plain background. ${name || 'Character'}: ${desc}`
    : '';
};

const draftFromProject = (project) => ({
  style_bible: project?.style_bible || '',
  cast: (project?.cast || []).map((m) => {
    const description = m.canonical_description || '';
    const existingPrompt = (m.image?.prompt || '').trim();
    return {
      id: m.id,
      name: m.name || '',
      canonical_description: description,
      aliases: m.aliases || [],
      is_primary: Boolean(m.is_primary),
      image: {
        id: m.image?.id || `img-${m.id}`,
        asset_type: 'character',
        prompt: existingPrompt || defaultImagePrompt('character', m.name, description),
        filename: m.image?.filename || null,
        description: m.image?.description || m.name || '',
      },
    };
  }),
  world: {
    canonical_description: project?.world?.canonical_description || '',
    image: {
      id: project?.world?.image?.id || 'world-img',
      asset_type: 'scenery',
      prompt: (project?.world?.image?.prompt || '').trim()
        || defaultImagePrompt('scenery', 'World', project?.world?.canonical_description || ''),
      filename: project?.world?.image?.filename || null,
      description: 'World',
    },
  },
});

const VideoStoryGenerator = ({ suppressModuleHeader = false }) => {
  const queryClient = useQueryClient();
  const { data: projects = [], isLoading } = useQuery('videoStories', api.getVideoStories);

  const [selected, setSelected] = useState(null);
  const [scenes, setScenes] = useState([]);
  const [castDraft, setCastDraft] = useState(draftFromProject(null));
  const [activeStep, setActiveStep] = useState(0);
  const [openDialog, setOpenDialog] = useState(false);
  const [form, setForm] = useState({ title: '', description: '', story_context: '' });
  const [busy, setBusy] = useState('');
  const [error, setError] = useState('');
  const [previewScene, setPreviewScene] = useState(null);
  const [previewShared, setPreviewShared] = useState(false);
  const [expandedSceneId, setExpandedSceneId] = useState(null);
  const [expandedCastId, setExpandedCastId] = useState(null);
  const [storyDraft, setStoryDraft] = useState({ story_context: '', description: '' });

  const syncScenes = useCallback((project) => {
    if (!project) {
      setScenes([]);
      setCastDraft(draftFromProject(null));
      return;
    }
    const nextScenes = (project.scenes || []).map((s, i) => ({
      ...s,
      order: s.order || i + 1,
      duration_seconds: s.duration_seconds ?? 5,
      continue_from_previous: s.continue_from_previous !== false && (s.order || i + 1) > 1,
      cut_note: s.cut_note || '',
    }));
    setScenes(nextScenes);
    setCastDraft(draftFromProject(project));
    const shared = sharedAssetsFromProject(project);
    const hasVideos = nextScenes.some((s) => s.status === 'video_ready' || s.video_filename);
    const hasShared = shared.length > 0;
    const hasPolish = nextScenes.some((s) => s.polished_prompt);
    // Steps: 0 Scenes, 1 Polish, 2 Cast & world, 3 Videos, 4 Review
    if (hasVideos) setActiveStep(4);
    else if (hasShared) setActiveStep(3);
    else if (hasPolish) setActiveStep(2);
    else if (nextScenes.length) setActiveStep(1);
    else setActiveStep(0);
  }, []);

  const selectProject = (p) => {
    setSelected(p);
    syncScenes(p);
    setStoryDraft({
      story_context: p?.story_context || '',
      description: p?.description || '',
    });
    setError('');
    setPreviewScene(null);
    setPreviewShared(false);
    setExpandedSceneId((p?.scenes || [])[0]?.id || null);
    setExpandedCastId((p?.cast || [])[0]?.id || 'world');
  };

  const createMutation = useMutation(api.createVideoStory, {
    onSuccess: async (res) => {
      queryClient.invalidateQueries('videoStories');
      setOpenDialog(false);
      setForm({ title: '', description: '', story_context: '' });
      const projectId = res?.id;
      if (!projectId) {
        setError('Project created but no id was returned');
        return;
      }
      const refreshed = await api.getVideoStory(projectId);
      selectProject(refreshed);
    },
    onError: (err) => {
      setError(formatApiError(err, 'Failed to create project'));
    },
  });

  const buildUpdatePayload = (project) => ({
    title: project.title,
    description: storyDraft.description ?? project.description ?? '',
    story_context: storyDraft.story_context ?? project.story_context ?? '',
    scenes: scenes.map((s, i) => ({
      ...s,
      id: s.id && !String(s.id).startsWith('new-') ? s.id : `scene-${i + 1}-${Date.now()}`,
      order: i + 1,
      title: s.title || `Scene ${i + 1}`,
      user_prompt: s.user_prompt || '',
      character_descriptions: s.character_descriptions || [],
      images: s.images || [],
      status: s.status || 'draft',
      notes: s.notes || '',
      duration_seconds: Number(s.duration_seconds) > 0 ? Number(s.duration_seconds) : 5,
      continue_from_previous: s.order > 1 ? s.continue_from_previous !== false : false,
      cut_note: s.cut_note || '',
    })),
    cast: (castDraft.cast || []).map((m) => ({
      id: m.id,
      name: m.name || '',
      canonical_description: m.canonical_description || '',
      aliases: m.aliases || [],
      is_primary: Boolean(m.is_primary),
      image: m.image ? {
        id: m.image.id,
        asset_type: 'character',
        prompt: m.image.prompt || '',
        filename: m.image.filename || null,
        description: m.image.description || m.name || '',
      } : null,
    })),
    world: {
      canonical_description: castDraft.world?.canonical_description || '',
      image: {
        id: castDraft.world?.image?.id || 'world-img',
        asset_type: 'scenery',
        prompt: castDraft.world?.image?.prompt || '',
        filename: castDraft.world?.image?.filename || null,
        description: 'World',
      },
    },
    style_bible: castDraft.style_bible || '',
  });

  const saveScenes = async (projectArg) => {
    const project = projectArg && typeof projectArg === 'object' && projectArg.id ? projectArg : selected;
    if (!project?.id) {
      setError('Select or create a project before saving');
      return null;
    }
    setBusy('save');
    setError('');
    try {
      await api.updateVideoStory(project.id, buildUpdatePayload(project));
      const refreshed = await api.getVideoStory(project.id);
      setSelected(refreshed);
      syncScenes(refreshed);
      queryClient.invalidateQueries('videoStories');
      return refreshed;
    } catch (err) {
      setError(formatApiError(err, 'Failed to save scenes'));
      return null;
    } finally {
      setBusy('');
    }
  };

  const saveCastWorld = async () => {
    if (!selected?.id) return null;
    setBusy('save-cast');
    setError('');
    try {
      await api.updateVideoStory(selected.id, buildUpdatePayload(selected));
      const refreshed = await api.getVideoStory(selected.id);
      setSelected(refreshed);
      syncScenes(refreshed);
      queryClient.invalidateQueries('videoStories');
      return refreshed;
    } catch (err) {
      setError(formatApiError(err, 'Failed to save cast & world'));
      return null;
    } finally {
      setBusy('');
    }
  };

  const ensureSaved = async () => {
    if (!selected) return null;
    return saveScenes(selected);
  };

  const updateCastMember = (memberId, patch) => {
    setCastDraft((prev) => ({
      ...prev,
      cast: (prev.cast || []).map((m) => {
        if (m.id !== memberId) return m;
        const next = { ...m, ...patch };
        if (patch.canonical_description !== undefined && !(m.image?.prompt || '').trim()) {
          next.image = {
            ...(m.image || {}),
            prompt: defaultImagePrompt('character', next.name, patch.canonical_description),
          };
        }
        if (patch.image) {
          next.image = { ...(m.image || {}), ...patch.image };
        }
        return next;
      }),
    }));
  };

  const updateWorldDraft = (patch) => {
    setCastDraft((prev) => {
      const world = { ...(prev.world || {}), ...patch };
      if (patch.canonical_description !== undefined && !(prev.world?.image?.prompt || '').trim()) {
        world.image = {
          ...(prev.world?.image || {}),
          prompt: defaultImagePrompt('scenery', 'World', patch.canonical_description),
        };
      }
      if (patch.image) {
        world.image = { ...(prev.world?.image || {}), ...patch.image };
      }
      return { ...prev, world };
    });
  };

  const resolveSceneId = (project, sceneId) => {
    if (!sceneId || !project) return sceneId;
    if (project.scenes.some((s) => s.id === sceneId)) return sceneId;
    const local = scenes.find((s) => s.id === sceneId);
    if (local?.order) {
      const match = project.scenes.find((s) => s.order === local.order);
      return match?.id || sceneId;
    }
    return sceneId;
  };

  const runPolish = async (sceneId = null) => {
    const project = await ensureSaved();
    if (!project) return;
    setBusy('polish');
    setError('');
    try {
      const result = await polishVideoStoryScenes(project.id, {
        scene_id: sceneId ? resolveSceneId(project, sceneId) : null,
        polish_all: !sceneId,
      });
      setSelected(result);
      syncScenes(result);
      setActiveStep(2);
      queryClient.invalidateQueries('videoStories');
    } catch (err) {
      setError(formatApiError(err, 'Failed to AI-polish scene prompts'));
    } finally {
      setBusy('');
    }
  };

  const runPolishContent = async ({ field, castMemberId, sourceText }) => {
    if (!selected?.id) return;
    const project = await saveCastWorld();
    if (!project) return;
    const busyKey = castMemberId ? `polish:${field}:${castMemberId}` : `polish:${field}`;
    setBusy(busyKey);
    setError('');
    try {
      const result = await polishVideoStoryContent(project.id, {
        field,
        cast_member_id: castMemberId || null,
        source_text: sourceText ?? undefined,
      });
      setSelected(result);
      syncScenes(result);
      setStoryDraft({
        story_context: result.story_context || '',
        description: result.description || '',
      });
      setCastDraft(draftFromProject(result));
      queryClient.invalidateQueries('videoStories');
    } catch (err) {
      setError(formatApiError(err, 'Failed to AI-polish content'));
    } finally {
      setBusy('');
    }
  };

  const runCastAndWorld = async (opts = {}) => {
    const {
      forceRegenerate = false,
      forceBible = false,
      castMemberId = null,
      regenWorld = false,
      openPreviewOnDone = true,
    } = opts;
    // Persist edited prompts/descriptions before regenerating images.
    const project = await saveCastWorld();
    if (!project) return;
    setBusy(castMemberId ? `cast:${castMemberId}` : regenWorld ? 'world' : 'images');
    setError('');
    try {
      const result = await generateVideoStoryImages(project.id, {
        generate_all: true,
        include_characters: true,
        include_scenery: true,
        force_regenerate: forceRegenerate,
        force_bible: forceBible,
        cast_member_id: castMemberId,
        regen_world: regenWorld,
      });
      setSelected(result);
      syncScenes(result);
      queryClient.invalidateQueries('videoStories');
      if (openPreviewOnDone && sharedAssetsFromProject(result).length) {
        setPreviewShared(true);
        setPreviewScene(null);
      }
      const sharedErr = result?.metadata?.shared_assets_error;
      if (sharedErr) setError(String(sharedErr));
    } catch (err) {
      setError(formatApiError(err, 'Failed to generate cast & world sheets'));
    } finally {
      setBusy('');
    }
  };

  const runVideos = async (sceneId = null, { forceRegenerate = false } = {}) => {
    const project = await ensureSaved();
    if (!project) return;
    if (!sharedAssetsFromProject(project).length) {
      setError('Generate Cast & world sheets once before videos (keeps characters consistent).');
      return;
    }
    const resolvedId = sceneId ? resolveSceneId(project, sceneId) : null;
    // Single-scene click always regenerates when a video already exists.
    const sceneObj = resolvedId
      ? (project.scenes || []).find((s) => s.id === resolvedId)
      : null;
    const force = forceRegenerate || Boolean(sceneObj?.video_filename);
    setBusy(resolvedId ? `video:${resolvedId}` : (force ? 'videos-regen' : 'videos'));
    setError('');
    try {
      const beforeNames = Object.fromEntries(
        (project.scenes || [])
          .filter((s) => s.video_filename)
          .map((s) => [s.id, s.video_filename]),
      );
      const result = await generateVideoStoryVideos(project.id, {
        scene_id: resolvedId,
        generate_all: !resolvedId,
        force_regenerate: force,
      });
      setSelected(result);
      syncScenes(result);
      setActiveStep(4);
      queryClient.invalidateQueries('videoStories');

      const scenes = result.scenes || [];
      const target = resolvedId ? scenes.find((s) => s.id === resolvedId) : null;
      const creditMsg = result?.metadata?.last_video_error
        || target?.error
        || scenes.map((s) => s.error).find((e) => e && /no credit|insufficient|top up/i.test(e));

      if (force) {
        const changed = resolvedId
          ? Boolean(target?.video_filename && target.video_filename !== beforeNames[resolvedId])
          : scenes.some((s) => s.video_filename && s.video_filename !== beforeNames[s.id]);
        if (!changed) {
          setError(
            creditMsg
              || 'Regenerate did not create new videos. Existing clips were kept. Check provider credit / scene errors, then try again.',
          );
          return;
        }
      }

      const ready = resolvedId
        ? scenes.find((s) => s.id === resolvedId && s.video_filename)
        : scenes.find((s) => s.video_filename);
      if (ready?.video_filename) {
        setPreviewScene(ready);
        setPreviewShared(false);
      }
    } catch (err) {
      setError(formatApiError(err, 'Failed to generate videos'));
    } finally {
      setBusy('');
    }
  };

  const addScene = () => {
    setScenes((prev) => {
      const next = [...prev, emptyScene(prev.length + 1)];
      setExpandedSceneId(next[next.length - 1].id);
      return next;
    });
  };

  const updateScene = (idx, patch) => {
    setScenes((prev) => prev.map((s, i) => (i === idx ? { ...s, ...patch } : s)));
  };

  const removeScene = (idx) => {
    setScenes((prev) => {
      const next = prev.filter((_, i) => i !== idx).map((s, i) => ({
        ...s,
        order: i + 1,
        continue_from_previous: i === 0 ? false : (s.continue_from_previous !== false),
      }));
      if (previewScene && prev[idx]?.id === previewScene.id) setPreviewScene(null);
      if (expandedSceneId && prev[idx]?.id === expandedSceneId) {
        setExpandedSceneId(next[0]?.id || null);
      }
      return next;
    });
  };

  const openPreview = (scene, e) => {
    if (e) e.stopPropagation();
    setPreviewShared(false);
    setPreviewScene(scene);
  };

  const castWorldCards = [
    ...(castDraft.cast || []).map((m) => ({
      id: m.id,
      name: m.name || 'Character',
      filename: m.image?.filename || null,
      description: m.canonical_description || '',
      prompt: m.image?.prompt || '',
      is_primary: Boolean(m.is_primary),
      kind: 'character',
    })),
    ...((castDraft.world?.canonical_description || castDraft.world?.image?.filename || castDraft.world?.image?.prompt)
      ? [{
        id: 'world',
        name: 'World',
        filename: castDraft.world?.image?.filename || null,
        description: castDraft.world?.canonical_description || '',
        prompt: castDraft.world?.image?.prompt || '',
        is_primary: false,
        kind: 'scenery',
      }]
      : []),
  ];
  const sharedAssets = castWorldCards.filter((a) => a.filename);
  const videoPreviewCount = scenes.filter(sceneHasPreview).length;
  const sharedError = selected?.metadata?.shared_assets_error;
  const castBusy = typeof busy === 'string' && (busy === 'images' || busy === 'save-cast' || busy.startsWith('cast:') || busy === 'world');

  const content = (
    <Box sx={{ height: '100%', minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', gap: 0.75 }}>
      <Grid container spacing={1} sx={{ flex: 1, minHeight: 0, height: '100%' }}>
        <Grid item xs={12} md={3} sx={{ display: 'flex', flexDirection: 'column', minHeight: 0, height: { md: '100%' } }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.75, flexShrink: 0 }}>
            <Typography sx={{ fontSize: '0.75rem', fontWeight: 600, flex: 1 }}>Projects</Typography>
            <Button size="small" startIcon={<AddIcon />} onClick={() => setOpenDialog(true)}>New</Button>
          </Box>
          <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
            {isLoading && <CircularProgress size={20} />}
            {projects.map((p) => (
              <Card
                key={p.id}
                variant="outlined"
                sx={{
                  mb: 0.5,
                  cursor: 'pointer',
                  borderColor: selected?.id === p.id ? 'primary.main' : 'divider',
                }}
                onClick={() => selectProject(p)}
              >
                <CardContent sx={{ py: 0.6, px: 1, '&:last-child': { pb: 0.6 } }}>
                  <Typography sx={{ fontSize: '0.72rem', fontWeight: 600 }} noWrap>{p.title}</Typography>
                  <Typography sx={{ fontSize: '0.6rem', color: 'text.secondary' }}>
                    {(p.scenes || []).length} scenes · {(p.cast || []).length} cast · {p.status}
                  </Typography>
                </CardContent>
              </Card>
            ))}
          </Box>
        </Grid>

        <Grid item xs={12} md={9} sx={{ display: 'flex', flexDirection: 'column', minHeight: 0, height: { md: '100%' } }}>
          {!selected ? (
            <Alert severity="info">Select or create a video story project.</Alert>
          ) : (
            <>
              <Stepper activeStep={activeStep} alternativeLabel sx={{ flexShrink: 0, mb: 0.5, '& .MuiStep-root': { px: 0.5 } }}>
                {PIPELINE_STEPS.map((label) => (
                  <Step key={label}>
                    <StepLabel sx={{ '& .MuiStepLabel-label': { fontSize: '0.58rem' } }}>{label}</StepLabel>
                  </Step>
                ))}
              </Stepper>

              <Box sx={{ flexShrink: 0, display: 'flex', gap: 0.75, flexWrap: 'wrap', mb: 0.75, alignItems: 'center' }}>
                <Button size="small" variant="outlined" onClick={() => saveScenes()} disabled={!!busy}>
                  Save
                </Button>
                <Button size="small" variant="contained" startIcon={<PolishIcon />} onClick={() => runPolish()} disabled={!!busy}>
                  AI polish all scenes
                </Button>
                <Button
                  size="small"
                  variant="contained"
                  color="secondary"
                  startIcon={<MovieIcon />}
                  onClick={() => runVideos()}
                  disabled={!!busy || !sharedAssets.length}
                >
                  Videos all
                </Button>
                {videoPreviewCount > 0 && (
                  <Button
                    size="small"
                    variant="outlined"
                    color="secondary"
                    startIcon={<RefreshIcon />}
                    onClick={() => runVideos(null, { forceRegenerate: true })}
                    disabled={!!busy || !sharedAssets.length}
                  >
                    Regenerate all videos
                  </Button>
                )}
                {videoPreviewCount > 0 && (
                  <Button
                    size="small"
                    variant="outlined"
                    color="success"
                    startIcon={<PreviewIcon />}
                    onClick={() => {
                      const first = scenes.find(sceneHasPreview);
                      if (first) openPreview(first);
                    }}
                  >
                    Preview videos ({videoPreviewCount})
                  </Button>
                )}
                {busy && <CircularProgress size={16} />}
              </Box>

              {selected.story_context || storyDraft.story_context || storyDraft.description ? null : (
                <Alert severity="info" sx={{ mb: 0.5, flexShrink: 0, py: 0, '& .MuiAlert-message': { fontSize: '0.65rem' } }}>
                  Optional: add story context below and use AI polish before scenes. Cast/world sheets use AI image generation.
                </Alert>
              )}

              <Box
                sx={{
                  flexShrink: 0,
                  mb: 0.75,
                  border: '1px solid',
                  borderColor: 'divider',
                  borderRadius: 1,
                  p: 0.75,
                  bgcolor: 'background.paper',
                }}
              >
                <Typography sx={{ fontSize: '0.7rem', fontWeight: 700, mb: 0.5 }}>Story (optional AI polish)</Typography>
                <Grid container spacing={0.75}>
                  <Grid item xs={12}>
                    <TextField
                      size="small"
                      fullWidth
                      multiline
                      minRows={2}
                      maxRows={4}
                      label="Story context"
                      value={storyDraft.story_context}
                      onChange={(e) => setStoryDraft((prev) => ({ ...prev, story_context: e.target.value }))}
                      helperText="Overall premise — used when polishing scenes and building cast"
                      inputProps={{ style: { fontSize: '0.72rem' } }}
                    />
                  </Grid>
                  <Grid item xs={12}>
                    <Box sx={{ display: 'flex', gap: 0.75, flexWrap: 'wrap' }}>
                      <Button
                        size="small"
                        variant="outlined"
                        startIcon={busy === 'polish:story_context' ? <CircularProgress size={12} /> : <PolishIcon />}
                        disabled={!!busy || !storyDraft.story_context.trim()}
                        onClick={() => runPolishContent({ field: 'story_context', sourceText: storyDraft.story_context })}
                      >
                        AI polish context
                      </Button>
                      <Button
                        size="small"
                        variant="outlined"
                        startIcon={busy === 'polish:description' ? <CircularProgress size={12} /> : <PolishIcon />}
                        disabled={!!busy || !storyDraft.description.trim()}
                        onClick={() => runPolishContent({ field: 'description', sourceText: storyDraft.description })}
                      >
                        AI polish description
                      </Button>
                    </Box>
                  </Grid>
                  <Grid item xs={12}>
                    <TextField
                      size="small"
                      fullWidth
                      multiline
                      minRows={1}
                      maxRows={2}
                      label="Short description"
                      value={storyDraft.description}
                      onChange={(e) => setStoryDraft((prev) => ({ ...prev, description: e.target.value }))}
                      inputProps={{ style: { fontSize: '0.72rem' } }}
                    />
                  </Grid>
                </Grid>
              </Box>

              {/* Cast & world — step before scenes, for series consistency */}
              {sharedError ? (
                <Alert severity="warning" sx={{ mb: 0.5, flexShrink: 0, py: 0, '& .MuiAlert-message': { fontSize: '0.65rem' } }}>
                  Shared assets: {String(sharedError)}
                </Alert>
              ) : null}

              <Box
                sx={{
                  flexShrink: 0,
                  mb: 0.75,
                  maxHeight: '42%',
                  minHeight: 140,
                  display: 'flex',
                  flexDirection: 'column',
                  border: '1px solid',
                  borderColor: sharedAssets.length ? 'success.main' : 'divider',
                  borderRadius: 1,
                  p: 0.75,
                  bgcolor: 'background.paper',
                  overflow: 'hidden',
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, flexWrap: 'wrap', mb: 0.5, flexShrink: 0 }}>
                  <CastIcon sx={{ fontSize: 16, color: 'text.secondary' }} />
                  <Typography sx={{ fontSize: '0.7rem', fontWeight: 700, flex: 1 }}>
                    1. Cast & world — AI image sheets
                  </Typography>
                  <Button size="small" variant="outlined" disabled={!!busy || !castWorldCards.length} onClick={() => saveCastWorld()}>
                    Save prompts
                  </Button>
                  <Button
                    size="small"
                    variant="contained"
                    startIcon={<ImageIcon />}
                    disabled={!!busy}
                    onClick={() => runCastAndWorld({ forceBible: !(castDraft.cast || []).length })}
                  >
                    {(castDraft.cast || []).length ? 'AI fill missing sheets' : 'AI build cast & world'}
                  </Button>
                  <Button
                    size="small"
                    variant="outlined"
                    startIcon={<ImageIcon />}
                    disabled={!!busy || !castWorldCards.length}
                    onClick={() => runCastAndWorld({ forceRegenerate: true, openPreviewOnDone: true })}
                  >
                    AI regenerate all
                  </Button>
                  {sharedAssets.length > 0 && (
                    <Button
                      size="small"
                      startIcon={<PreviewIcon />}
                      onClick={() => { setPreviewShared(true); setPreviewScene(null); }}
                    >
                      Preview
                    </Button>
                  )}
                  {castBusy && <CircularProgress size={14} />}
                </Box>
                <Typography sx={{ fontSize: '0.58rem', color: 'text.secondary', mb: 0.5, flexShrink: 0 }}>
                  Polish descriptions with AI (optional), edit image prompts, save, then AI-generate reference sheets for video consistency.
                </Typography>
                <Box sx={{ display: 'flex', gap: 0.75, mb: 0.75, flexShrink: 0, flexWrap: 'wrap' }}>
                  <TextField
                    size="small"
                    fullWidth
                    label="Style bible"
                    value={castDraft.style_bible || ''}
                    onChange={(e) => setCastDraft((prev) => ({ ...prev, style_bible: e.target.value }))}
                    sx={{ flex: '1 1 200px', '& .MuiInputBase-input': { fontSize: '0.7rem' } }}
                  />
                  <Button
                    size="small"
                    variant="outlined"
                    startIcon={busy === 'polish:style_bible' ? <CircularProgress size={12} /> : <PolishIcon />}
                    disabled={!!busy || !(castDraft.style_bible || '').trim()}
                    onClick={() => runPolishContent({ field: 'style_bible', sourceText: castDraft.style_bible })}
                    sx={{ alignSelf: 'flex-end' }}
                  >
                    AI polish style
                  </Button>
                </Box>

                {(castDraft.cast || []).length === 0 && !castDraft.world?.canonical_description ? (
                  <Alert severity="info" sx={{ py: 0, '& .MuiAlert-message': { fontSize: '0.65rem' } }}>
                    No cast yet. Add scene prompts, polish, then click Build cast & world.
                  </Alert>
                ) : (
                  <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', pr: 0.25 }}>
                    {(castDraft.cast || []).map((member) => {
                      const isThisBusy = busy === `cast:${member.id}` || busy === 'images';
                      const expanded = expandedCastId === member.id;
                      return (
                        <Accordion
                          key={member.id}
                          expanded={expanded}
                          onChange={(_, isExp) => setExpandedCastId(isExp ? member.id : null)}
                          disableGutters
                          elevation={0}
                          sx={{
                            mb: 0.5,
                            border: '1px solid',
                            borderColor: member.is_primary ? 'primary.main' : 'divider',
                            borderRadius: '6px !important',
                            '&:before': { display: 'none' },
                          }}
                        >
                          <AccordionSummary expandIcon={<ExpandMoreIcon sx={{ fontSize: 18 }} />} sx={{ minHeight: 36, px: 1 }}>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, width: '100%', pr: 0.5 }}>
                              {member.image?.filename ? (
                                <Box
                                  component="img"
                                  src={getVideoStoryFileUrl(selected.id, 'shared', member.image.filename)}
                                  alt={member.name}
                                  sx={{ width: 40, height: 30, objectFit: 'cover', borderRadius: 0.5 }}
                                />
                              ) : (
                                <Box sx={{ width: 40, height: 30, borderRadius: 0.5, bgcolor: 'action.hover' }} />
                              )}
                              <Typography sx={{ fontSize: '0.68rem', fontWeight: 600, flex: 1 }} noWrap>
                                {member.is_primary ? '★ ' : ''}{member.name || 'Character'}
                              </Typography>
                              <Button
                                size="small"
                                variant="contained"
                                startIcon={isThisBusy ? <CircularProgress size={10} color="inherit" /> : <ImageIcon />}
                                disabled={!!busy}
                                onClick={(e) => {
                                  e.stopPropagation();
                                  runCastAndWorld({
                                    castMemberId: member.id,
                                    forceRegenerate: true,
                                    openPreviewOnDone: false,
                                  });
                                }}
                                sx={{ fontSize: '0.58rem', py: 0.1, minHeight: 24 }}
                              >
                                {member.image?.filename ? 'AI redo image' : 'AI image'}
                              </Button>
                            </Box>
                          </AccordionSummary>
                          <AccordionDetails sx={{ px: 1, pt: 0, pb: 1 }}>
                            <Grid container spacing={0.75}>
                              <Grid item xs={12} sm={4}>
                                <TextField
                                  size="small"
                                  fullWidth
                                  label="Name"
                                  value={member.name || ''}
                                  onChange={(e) => updateCastMember(member.id, { name: e.target.value })}
                                  inputProps={{ style: { fontSize: '0.72rem' } }}
                                />
                              </Grid>
                              <Grid item xs={12} sm={8}>
                                <Box sx={{ display: 'flex', gap: 0.5, alignItems: 'flex-start' }}>
                                  <TextField
                                    size="small"
                                    fullWidth
                                    multiline
                                    minRows={2}
                                    maxRows={4}
                                    label="Look description (locked identity)"
                                    value={member.canonical_description || ''}
                                    onChange={(e) => updateCastMember(member.id, { canonical_description: e.target.value })}
                                    helperText="Used in every scene video prompt"
                                    inputProps={{ style: { fontSize: '0.7rem' } }}
                                  />
                                  <Button
                                    size="small"
                                    variant="outlined"
                                    startIcon={busy === `polish:cast_description:${member.id}` ? <CircularProgress size={10} /> : <PolishIcon />}
                                    disabled={!!busy || !(member.canonical_description || '').trim()}
                                    onClick={() => runPolishContent({
                                      field: 'cast_description',
                                      castMemberId: member.id,
                                      sourceText: member.canonical_description,
                                    })}
                                    sx={{ flexShrink: 0, mt: 0.5, fontSize: '0.58rem' }}
                                  >
                                    AI polish
                                  </Button>
                                </Box>
                              </Grid>
                              <Grid item xs={12}>
                                <TextField
                                  size="small"
                                  fullWidth
                                  multiline
                                  minRows={2}
                                  maxRows={5}
                                  label="Image generation prompt"
                                  value={member.image?.prompt || ''}
                                  onChange={(e) => updateCastMember(member.id, { image: { prompt: e.target.value } })}
                                  helperText="Edit prompt, then AI image to redraw the sheet"
                                  inputProps={{ style: { fontSize: '0.7rem' } }}
                                />
                              </Grid>
                            </Grid>
                          </AccordionDetails>
                        </Accordion>
                      );
                    })}

                    <Accordion
                      expanded={expandedCastId === 'world'}
                      onChange={(_, isExp) => setExpandedCastId(isExp ? 'world' : null)}
                      disableGutters
                      elevation={0}
                      sx={{
                        mb: 0.5,
                        border: '1px solid',
                        borderColor: 'divider',
                        borderRadius: '6px !important',
                        '&:before': { display: 'none' },
                      }}
                    >
                      <AccordionSummary expandIcon={<ExpandMoreIcon sx={{ fontSize: 18 }} />} sx={{ minHeight: 36, px: 1 }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, width: '100%', pr: 0.5 }}>
                          {castDraft.world?.image?.filename ? (
                            <Box
                              component="img"
                              src={getVideoStoryFileUrl(selected.id, 'shared', castDraft.world.image.filename)}
                              alt="World"
                              sx={{ width: 40, height: 30, objectFit: 'cover', borderRadius: 0.5 }}
                            />
                          ) : (
                            <Box sx={{ width: 40, height: 30, borderRadius: 0.5, bgcolor: 'action.hover' }} />
                          )}
                          <Typography sx={{ fontSize: '0.68rem', fontWeight: 600, flex: 1 }}>World</Typography>
                          <Button
                            size="small"
                            variant="contained"
                            startIcon={busy === 'world' || busy === 'images' ? <CircularProgress size={10} color="inherit" /> : <ImageIcon />}
                            disabled={!!busy}
                            onClick={(e) => {
                              e.stopPropagation();
                              runCastAndWorld({
                                regenWorld: true,
                                forceRegenerate: true,
                                openPreviewOnDone: false,
                              });
                            }}
                            sx={{ fontSize: '0.58rem', py: 0.1, minHeight: 24 }}
                          >
                            {castDraft.world?.image?.filename ? 'AI redo image' : 'AI image'}
                          </Button>
                        </Box>
                      </AccordionSummary>
                      <AccordionDetails sx={{ px: 1, pt: 0, pb: 1 }}>
                        <Grid container spacing={0.75}>
                          <Grid item xs={12}>
                            <Box sx={{ display: 'flex', gap: 0.5, alignItems: 'flex-start' }}>
                              <TextField
                                size="small"
                                fullWidth
                                multiline
                                minRows={2}
                                maxRows={4}
                                label="World description (locked identity)"
                                value={castDraft.world?.canonical_description || ''}
                                onChange={(e) => updateWorldDraft({ canonical_description: e.target.value })}
                                helperText="Used in every scene video prompt"
                                inputProps={{ style: { fontSize: '0.7rem' } }}
                              />
                              <Button
                                size="small"
                                variant="outlined"
                                startIcon={busy === 'polish:world_description' ? <CircularProgress size={10} /> : <PolishIcon />}
                                disabled={!!busy || !(castDraft.world?.canonical_description || '').trim()}
                                onClick={() => runPolishContent({
                                  field: 'world_description',
                                  sourceText: castDraft.world?.canonical_description,
                                })}
                                sx={{ flexShrink: 0, mt: 0.5, fontSize: '0.58rem' }}
                              >
                                AI polish
                              </Button>
                            </Box>
                          </Grid>
                          <Grid item xs={12}>
                            <TextField
                              size="small"
                              fullWidth
                              multiline
                              minRows={2}
                              maxRows={5}
                              label="World image generation prompt"
                              value={castDraft.world?.image?.prompt || ''}
                              onChange={(e) => updateWorldDraft({ image: { prompt: e.target.value } })}
                              helperText="Edit prompt, then AI image to redraw the world sheet"
                              inputProps={{ style: { fontSize: '0.7rem' } }}
                            />
                          </Grid>
                        </Grid>
                      </AccordionDetails>
                    </Accordion>
                  </Box>
                )}
              </Box>

              <Typography sx={{ fontSize: '0.7rem', fontWeight: 700, mb: 0.5, flexShrink: 0 }}>
                2. Scenes
              </Typography>

              <Box
                sx={{
                  flex: '1 1 0',
                  minHeight: 0,
                  overflowY: 'auto',
                  overflowX: 'hidden',
                  pr: 0.5,
                  WebkitOverflowScrolling: 'touch',
                }}
              >
                {scenes.map((scene, idx) => {
                  const canPreview = sceneHasPreview(scene);
                  const expanded = expandedSceneId === scene.id;
                  return (
                    <Accordion
                      key={scene.id || idx}
                      expanded={expanded}
                      onChange={(_, isExpanded) => setExpandedSceneId(isExpanded ? scene.id : null)}
                      disableGutters
                      elevation={0}
                      sx={{
                        mb: 0.5,
                        border: '1px solid',
                        borderColor: 'divider',
                        borderRadius: '6px !important',
                        '&:before': { display: 'none' },
                        bgcolor: 'background.paper',
                      }}
                    >
                      <AccordionSummary
                        expandIcon={<ExpandMoreIcon sx={{ fontSize: 18 }} />}
                        sx={{
                          minHeight: 40,
                          px: 1,
                          '& .MuiAccordionSummary-content': { my: 0.5, alignItems: 'center', gap: 0.75 },
                        }}
                      >
                        <Typography sx={{ fontSize: '0.72rem', fontWeight: 600, flex: 1, minWidth: 0 }} noWrap>
                          {scene.order}. {scene.title || 'Untitled scene'}
                        </Typography>
                        <Chip
                          size="small"
                          label={scene.status}
                          color={SCENE_STATUS_COLOR[scene.status] || 'default'}
                          sx={{ fontSize: '0.55rem', height: 18 }}
                        />
                        <Chip
                          size="small"
                          variant="outlined"
                          label={`${scene.duration_seconds || 5}s`}
                          sx={{ fontSize: '0.55rem', height: 18 }}
                        />
                        {scene.order > 1 && (
                          <Chip
                            size="small"
                            variant="outlined"
                            color={scene.continue_from_previous === false ? 'warning' : 'default'}
                            label={scene.continue_from_previous === false ? 'cut' : 'cont'}
                            sx={{ fontSize: '0.55rem', height: 18 }}
                          />
                        )}
                        {canPreview && (
                          <Button
                            size="small"
                            variant="contained"
                            color="success"
                            startIcon={<PreviewIcon />}
                            onClick={(e) => openPreview(scene, e)}
                            sx={{ fontSize: '0.62rem', py: 0.15, px: 0.75, minHeight: 24 }}
                          >
                            Preview
                          </Button>
                        )}
                        <IconButton
                          size="small"
                          onClick={(e) => { e.stopPropagation(); removeScene(idx); }}
                          sx={{ p: 0.25 }}
                        >
                          <DeleteIcon sx={{ fontSize: 16 }} />
                        </IconButton>
                      </AccordionSummary>
                      <AccordionDetails sx={{ px: 1, pt: 0, pb: 1 }}>
                        <Grid container spacing={0.75}>
                          <Grid item xs={12} sm={4}>
                            <TextField
                              fullWidth
                              size="small"
                              label="Title"
                              value={scene.title}
                              onChange={(e) => updateScene(idx, { title: e.target.value })}
                              inputProps={{ style: { fontSize: '0.75rem' } }}
                            />
                          </Grid>
                          <Grid item xs={12} sm={8}>
                            <TextField
                              fullWidth
                              size="small"
                              multiline
                              minRows={1}
                              maxRows={2}
                              label="Short query"
                              placeholder='e.g. "he runs on the street"'
                              value={scene.user_prompt}
                              onChange={(e) => updateScene(idx, { user_prompt: e.target.value })}
                              helperText="Keep it short — polish expands it to a timed clip from the previous scene"
                              inputProps={{ style: { fontSize: '0.75rem' } }}
                            />
                          </Grid>
                          <Grid item xs={6} sm={2}>
                            <TextField
                              fullWidth
                              size="small"
                              type="number"
                              label="Seconds"
                              value={scene.duration_seconds ?? 5}
                              onChange={(e) => {
                                const n = Number(e.target.value);
                                updateScene(idx, { duration_seconds: n > 0 ? n : 5 });
                              }}
                              inputProps={{ min: 1, max: 30, step: 1, style: { fontSize: '0.75rem' } }}
                            />
                          </Grid>
                          <Grid item xs={6} sm={2} sx={{ display: 'flex', alignItems: 'center' }}>
                            <FormControlLabel
                              control={(
                                <Switch
                                  size="small"
                                  checked={scene.order > 1 && scene.continue_from_previous !== false}
                                  disabled={scene.order <= 1 || !!busy}
                                  onChange={(e) => updateScene(idx, {
                                    continue_from_previous: e.target.checked,
                                    cut_note: e.target.checked ? '' : (scene.cut_note || ''),
                                  })}
                                />
                              )}
                              label={<Typography sx={{ fontSize: '0.62rem' }}>Continue last</Typography>}
                            />
                          </Grid>
                          {scene.order > 1 && scene.continue_from_previous === false ? (
                            <Grid item xs={12}>
                              <TextField
                                fullWidth
                                size="small"
                                label="Hard cut note (optional)"
                                placeholder="e.g. jump to night / new location / different character focus"
                                value={scene.cut_note || ''}
                                onChange={(e) => updateScene(idx, { cut_note: e.target.value })}
                                helperText="Polish will treat this as a break instead of continuing the previous shot"
                                inputProps={{ style: { fontSize: '0.72rem' } }}
                              />
                            </Grid>
                          ) : null}
                          {scene.notes ? (
                            <Grid item xs={12}>
                              <Typography sx={{ fontSize: '0.58rem', color: 'text.secondary' }}>
                                Continuity: {scene.notes}
                              </Typography>
                            </Grid>
                          ) : null}
                          {scene.polished_prompt ? (
                            <Grid item xs={12}>
                              <TextField
                                fullWidth
                                size="small"
                                multiline
                                minRows={2}
                                maxRows={4}
                                label={`AI-polished clip (${scene.duration_seconds || 5}s · ${(scene.polished_prompt || '').trim().split(/\s+/).filter(Boolean).length} words)`}
                                value={scene.polished_prompt}
                                onChange={(e) => updateScene(idx, { polished_prompt: e.target.value })}
                                helperText="Editable — or re-run AI polish from your short query"
                                inputProps={{ style: { fontSize: '0.7rem' } }}
                              />
                            </Grid>
                          ) : null}
                          {scene.video_filename && (
                            <Grid item xs={12}>
                              <Alert
                                severity="success"
                                sx={{ py: 0, '& .MuiAlert-message': { fontSize: '0.65rem', width: '100%' } }}
                                action={(
                                  <Button color="inherit" size="small" startIcon={<PreviewIcon />} onClick={() => openPreview(scene)}>
                                    Open preview
                                  </Button>
                                )}
                              >
                                Video ready
                              </Alert>
                            </Grid>
                          )}
                          {scene.error ? (
                            <Grid item xs={12}>
                              <Alert severity="error" sx={{ py: 0, '& .MuiAlert-message': { fontSize: '0.65rem' } }}>
                                {scene.error}
                              </Alert>
                            </Grid>
                          ) : null}
                          <Grid item xs={12}>
                            <Box sx={{ display: 'flex', gap: 0.75, flexWrap: 'wrap', alignItems: 'center' }}>
                              <Button size="small" variant="outlined" startIcon={<PolishIcon />} onClick={() => runPolish(scene.id)} disabled={!!busy || !scene.user_prompt}>
                                AI polish scene
                              </Button>
                              <Button
                                size="small"
                                variant="contained"
                                startIcon={scene.video_filename ? <RefreshIcon /> : <MovieIcon />}
                                onClick={() => runVideos(scene.id, { forceRegenerate: Boolean(scene.video_filename) })}
                                disabled={!!busy || !sharedAssets.length}
                              >
                                {scene.video_filename ? 'Regen video' : 'Video'}
                              </Button>
                              <Typography sx={{ fontSize: '0.58rem', color: 'text.secondary' }}>
                                {i2vReferenceHint(scene)}
                              </Typography>
                              {canPreview && (
                                <Button
                                  size="small"
                                  color="success"
                                  variant="outlined"
                                  startIcon={<PreviewIcon />}
                                  onClick={() => openPreview(scene)}
                                >
                                  Preview
                                </Button>
                              )}
                            </Box>
                          </Grid>
                        </Grid>
                      </AccordionDetails>
                    </Accordion>
                  );
                })}
                <Button size="small" startIcon={<AddIcon />} onClick={addScene} sx={{ mt: 0.5, mb: 1 }}>
                  Add scene
                </Button>
              </Box>
            </>
          )}
          {error ? (
            <Alert severity="error" sx={{ mt: 0.75, flexShrink: 0 }}>
              {typeof error === 'string' ? error : formatApiError({ response: { data: { detail: error } } })}
            </Alert>
          ) : null}
        </Grid>
      </Grid>

      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>New video story</DialogTitle>
        <DialogContent>
          <TextField fullWidth size="small" label="Title" sx={{ mt: 1, mb: 1 }} value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} />
          <TextField fullWidth size="small" label="Description" sx={{ mb: 1 }} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          <TextField
            fullWidth
            size="small"
            multiline
            minRows={3}
            label="Story context"
            value={form.story_context}
            onChange={(e) => setForm({ ...form, story_context: e.target.value })}
            helperText="Overall premise — used when locking cast/world and polishing scenes"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button
            variant="contained"
            disabled={!form.title.trim() || createMutation.isLoading}
            onClick={() => createMutation.mutate(form)}
          >
            Create
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={Boolean(selected?.id && (previewShared || previewScene))}
        onClose={() => { setPreviewScene(null); setPreviewShared(false); }}
        maxWidth="md"
        fullWidth
        PaperProps={{
          sx: {
            height: 'min(88vh, 720px)',
            maxHeight: '88vh',
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
          },
        }}
      >
        <DialogTitle sx={{ flexShrink: 0, fontSize: '0.9rem', py: 1, px: 2 }}>
          {previewShared
            ? 'Preview · Shared cast & world'
            : `Preview · Scene ${previewScene?.order}: ${previewScene?.title || 'Untitled'}`}
        </DialogTitle>
        <DialogContent
          dividers
          sx={{
            flex: 1,
            minHeight: 0,
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            p: 1.5,
            '&.MuiDialogContent-root': { pt: 1.5 },
          }}
        >
          {selected?.id && previewShared ? (
            <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', minHeight: 0, gap: 1 }}>
              {(selected.style_bible || selected.world?.canonical_description) ? (
                <Box sx={{ flexShrink: 0, display: 'flex', flexDirection: 'column', gap: 0.25 }}>
                  {selected.style_bible ? (
                    <Typography sx={{ fontSize: '0.68rem' }} noWrap title={selected.style_bible}>
                      <strong>Style:</strong> {selected.style_bible}
                    </Typography>
                  ) : null}
                  {selected.world?.canonical_description ? (
                    <Typography sx={{ fontSize: '0.68rem' }} noWrap title={selected.world.canonical_description}>
                      <strong>World:</strong> {selected.world.canonical_description}
                    </Typography>
                  ) : null}
                </Box>
              ) : null}
              {sharedAssets.length > 0 ? (
                <Box
                  sx={{
                    flex: 1,
                    minHeight: 0,
                    display: 'flex',
                    gap: 1,
                    overflowX: 'auto',
                    overflowY: 'hidden',
                    alignItems: 'stretch',
                    pb: 0.25,
                  }}
                >
                  {sharedAssets.map((asset) => (
                    <Box
                      key={asset.id}
                      sx={{
                        flex: '0 0 auto',
                        width: 180,
                        maxWidth: '40vw',
                        border: '1px solid',
                        borderColor: 'divider',
                        borderRadius: 1,
                        p: 0.75,
                        display: 'flex',
                        flexDirection: 'column',
                        minHeight: 0,
                      }}
                    >
                      <Box
                        component="img"
                        src={getVideoStoryFileUrl(selected.id, 'shared', asset.filename)}
                        alt={asset.name}
                        sx={{
                          flex: 1,
                          minHeight: 0,
                          width: '100%',
                          objectFit: 'contain',
                          borderRadius: 0.75,
                          display: 'block',
                          bgcolor: 'action.hover',
                        }}
                      />
                      <Typography sx={{ fontSize: '0.65rem', fontWeight: 600, mt: 0.5 }} noWrap>
                        {asset.is_primary ? 'Primary · ' : ''}{asset.name}
                      </Typography>
                      <Typography sx={{ fontSize: '0.58rem', color: 'text.secondary' }} noWrap title={asset.description}>
                        {asset.description}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              ) : (
                <Alert severity="info" sx={{ flexShrink: 0 }}>No shared cast/world sheets yet. Run Cast & world.</Alert>
              )}
            </Box>
          ) : null}

          {selected?.id && previewScene && !previewShared ? (
            <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', minHeight: 0, gap: 1 }}>
              {sharedAssets.length > 0 && (
                <Box sx={{ flexShrink: 0 }}>
                  <Typography sx={{ fontSize: '0.65rem', fontWeight: 600, mb: 0.5, color: 'text.secondary' }}>
                    Cast & world reference
                  </Typography>
                  <Box sx={{ display: 'flex', gap: 0.5, overflowX: 'auto', overflowY: 'hidden' }}>
                    {sharedAssets.map((asset) => (
                      <Box
                        key={asset.id}
                        component="img"
                        src={getVideoStoryFileUrl(selected.id, 'shared', asset.filename)}
                        alt={asset.name}
                        title={asset.name}
                        sx={{
                          width: 64,
                          height: 48,
                          flexShrink: 0,
                          objectFit: 'cover',
                          borderRadius: 0.5,
                          border: '1px solid',
                          borderColor: 'divider',
                        }}
                      />
                    ))}
                  </Box>
                </Box>
              )}

              {previewScene.video_filename ? (
                <Box
                  sx={{
                    flex: 1,
                    minHeight: 0,
                    display: 'flex',
                    flexDirection: 'column',
                    borderRadius: 1,
                    overflow: 'hidden',
                    bgcolor: 'common.black',
                  }}
                >
                  <Box
                    component="video"
                    controls
                    src={getVideoStoryFileUrl(selected.id, 'videos', previewScene.video_filename)}
                    sx={{
                      flex: 1,
                      minHeight: 0,
                      width: '100%',
                      height: '100%',
                      maxHeight: '100%',
                      objectFit: 'contain',
                      display: 'block',
                    }}
                  />
                </Box>
              ) : (
                <Alert severity="info" sx={{ flexShrink: 0 }}>No generated video for this scene yet.</Alert>
              )}

              {scenes.filter(sceneHasPreview).length > 1 && (
                <Box sx={{ flexShrink: 0, display: 'flex', gap: 0.5, flexWrap: 'wrap', justifyContent: 'center' }}>
                  {scenes.filter(sceneHasPreview).map((s) => (
                    <Chip
                      key={s.id}
                      size="small"
                      label={`Scene ${s.order}`}
                      color={s.id === previewScene.id ? 'primary' : 'default'}
                      onClick={() => setPreviewScene(s)}
                      clickable
                      sx={{ height: 24, '& .MuiChip-label': { px: 1, fontSize: '0.65rem' } }}
                    />
                  ))}
                </Box>
              )}
            </Box>
          ) : null}
        </DialogContent>
        <DialogActions sx={{ flexShrink: 0, py: 1, px: 2, gap: 0.5 }}>
          {!previewShared && previewScene?.video_filename && (
            <Button
              size="small"
              startIcon={<DownloadIcon />}
              href={getVideoStoryFileUrl(selected.id, 'videos', previewScene.video_filename, true)}
              download
            >
              Download video
            </Button>
          )}
          {previewShared && sharedAssets.length > 0 && sharedAssets.map((asset) => (
            <Button
              key={asset.id}
              size="small"
              startIcon={<DownloadIcon />}
              href={getVideoStoryFileUrl(selected.id, 'shared', asset.filename, true)}
              download
              sx={{ display: { xs: 'none', sm: 'inline-flex' } }}
            >
              {asset.name}
            </Button>
          ))}
          {!previewShared && sharedAssets.length > 0 && (
            <Button size="small" startIcon={<ImageIcon />} onClick={() => { setPreviewShared(true); setPreviewScene(null); }}>
              Cast & world
            </Button>
          )}
          <Box sx={{ flex: 1 }} />
          <Button size="small" onClick={() => { setPreviewScene(null); setPreviewShared(false); }}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );

  if (suppressModuleHeader) return content;

  return (
    <ModuleShell
      title="Video story generator"
      helpText="Write short scene queries, optionally AI-polish story context and cast/world copy, then AI-generate cast/world image sheets. AI-polish scenes expands each into a timed clip. Generate videos last. Configure VIDEO_STORY_* in .env."
    >
      {content}
    </ModuleShell>
  );
};

export default VideoStoryGenerator;
