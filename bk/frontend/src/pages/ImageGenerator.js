import React, { useState, useEffect, useMemo } from 'react';
import {
  Box,
  Typography,
  TextField,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
  Tooltip,
  Paper,
  ImageList,
  ImageListItem,
  ImageListItemBar,
  Chip,
  Popover,
  Stack,
} from '@mui/material';
import {
  Image as ImageIcon,
  AutoFixHigh as PolishIcon,
  Download as DownloadIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  Close as CloseIcon,
  ContentCopy as CopyIcon,
  Tune as TuneIcon,
} from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import api from '../services/api';
import ModuleShell from '../components/ModuleShell';

const ImageGenerator = () => {
  const queryClient = useQueryClient();
  const [prompt, setPrompt] = useState('');
  const [polishedPrompt, setPolishedPrompt] = useState('');
  const [usePolished, setUsePolished] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState('qwen');
  const [selectedModel, setSelectedModel] = useState('');
  const [generatedImageUrl, setGeneratedImageUrl] = useState(null);
  const [imageDialogOpen, setImageDialogOpen] = useState(false);
  const [selectedImage, setSelectedImage] = useState(null);
  const [error, setError] = useState(null);
  const [isPolishing, setIsPolishing] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [polishAnchor, setPolishAnchor] = useState(null);

  // Fetch providers
  const { data: providersData = { providers: [] } } = useQuery('providers', api.getProviders);
  const providers = providersData.providers || [];

  // Fetch models and transform to {provider: [model1, model2, ...]} format
  const { data: modelsData = [] } = useQuery('models', api.getModels);
  const models = useMemo(() => {
    const modelsByProvider = {};
    (Array.isArray(modelsData) ? modelsData : []).forEach((model) => {
      const provider = model.provider;
      if (!modelsByProvider[provider]) {
        modelsByProvider[provider] = [];
      }
      if (model.name && !modelsByProvider[provider].includes(model.name)) {
        modelsByProvider[provider].push(model.name);
      }
    });
    return modelsByProvider;
  }, [modelsData]);

  // Fetch saved images
  const { data: savedImages = [], isLoading: imagesLoading, refetch: refetchImages } = useQuery(
    'generatedImages',
    api.getGeneratedImages,
    {
      refetchOnWindowFocus: false,
    }
  );

  // Set default model when provider changes
  useEffect(() => {
    if (models[selectedProvider] && models[selectedProvider].length > 0) {
      setSelectedModel(models[selectedProvider][0]);
    }
  }, [selectedProvider, models]);

  // Polish prompt mutation
  const polishPromptMutation = useMutation(
    (data) => api.polishImagePrompt(data),
    {
      onSuccess: (data) => {
        setPolishedPrompt(data.polished_prompt);
        setUsePolished(true);
        setIsPolishing(false);
      },
      onError: (err) => {
        setError(err.response?.data?.detail || 'Failed to polish prompt');
        setIsPolishing(false);
      },
    }
  );

  // Generate image mutation
  const generateImageMutation = useMutation(
    (data) => api.generateImage(data),
    {
      onSuccess: (data) => {
        setGeneratedImageUrl(data.image_url);
        setImageDialogOpen(true);
        setIsGenerating(false);
        refetchImages();
      },
      onError: (err) => {
        setError(err.response?.data?.detail || 'Failed to generate image');
        setIsGenerating(false);
      },
    }
  );

  // Delete image mutation
  const deleteImageMutation = useMutation(
    (filename) => api.deleteGeneratedImage(filename),
    {
      onSuccess: () => {
        refetchImages();
      },
      onError: (err) => {
        setError(err.response?.data?.detail || 'Failed to delete image');
      },
    }
  );

  const handlePolishPrompt = () => {
    if (!prompt.trim()) {
      setError('Please enter a prompt first');
      return;
    }
    setError(null);
    setIsPolishing(true);
    polishPromptMutation.mutate({
      prompt: prompt,
      provider: selectedProvider,
      model: selectedModel,
    });
  };

  const handleGenerateImage = () => {
    const finalPrompt = usePolished && polishedPrompt ? polishedPrompt : prompt;
    if (!finalPrompt.trim()) {
      setError('Please enter a prompt');
      return;
    }
    setError(null);
    setIsGenerating(true);
    generateImageMutation.mutate({
      prompt: finalPrompt,
      save: true,
    });
  };

  const handleImageClick = (image) => {
    setSelectedImage(image);
    setImageDialogOpen(true);
  };

  const handleDeleteImage = (filename, e) => {
    e.stopPropagation();
    if (window.confirm('Delete this image?')) {
      deleteImageMutation.mutate(filename);
    }
  };

  const handleDownloadImage = async (imageUrl, filename) => {
    try {
      const response = await fetch(imageUrl);
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename || 'generated-image.png';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      setError('Failed to download image');
    }
  };

  const handleCopyUrl = (url) => {
    navigator.clipboard.writeText(url);
  };

  return (
    <ModuleShell
      title="Image generator"
      helpText="Creates images via Pollinations. Optional prompt polish uses your selected LLM to refine the text before generation. Gallery shows server-stored outputs."
    >
    <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', width: '100%', pb: 1 }}>
      {error && (
        <Alert severity="error" sx={{ mb: 1.25, flexShrink: 0 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {/* Compact prompt + actions */}
      <Paper variant="outlined" sx={{ mb: 1, flexShrink: 0, px: 1.25, py: 1 }}>
        <Stack spacing={0.75}>
          <TextField
            fullWidth
            multiline
            minRows={1}
            maxRows={2}
            size="small"
            label="Prompt"
            placeholder="Describe the image…"
            value={prompt}
            onChange={(e) => {
              setPrompt(e.target.value);
              setUsePolished(false);
            }}
            sx={{
              '& .MuiOutlinedInput-root': { fontSize: '0.82rem' },
              '& .MuiInputLabel-root': { fontSize: '0.8rem' },
            }}
          />

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, flexWrap: 'wrap' }}>
            <Button
              size="small"
              variant="contained"
              startIcon={isGenerating ? <CircularProgress size={14} color="inherit" /> : <ImageIcon fontSize="small" />}
              onClick={handleGenerateImage}
              disabled={isGenerating || (!prompt.trim() && !polishedPrompt)}
              sx={{ minWidth: 120, py: 0.5, fontSize: '0.78rem' }}
            >
              Generate
            </Button>

            <Button
              size="small"
              variant="outlined"
              color="secondary"
              startIcon={isPolishing ? <CircularProgress size={14} /> : <PolishIcon fontSize="small" />}
              onClick={handlePolishPrompt}
              disabled={isPolishing || !prompt.trim()}
              sx={{ py: 0.5, fontSize: '0.78rem' }}
            >
              Polish
            </Button>

            <Tooltip title="Polish LLM settings">
              <IconButton
                size="small"
                onClick={(e) => setPolishAnchor(e.currentTarget)}
                sx={{ border: '1px solid', borderColor: 'divider', p: 0.5 }}
              >
                <TuneIcon sx={{ fontSize: 18 }} />
              </IconButton>
            </Tooltip>

            <Typography variant="caption" color="text.secondary" sx={{ ml: 'auto', fontSize: '0.68rem' }}>
              {selectedProvider}/{selectedModel || 'default'}
            </Typography>
          </Box>

          <Popover
            open={Boolean(polishAnchor)}
            anchorEl={polishAnchor}
            onClose={() => setPolishAnchor(null)}
            anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
            transformOrigin={{ vertical: 'top', horizontal: 'left' }}
            slotProps={{ paper: { sx: { p: 1.25, width: 260 } } }}
          >
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.75 }}>
              Used only when polishing prompts
            </Typography>
            <Stack spacing={1}>
              <FormControl size="small" fullWidth>
                <InputLabel>Provider</InputLabel>
                <Select
                  value={selectedProvider}
                  label="Provider"
                  onChange={(e) => setSelectedProvider(e.target.value)}
                  MenuProps={{ PaperProps: { sx: { maxHeight: 240 } } }}
                >
                  {providers.map((provider) => (
                    <MenuItem key={provider} value={provider} sx={{ fontSize: '0.8rem' }}>
                      {provider.charAt(0).toUpperCase() + provider.slice(1)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              <FormControl size="small" fullWidth>
                <InputLabel>Model</InputLabel>
                <Select
                  value={selectedModel}
                  label="Model"
                  onChange={(e) => setSelectedModel(e.target.value)}
                  MenuProps={{ PaperProps: { sx: { maxHeight: 280 } } }}
                >
                  {(models[selectedProvider] || []).map((model) => (
                    <MenuItem key={model} value={model} sx={{ fontSize: '0.78rem' }}>
                      {model}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Stack>
          </Popover>

          {polishedPrompt && (
            <Box
              sx={{
                px: 0.75,
                py: 0.5,
                borderRadius: 1,
                border: '1px solid',
                borderColor: usePolished ? 'primary.main' : 'divider',
                bgcolor: usePolished ? (t) => `${t.palette.primary.main}14` : 'transparent',
                cursor: 'pointer',
              }}
              onClick={() => setUsePolished(!usePolished)}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
                <Typography variant="caption" color="primary" sx={{ fontWeight: 600 }}>
                  Polished
                </Typography>
                <Chip
                  size="small"
                  label={usePolished ? 'Using' : 'Use'}
                  color={usePolished ? 'primary' : 'default'}
                  sx={{ height: 18, fontSize: '0.62rem' }}
                />
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, lineHeight: 1.3 }}>
                  {polishedPrompt.length > 120 ? `${polishedPrompt.substring(0, 120)}…` : polishedPrompt}
                </Typography>
              </Box>
            </Box>
          )}
        </Stack>
      </Paper>

      {/* Generated Images Gallery — fills remaining height, scroll inside */}
      <Card sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <CardContent sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', pt: 1.5, pb: 1, '&:last-child': { pb: 1 } }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1, flexShrink: 0 }}>
            <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <ImageIcon /> Generated Images
            </Typography>
            <Tooltip title="Refresh">
              <IconButton onClick={() => refetchImages()} size="small">
                <RefreshIcon />
              </IconButton>
            </Tooltip>
          </Box>

          {imagesLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          ) : savedImages.length === 0 ? (
            <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', py: 4 }}>
              No images generated yet. Create your first image above!
            </Typography>
          ) : (
            <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', overflowX: 'hidden', pr: 0.25 }}>
              <ImageList cols={3} gap={16} sx={{ m: 0 }}>
              {savedImages.map((image) => (
                <ImageListItem 
                  key={image.filename}
                  sx={{ 
                    cursor: 'pointer',
                    borderRadius: 2,
                    overflow: 'hidden',
                    border: (t) => `1px solid ${t.palette.primary.main}33`,
                    transition: 'all 0.3s',
                    '&:hover': {
                      transform: 'scale(1.02)',
                      borderColor: 'primary.main',
                      boxShadow: (t) => `0 4px 20px ${t.palette.primary.main}4d`,
                    },
                  }}
                  onClick={() => handleImageClick(image)}
                >
                  <img
                    src={image.url}
                    alt={image.prompt || 'Generated image'}
                    loading="lazy"
                    style={{ 
                      height: 260, 
                      objectFit: 'cover',
                    }}
                  />
                  <ImageListItemBar
                    title={image.prompt ? (image.prompt.length > 40 ? image.prompt.substring(0, 40) + '...' : image.prompt) : 'Image'}
                    subtitle={image.created_at}
                    actionIcon={
                      <Box sx={{ display: 'flex', gap: 0.5, pr: 1 }}>
                        <Tooltip title="Download">
                          <IconButton
                            size="small"
                            sx={{ color: 'white' }}
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDownloadImage(image.url, image.filename);
                            }}
                          >
                            <DownloadIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Delete">
                          <IconButton
                            size="small"
                            sx={{ color: 'white' }}
                            onClick={(e) => handleDeleteImage(image.filename, e)}
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </Box>
                    }
                    sx={{
                      background: 'linear-gradient(to top, rgba(0,0,0,0.9) 0%, rgba(0,0,0,0) 100%)',
                    }}
                  />
                </ImageListItem>
              ))}
            </ImageList>
            </Box>
          )}
        </CardContent>
      </Card>

      {/* Image Preview Dialog */}
      <Dialog 
        open={imageDialogOpen} 
        onClose={() => {
          setImageDialogOpen(false);
          setSelectedImage(null);
          setGeneratedImageUrl(null);
        }}
        maxWidth="lg"
        fullWidth
      >
        <DialogTitle sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Typography variant="h6">
            {selectedImage ? 'Image Details' : 'Generated Image'}
          </Typography>
          <IconButton onClick={() => {
            setImageDialogOpen(false);
            setSelectedImage(null);
            setGeneratedImageUrl(null);
          }}>
            <CloseIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent>
          <Box sx={{ textAlign: 'center' }}>
            <img
              src={selectedImage?.url || generatedImageUrl}
              alt="Generated"
              style={{ 
                maxWidth: '100%', 
                maxHeight: '60vh',
                borderRadius: 8,
                boxShadow: '0 4px 20px rgba(0,0,0,0.5)',
              }}
            />
            {(selectedImage?.prompt || (usePolished ? polishedPrompt : prompt)) && (
              <Paper variant="outlined" sx={{ mt: 2, p: 2, textAlign: 'left' }}>
                <Typography variant="subtitle2" color="primary" gutterBottom>
                  Prompt:
                </Typography>
                <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>
                  {selectedImage?.prompt || (usePolished ? polishedPrompt : prompt)}
                </Typography>
              </Paper>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button
            startIcon={<CopyIcon />}
            onClick={() => handleCopyUrl(selectedImage?.url || generatedImageUrl)}
          >
            Copy URL
          </Button>
          <Button
            startIcon={<DownloadIcon />}
            onClick={() => handleDownloadImage(
              selectedImage?.url || generatedImageUrl, 
              selectedImage?.filename || 'generated-image.png'
            )}
          >
            Download
          </Button>
          <Button onClick={() => {
            setImageDialogOpen(false);
            setSelectedImage(null);
            setGeneratedImageUrl(null);
          }}>
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
    </ModuleShell>
  );
};

export default ImageGenerator;
