import React, { useCallback, useEffect, useRef, useState } from 'react';
import AddIcon from '@mui/icons-material/Add';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import ChatBubbleOutlineIcon from '@mui/icons-material/ChatBubbleOutline';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import EditOutlinedIcon from '@mui/icons-material/EditOutlined';
import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  CircularProgress,
  Drawer,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
} from '@mui/material';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { ADMIN_BASE_PATH } from '../../adminPaths';
import CommentsModal from '../../components/admin/CommentsModal';
import { useConfirm } from '../../components/ConfirmDialog';

const MEDIA_ACCEPT = '.jpg,.jpeg,.png,.gif,.webp,.bmp,.svg,.heic,.heif,.mp4,.webm,.mov,.avi,.mkv,.m4v';
const MAX_IMAGE_BYTES = 20 * 1024 * 1024;
const MAX_VIDEO_BYTES = 200 * 1024 * 1024;
const MAX_STORY_MEDIA = 4;

function dateTimeDisplay(value) {
  if (!value) return '—';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function authorInitials(name) {
  const parts = String(name || '')
    .trim()
    .split(/\s+/)
    .filter(Boolean);
  if (!parts.length) return '?';
  if (parts.length === 1) return parts[0].slice(0, 1).toUpperCase();
  return (parts[0].slice(0, 1) + parts[parts.length - 1].slice(0, 1)).toUpperCase();
}

function isVideoFile(file) {
  const name = String(file?.name || '').toLowerCase();
  return ['.mp4', '.webm', '.mov', '.avi', '.mkv', '.m4v'].some((ext) => name.endsWith(ext));
}

function validateMediaFile(file) {
  if (!file) return 'Choose an image or video file.';
  const name = String(file.name || '').toLowerCase();
  const image = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.svg', '.heic', '.heif'].some((ext) =>
    name.endsWith(ext)
  );
  const video = isVideoFile(file);
  if (!image && !video) return 'Only images and videos are allowed.';
  if (image && file.size > MAX_IMAGE_BYTES) return 'Images must be under 20 MB.';
  if (video && file.size > MAX_VIDEO_BYTES) return 'Videos must be under 200 MB.';
  return '';
}

function PostMediaPreview({ postId, attachment }) {
  const [blobUrl, setBlobUrl] = useState('');
  const [busy, setBusy] = useState(true);
  const [err, setErr] = useState('');

  useEffect(() => {
    if (!attachment?.id || !postId) return undefined;
    let alive = true;
    let objectUrl = '';
    async function run() {
      setBusy(true);
      setErr('');
      try {
        const url = tranEndpoints.entityAttachmentDownload('story-posts', postId, attachment.id);
        const res = await tranApi.get(url, { responseType: 'blob' });
        if (!alive) return;
        objectUrl = URL.createObjectURL(res.data);
        setBlobUrl(objectUrl);
      } catch (e) {
        if (!alive) return;
        setErr(e.response?.data?.error || e.message || 'Preview unavailable');
      } finally {
        if (alive) setBusy(false);
      }
    }
    run();
    return () => {
      alive = false;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [postId, attachment?.id]);

  if (busy) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 3 }}>
        <CircularProgress size={24} />
      </Box>
    );
  }
  if (err) {
    return (
      <Typography variant="body2" color="text.secondary">
        {err}
      </Typography>
    );
  }
  if (attachment.kind === 'video') {
    return (
      <Box
        component="video"
        src={blobUrl}
        controls
        sx={{ width: '100%', maxHeight: 220, borderRadius: 1, bgcolor: 'black' }}
      />
    );
  }
  return (
    <Box
      component="img"
      src={blobUrl}
      alt={attachment.original_name}
      sx={{ width: '100%', maxHeight: 220, objectFit: 'cover', borderRadius: 1, bgcolor: 'action.hover' }}
    />
  );
}

function normalizePost(raw) {
  if (!raw || typeof raw !== 'object') return null;
  const id = raw.id ?? raw.ID;
  if (id == null || id === '') return null;
  return { ...raw, id: Number(id) || id };
}

function mergePosts(prev, incoming) {
  const map = new Map();
  (prev || []).forEach((p) => {
    const n = normalizePost(p);
    if (n) map.set(String(n.id), n);
  });
  (incoming || []).forEach((p) => {
    const n = normalizePost(p);
    if (n) map.set(String(n.id), { ...(map.get(String(n.id)) || {}), ...n });
  });
  return Array.from(map.values()).sort(
    (a, b) => new Date(b.created_on || 0) - new Date(a.created_on || 0)
  );
}

function countCommentTree(tree) {
  let n = 0;
  for (const c of tree) {
    n += 1;
    if (Array.isArray(c.replies) && c.replies.length) n += countCommentTree(c.replies);
  }
  return n;
}

export default function StoryBoard() {
  const { confirm } = useConfirm();
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const [composeOpen, setComposeOpen] = useState(false);
  const [editing, setEditing] = useState(null);
  const [draftTitle, setDraftTitle] = useState('');
  const [draftContent, setDraftContent] = useState('');
  const [pickedMedia, setPickedMedia] = useState([]);
  const [aiPrompt, setAiPrompt] = useState('');
  const [aiImageCount, setAiImageCount] = useState(1);
  const [aiGenerating, setAiGenerating] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState('');

  const [commentsOpen, setCommentsOpen] = useState(false);
  const [commentsPost, setCommentsPost] = useState(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailPost, setDetailPost] = useState(null);

  const deepLinkHandledRef = useRef('');
  const deepLinkFetchRef = useRef('');

  const load = useCallback(async () => {
    const res = await tranApi.get(tranEndpoints.storyPosts, {
      params: { _: Date.now() },
    });
    const next = Array.isArray(res.data) ? res.data.map((p) => normalizePost(p)).filter(Boolean) : [];
    setPosts(next);
    setError('');
    return next;
  }, []);

  useEffect(() => {
    load()
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load Stories'))
      .finally(() => setLoading(false));
  }, [load]);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const postId = params.get('post');
    if (!postId || loading) return;
    if (deepLinkHandledRef.current === postId) return;

    const openDeepLink = async (id) => {
      deepLinkHandledRef.current = postId;
      window.history.replaceState({}, '', `${ADMIN_BASE_PATH}/stories`);
      const found = posts.find((p) => String(p.id) === String(id));
      if (found) await openDetail(found);
    };

    const found = posts.find((p) => String(p.id) === String(postId));
    if (found) {
      void openDeepLink(found.id);
      return;
    }

    if (deepLinkFetchRef.current === postId) return;
    deepLinkFetchRef.current = postId;

    let alive = true;
    tranApi
      .get(tranEndpoints.storyPostFull(postId))
      .then((res) => {
        if (!alive) return;
        const p = normalizePost(res.data);
        if (!p) {
          deepLinkHandledRef.current = postId;
          return;
        }
        setPosts((prev) => mergePosts(prev, [p]));
        void openDeepLink(p.id);
      })
      .catch(() => {
        if (alive) deepLinkHandledRef.current = postId;
      });
    return () => {
      alive = false;
    };
  }, [posts, loading]);

  const handleCommentThreadChanged = useCallback(
    (tree) => {
      if (!commentsPost?.id) return;
      const count = countCommentTree(Array.isArray(tree) ? tree : []);
      setPosts((prev) =>
        prev.map((p) => (String(p.id) === String(commentsPost.id) ? { ...p, comment_count: count } : p))
      );
      setDetailPost((prev) =>
        prev && String(prev.id) === String(commentsPost.id) ? { ...prev, comment_count: count } : prev
      );
    },
    [commentsPost?.id]
  );

  const openCreate = () => {
    setEditing(null);
    setDraftTitle('');
    setDraftContent('');
    setPickedMedia([]);
    setAiPrompt('');
    setAiImageCount(1);
    setFormError('');
    setComposeOpen(true);
  };

  const openDetail = async (post) => {
    setFormError('');
    try {
      const full = await tranApi.get(tranEndpoints.storyPostFull(post.id));
      const data = full.data || post;
      setDetailPost({ ...post, ...data, attachments: data.attachments || [] });
      setDetailOpen(true);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to load story.');
    }
  };

  const openEdit = async (post) => {
    setFormError('');
    setPickedMedia([]);
    setAiPrompt('');
    try {
      const full = await tranApi.get(tranEndpoints.storyPostFull(post.id));
      const data = full.data || post;
      setEditing({ ...post, ...data, attachments: data.attachments || [] });
      setDraftTitle(data.title || '');
      setDraftContent(data.content || '');
      setComposeOpen(true);
      setDetailOpen(false);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to load post.');
    }
  };

  const generateWithAI = async () => {
    const prompt = aiPrompt.trim();
    if (!prompt) {
      setFormError('Enter a prompt to generate a story.');
      return;
    }
    setAiGenerating(true);
    setFormError('');
    try {
      const res = await tranApi.post(
        tranEndpoints.storyPostAIGenerate,
        { prompt, image_count: Number(aiImageCount) || 0 },
        { timeout: 240000 }
      );
      const title = String(res.data?.title || '').trim();
      const content = String(res.data?.content || '').trim();
      if (title) setDraftTitle(title);
      if (content) setDraftContent(content);
      const images = Array.isArray(res.data?.images) ? res.data.images : [];
      const files = [];
      for (const img of images) {
        const b64 = img?.data_base64;
        if (!b64) continue;
        const bin = atob(b64);
        const bytes = new Uint8Array(bin.length);
        for (let i = 0; i < bin.length; i += 1) bytes[i] = bin.charCodeAt(i);
        const type = img.content_type || 'image/png';
        const name = img.filename || `story-ai-${files.length + 1}.png`;
        files.push(new File([bytes], name, { type }));
      }
      if (files.length) {
        setPickedMedia(files.slice(0, MAX_STORY_MEDIA));
      }
      const warnings = Array.isArray(res.data?.image_warnings) ? res.data.image_warnings : [];
      if (warnings.length) {
        setFormError(`Story text ready. Some images failed: ${warnings.join('; ')}`);
      }
    } catch (err) {
      setFormError(err.response?.data?.error || err.message || 'AI generation failed.');
    } finally {
      setAiGenerating(false);
    }
  };

  const submitPost = async () => {
    const title = draftTitle.trim();
    const content = draftContent.trim();
    if (!title) {
      setFormError('Title is required.');
      return;
    }
    if (!content) {
      setFormError('Content is required.');
      return;
    }
    const mediaList = Array.isArray(pickedMedia) ? pickedMedia : [];
    if (mediaList.length > MAX_STORY_MEDIA) {
      setFormError(`Up to ${MAX_STORY_MEDIA} media files are allowed.`);
      return;
    }
    for (const file of mediaList) {
      const mediaErr = validateMediaFile(file);
      if (mediaErr) {
        setFormError(mediaErr);
        return;
      }
    }

    setSubmitting(true);
    setFormError('');
    let savedPost = null;
    let postId = editing?.id;
    let mediaWarning = '';
    try {
      if (editing) {
        const res = await tranApi.put(tranEndpoints.storyPost(editing.id), { title, content });
        savedPost = normalizePost(res.data) || normalizePost({ ...editing, title, content });
        postId = savedPost?.id ?? editing.id;
      } else {
        const res = await tranApi.post(tranEndpoints.storyPosts, { title, content });
        savedPost = normalizePost(res.data) || normalizePost({ title, content });
        postId = savedPost?.id;
      }

      if (savedPost) {
        setPosts((prev) => mergePosts(prev, [savedPost]));
      }

      if (mediaList.length && postId) {
        try {
          if (editing?.attachments?.length) {
            await Promise.all(
              editing.attachments.map((att) =>
                tranApi.delete(tranEndpoints.entityAttachment('story-posts', postId, att.id))
              )
            );
          }
          const form = new FormData();
          mediaList.forEach((file) => form.append('files', file));
          const uploadRes = await tranApi.post(tranEndpoints.entityAttachments('story-posts', postId), form, {
            headers: { 'Content-Type': 'multipart/form-data' },
          });
          const attachments = uploadRes.data?.attachments;
          if (Array.isArray(attachments)) {
            const withMedia = { ...savedPost, id: postId, attachments };
            setPosts((prev) => mergePosts(prev, [withMedia]));
            savedPost = withMedia;
          }
        } catch (uploadErr) {
          mediaWarning = uploadErr.response?.data?.error || uploadErr.message || 'Media upload failed.';
        }
      }

      try {
        await load();
      } catch (refreshErr) {
        if (!savedPost) {
          throw refreshErr;
        }
      }

      setComposeOpen(false);
      setDraftTitle('');
      setDraftContent('');
      setPickedMedia([]);
      setAiPrompt('');
      setEditing(null);
      if (mediaWarning) {
        setError(`Post saved. ${mediaWarning}`);
      }
    } catch (err) {
      if (savedPost) {
        setError(err.response?.data?.error || err.message || 'Post saved, but refresh failed.');
        setComposeOpen(false);
      } else {
        setFormError(err.response?.data?.error || err.message || 'Failed to save post.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  const deletePost = async (post) => {
    const ok = await confirm({
      title: 'Delete story',
      message: `Delete "${post.title}"?`,
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    try {
      await tranApi.delete(tranEndpoints.storyPost(post.id));
      setPosts((prev) => prev.filter((p) => p.id !== post.id));
      setDetailOpen(false);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to delete story.');
    }
  };

  const downloadMedia = async (post, attachment) => {
    try {
      const url = tranEndpoints.entityAttachmentDownload('story-posts', post.id, attachment.id);
      const res = await tranApi.get(url, { responseType: 'blob' });
      const blobUrl = URL.createObjectURL(res.data);
      const a = document.createElement('a');
      a.href = blobUrl;
      a.download = attachment.original_name || 'download';
      a.rel = 'noopener';
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(blobUrl);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Download failed.');
    }
  };

  const openComments = (post) => {
    setCommentsPost(post);
    setCommentsOpen(true);
  };

  const sortedPosts = posts;

  if (error && !posts.length) return <Alert severity="error">{error}</Alert>;

  return (
    <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', gap: 1.5 }}>
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        justifyContent="space-between"
        alignItems={{ xs: 'stretch', sm: 'center' }}
        spacing={1}
      >
        <Box>
          <Typography variant="h6">Stories</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ display: { xs: 'none', sm: 'block' } }}>
            Task-like story tiles — open a card for details and notes.
          </Typography>
        </Box>
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ alignSelf: { xs: 'stretch', sm: 'auto' } }}>
          <Button
            variant="outlined"
            startIcon={<AutoAwesomeIcon />}
            onClick={openCreate}
            sx={{ alignSelf: { xs: 'stretch', sm: 'auto' } }}
          >
            Generate with AI
          </Button>
          <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} sx={{ alignSelf: { xs: 'stretch', sm: 'auto' } }}>
            New story
          </Button>
        </Stack>
      </Stack>

      {error ? (
        <Alert severity="warning" onClose={() => setError('')}>
          {error}
        </Alert>
      ) : null}

      {loading ? (
        <CircularProgress sx={{ mt: 2 }} />
      ) : sortedPosts.length === 0 ? (
        <Card variant="outlined">
          <CardContent>
            <Typography color="text.secondary">No stories yet. Create the first one.</Typography>
          </CardContent>
        </Card>
      ) : (
        <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', pr: 0.5 }}>
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: {
                xs: '1fr',
                sm: 'repeat(2, minmax(0, 1fr))',
                md: 'repeat(3, minmax(0, 1fr))',
              },
              gap: 1.5,
              alignItems: 'stretch',
            }}
          >
            {sortedPosts.map((post) => (
              <Card key={post.id} variant="outlined" sx={{ height: 170, display: 'flex' }}>
                <CardActionArea sx={{ height: 1, p: 0 }} onClick={() => openDetail(post)}>
                  <CardContent sx={{ height: 1, display: 'flex', flexDirection: 'column', gap: 1 }}>
                    <Typography
                      variant="subtitle1"
                      sx={{
                        fontWeight: 600,
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                      }}
                      title={post.title}
                    >
                      {post.title || 'Untitled'}
                    </Typography>
                    <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                      <Typography variant="body2" color="text.secondary" noWrap title={post.content || ''}>
                        {post.content || 'No content'}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        {dateTimeDisplay(post.created_on)}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        {post.comment_count
                          ? `${post.comment_count} note${post.comment_count === 1 ? '' : 's'}`
                          : 'No notes'}
                      </Typography>
                    </Box>
                  </CardContent>
                </CardActionArea>
              </Card>
            ))}
          </Box>
        </Box>
      )}

      <Drawer
        anchor="right"
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        PaperProps={{
          sx: {
            width: { xs: '100vw', sm: '66.666vw' },
            maxWidth: '100vw',
            height: '100%',
            display: 'flex',
            flexDirection: 'column',
            pt: { xs: 'env(safe-area-inset-top)', sm: 0 },
            pb: { xs: 'env(safe-area-inset-bottom)', sm: 0 },
          },
        }}
      >
        <Box sx={{ px: 2.5, py: 2, borderBottom: '1px solid', borderColor: 'divider', flexShrink: 0 }}>
          <Typography variant="h6">{detailPost?.title || 'Story'}</Typography>
          {detailPost ? (
            <Stack direction="row" spacing={1.5} alignItems="center" sx={{ mt: 1.5 }}>
              <Avatar sx={{ bgcolor: 'primary.main', width: 36, height: 36 }}>
                {authorInitials(detailPost.author_name)}
              </Avatar>
              <Box>
                <Typography variant="subtitle2" fontWeight={700}>
                  {detailPost.author_name || `User #${detailPost.author_user_id}`}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {dateTimeDisplay(detailPost.created_on)}
                </Typography>
              </Box>
            </Stack>
          ) : null}
        </Box>
        <Box
          sx={{
            flex: 1,
            minHeight: 0,
            display: 'flex',
            flexDirection: 'column',
            gap: 2,
            px: 2.5,
            py: 2,
            overflow: 'hidden',
          }}
        >
          {detailPost ? (
            <>
              <Typography
                variant="body1"
                component="div"
                sx={{
                  whiteSpace: 'pre-wrap',
                  flex: 1,
                  minHeight: 0,
                  overflowY: 'auto',
                  pr: 0.5,
                }}
              >
                {detailPost.content}
              </Typography>
              {(detailPost.attachments || []).filter((a) => a.kind === 'image' || a.kind === 'video').length ? (
                <Box
                  sx={{
                    flexShrink: 0,
                    display: 'grid',
                    gridTemplateColumns: {
                      xs: '1fr',
                      sm: 'repeat(2, minmax(0, 1fr))',
                    },
                    gap: 1,
                    maxHeight: 320,
                    overflowY: 'auto',
                  }}
                >
                  {(detailPost.attachments || [])
                    .filter((a) => a.kind === 'image' || a.kind === 'video')
                    .map((media) => (
                      <PostMediaPreview key={media.id} postId={detailPost.id} attachment={media} />
                    ))}
                </Box>
              ) : null}
            </>
          ) : null}
        </Box>
        {detailPost ? (
          <Box
            sx={{
              px: 2.5,
              py: 1.5,
              borderTop: '1px solid',
              borderColor: 'divider',
              flexShrink: 0,
            }}
          >
            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
              <Button size="small" startIcon={<ChatBubbleOutlineIcon />} onClick={() => openComments(detailPost)}>
                {detailPost.comment_count
                  ? `${detailPost.comment_count} note${detailPost.comment_count === 1 ? '' : 's'}`
                  : 'Notes'}
              </Button>
              <Button size="small" startIcon={<EditOutlinedIcon />} onClick={() => openEdit(detailPost)}>
                Edit
              </Button>
              <Button size="small" color="error" startIcon={<DeleteOutlineIcon />} onClick={() => deletePost(detailPost)}>
                Delete
              </Button>
              {(detailPost.attachments || [])
                .filter((a) => a.kind === 'image' || a.kind === 'video')
                .map((media) => (
                  <Button
                    key={`dl-${media.id}`}
                    size="small"
                    startIcon={<DownloadIcon />}
                    onClick={() => downloadMedia(detailPost, media)}
                  >
                    Download media
                  </Button>
                ))}
            </Stack>
          </Box>
        ) : null}
      </Drawer>

      <Drawer
        anchor="right"
        open={composeOpen}
        onClose={() => !submitting && setComposeOpen(false)}
        PaperProps={{
          sx: {
            width: { xs: '100vw', sm: '66.666vw' },
            maxWidth: '100vw',
            height: '100%',
            display: 'flex',
            flexDirection: 'column',
            pt: { xs: 'env(safe-area-inset-top)', sm: 0 },
            pb: { xs: 'env(safe-area-inset-bottom)', sm: 0 },
          },
        }}
      >
        <Box sx={{ px: 2.5, py: 2, borderBottom: '1px solid', borderColor: 'divider', flexShrink: 0 }}>
          <Typography variant="h6">{editing ? 'Edit story' : 'New story'}</Typography>
        </Box>
        <Box
          sx={{
            flex: 1,
            minHeight: 0,
            display: 'flex',
            flexDirection: 'column',
            gap: 1.5,
            p: 2.5,
            overflow: 'hidden',
          }}
        >
          {!editing ? (
            <Box
              sx={{
                flexShrink: 0,
                p: 1.5,
                borderRadius: 1,
                border: '1px solid',
                borderColor: 'divider',
                bgcolor: 'action.hover',
                display: 'grid',
                gap: 1.25,
              }}
            >
              <Typography variant="subtitle2">Generate with AI</Typography>
              <TextField
                label="Prompt"
                placeholder="e.g. A morning bus route welcome story for new drivers"
                fullWidth
                size="small"
                multiline
                minRows={2}
                value={aiPrompt}
                onChange={(e) => setAiPrompt(e.target.value)}
                disabled={aiGenerating || submitting}
              />
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ sm: 'center' }}>
                <FormControl size="small" sx={{ minWidth: 140 }}>
                  <InputLabel id="story-ai-images-label">Images</InputLabel>
                  <Select
                    labelId="story-ai-images-label"
                    label="Images"
                    value={aiImageCount}
                    onChange={(e) => setAiImageCount(Number(e.target.value))}
                    disabled={aiGenerating || submitting}
                  >
                    <MenuItem value={0}>No images</MenuItem>
                    <MenuItem value={1}>1 image</MenuItem>
                    <MenuItem value={2}>2 images</MenuItem>
                    <MenuItem value={3}>3 images</MenuItem>
                    <MenuItem value={4}>4 images</MenuItem>
                  </Select>
                </FormControl>
                <Button
                  variant="contained"
                  color="secondary"
                  startIcon={aiGenerating ? <CircularProgress size={16} color="inherit" /> : <AutoAwesomeIcon />}
                  onClick={() => void generateWithAI()}
                  disabled={aiGenerating || submitting}
                >
                  {aiGenerating ? 'Generating…' : 'Generate draft'}
                </Button>
              </Stack>
              <Typography variant="caption" color="text.secondary">
                Fills title and content below, and attaches generated images you can edit before saving.
              </Typography>
            </Box>
          ) : null}
          <TextField
            label="Title"
            required
            fullWidth
            size="small"
            value={draftTitle}
            onChange={(e) => setDraftTitle(e.target.value)}
            sx={{ flexShrink: 0 }}
            disabled={aiGenerating}
          />
          <TextField
            label="Content"
            required
            fullWidth
            multiline
            value={draftContent}
            onChange={(e) => setDraftContent(e.target.value)}
            disabled={aiGenerating}
            sx={{
              flex: 1,
              minHeight: 0,
              display: 'flex',
              flexDirection: 'column',
              '& .MuiInputBase-root': {
                flex: 1,
                alignItems: 'stretch',
                overflow: 'hidden',
              },
              '& textarea': {
                height: '100% !important',
                overflowY: 'auto !important',
                resize: 'none',
              },
            }}
            InputProps={{
              sx: { height: '100%', alignItems: 'stretch' },
            }}
          />
          <Box sx={{ flexShrink: 0 }}>
            <Typography variant="subtitle2" sx={{ mb: 0.75 }}>
              Media (optional)
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
              Up to {MAX_STORY_MEDIA} images (under 20 MB each) or videos (under 200 MB).
            </Typography>
            <Button component="label" variant="outlined" size="small" disabled={submitting || aiGenerating}>
              {pickedMedia.length ? 'Add more files' : 'Choose image or video'}
              <input
                type="file"
                hidden
                multiple
                accept={MEDIA_ACCEPT}
                onChange={(e) => {
                  const files = Array.from(e.target.files || []);
                  if (!files.length) {
                    e.target.value = '';
                    return;
                  }
                  const next = [...pickedMedia];
                  for (const file of files) {
                    const mediaErr = validateMediaFile(file);
                    if (mediaErr) {
                      setFormError(mediaErr);
                      e.target.value = '';
                      return;
                    }
                    if (next.length >= MAX_STORY_MEDIA) {
                      setFormError(`Up to ${MAX_STORY_MEDIA} media files are allowed.`);
                      break;
                    }
                    next.push(file);
                  }
                  setFormError('');
                  setPickedMedia(next);
                  e.target.value = '';
                }}
              />
            </Button>
            {pickedMedia.length ? (
              <Stack spacing={0.5} sx={{ mt: 1 }}>
                {pickedMedia.map((file, idx) => (
                  <Stack key={`${file.name}-${idx}`} direction="row" spacing={1} alignItems="center">
                    <Typography variant="body2" sx={{ flex: 1, minWidth: 0 }} noWrap>
                      {file.name}
                    </Typography>
                    <Button
                      size="small"
                      onClick={() => setPickedMedia((prev) => prev.filter((_, i) => i !== idx))}
                      disabled={submitting || aiGenerating}
                    >
                      Remove
                    </Button>
                  </Stack>
                ))}
              </Stack>
            ) : editing?.attachments?.length ? (
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Current media kept unless you upload new files (new uploads replace existing media).
              </Typography>
            ) : null}
          </Box>
          {formError ? <Alert severity="error" sx={{ flexShrink: 0 }}>{formError}</Alert> : null}
        </Box>
        <Box
          sx={{
            px: 2.5,
            py: 1.5,
            borderTop: '1px solid',
            borderColor: 'divider',
            display: 'flex',
            justifyContent: 'flex-end',
            gap: 1,
            flexShrink: 0,
          }}
        >
          <Button onClick={() => setComposeOpen(false)} disabled={submitting || aiGenerating}>
            Cancel
          </Button>
          <Button variant="contained" onClick={submitPost} disabled={submitting || aiGenerating}>
            {submitting ? 'Saving…' : editing ? 'Save' : 'Create'}
          </Button>
        </Box>
      </Drawer>

      <CommentsModal
        open={commentsOpen}
        onClose={() => setCommentsOpen(false)}
        entityType="story_post"
        recordId={commentsPost?.id}
        recordLabel={commentsPost?.title}
        onThreadChanged={handleCommentThreadChanged}
        noun="note"
      />
    </Box>
  );
}

