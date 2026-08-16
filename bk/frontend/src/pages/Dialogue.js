import React, { useState } from 'react';
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
  Autocomplete,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  Paper,
  Tabs,
  Tab,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  Send as SendIcon,
} from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import ReactMarkdown from 'react-markdown';
import api from '../services/api';
import ModuleShell from '../components/ModuleShell';
import Conversation from '../components/Conversation';

const Dialogue = () => {
  const queryClient = useQueryClient();
  const [tabValue, setTabValue] = useState(0);
  const { data: dialogues = [], isLoading, error } = useQuery('dialogues', api.getDialogues, { staleTime: 5 * 60 * 1000 }); // Cache for 5 minutes
  const { data: models = [] } = useQuery('models', api.getModels, { enabled: true, staleTime: 5 * 60 * 1000 });
  const { data: providersData } = useQuery('providers', api.getProviders, { enabled: true, staleTime: 5 * 60 * 1000 });
  const { data: collections = [] } = useQuery('collections', api.getRAGCollections, { staleTime: 5 * 60 * 1000 });
  const { data: requestTools = [] } = useQuery('request-tools', api.getRequestTools, { staleTime: 5 * 60 * 1000 });
  const [openCreateDialog, setOpenCreateDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [editingDialogueId, setEditingDialogueId] = useState(null);
  const [selectedDialogue, setSelectedDialogue] = useState(null);
  const [deleteConfirmDialog, setDeleteConfirmDialog] = useState({ open: false, dialogueId: null });
  const [createForm, setCreateForm] = useState({
    name: '',
    description: '',
    system_prompt: '',
    rag_collection: '',
    request_tools: [],
    llm_provider: '',
    model_name: '',
    max_turns: 5,
  });

  // Conversation state
  const [conversationId, setConversationId] = useState(null);
  const [conversationHistory, setConversationHistory] = useState([]);
  const [currentMessage, setCurrentMessage] = useState('');
  const [isWaitingForResponse, setIsWaitingForResponse] = useState(false);
  const [conversationError, setConversationError] = useState('');
  const [conversationMeta, setConversationMeta] = useState(null);
  const [isComplete, setIsComplete] = useState(false);

  const createMutation = useMutation(api.createDialogue, {
    onSuccess: () => {
      queryClient.invalidateQueries('dialogues');
      setOpenCreateDialog(false);
      setCreateForm({
        name: '',
        description: '',
        system_prompt: '',
        rag_collection: '',
        request_tools: [],
        llm_provider: '',
        model_name: '',
        max_turns: 5,
      });
    },
  });

  const updateMutation = useMutation(
    ({ dialogueId, payload }) => api.updateDialogue(dialogueId, payload),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('dialogues');
        setOpenEditDialog(false);
        setEditingDialogueId(null);
        setCreateForm({
          name: '',
          description: '',
          system_prompt: '',
          rag_collection: '',
          request_tools: [],
          llm_provider: '',
          model_name: '',
          max_turns: 5,
        });
      },
    }
  );

  const deleteMutation = useMutation(api.deleteDialogue, {
    onSuccess: () => {
      queryClient.invalidateQueries('dialogues');
      if (selectedDialogue && selectedDialogue.id === editingDialogueId) {
        setSelectedDialogue(null);
        resetConversation();
      }
    },
  });

  const resetConversation = () => {
    setConversationId(null);
    setConversationHistory([]);
    setCurrentMessage('');
    setIsComplete(false);
    setConversationError('');
    setConversationMeta(null);
  };

  const handleStartConversation = async () => {
    if (!selectedDialogue || !currentMessage.trim()) return;

    setIsWaitingForResponse(true);
    setConversationError('');
    resetConversation();

    try {
      const res = await api.startDialogue(selectedDialogue.id, {
        initial_message: currentMessage,
      });

      setConversationId(res.conversation_id);
      setConversationHistory(res.conversation_history || []);
      setIsComplete(res.is_complete);
      setConversationMeta({
        model_used: res.model_used,
        rag_collection_used: res.rag_collection_used,
        turn_number: res.turn_number,
        max_turns: res.max_turns,
      });
      setCurrentMessage('');
    } catch (err) {
      console.error('Dialogue start error:', err);
      setConversationError(err.response?.data?.detail || err.message || 'Failed to start dialogue. Please try again.');
    } finally {
      setIsWaitingForResponse(false);
    }
  };

  const handleContinueConversation = async () => {
    if (!selectedDialogue || !conversationId || !currentMessage.trim()) return;

    setIsWaitingForResponse(true);
    setConversationError('');

    try {
      const res = await api.continueDialogue(selectedDialogue.id, {
        conversation_id: conversationId,
        user_message: currentMessage,
      });

      setConversationHistory(res.conversation_history || []);
      setIsComplete(res.is_complete);
      setConversationMeta({
        model_used: res.model_used,
        rag_collection_used: res.rag_collection_used,
        turn_number: res.turn_number,
        max_turns: res.max_turns,
      });
      setCurrentMessage('');
    } catch (err) {
      console.error('Dialogue continue error:', err);
      setConversationError(err.response?.data?.detail || err.message || 'Failed to continue dialogue. Please try again.');
    } finally {
      setIsWaitingForResponse(false);
    }
  };

  const handleSendMessage = () => {
    if (conversationId) {
      handleContinueConversation();
    } else {
      handleStartConversation();
    }
  };

  const handleCreate = () => {
    createMutation.mutate({
      name: createForm.name,
      description: createForm.description,
      system_prompt: createForm.system_prompt,
      rag_collection: createForm.rag_collection || null,
      db_tools: [],
      request_tools: createForm.request_tools || [],
      llm_provider: createForm.llm_provider || null,
      model_name: createForm.model_name || null,
      max_turns: createForm.max_turns,
    });
  };

  const handleEditDialogue = (dialogue) => {
    setEditingDialogueId(dialogue.id);
    setCreateForm({
      name: dialogue.name || '',
      description: dialogue.description || '',
      system_prompt: dialogue.system_prompt || '',
      rag_collection: dialogue.rag_collection || '',
      request_tools: dialogue.request_tools || [],
      llm_provider: dialogue.llm_provider || '',
      model_name: dialogue.model_name || '',
      max_turns: dialogue.max_turns || 5,
    });
    setOpenEditDialog(true);
  };

  const handleUpdate = () => {
    updateMutation.mutate({
      dialogueId: editingDialogueId,
      payload: {
        name: createForm.name,
        description: createForm.description,
        system_prompt: createForm.system_prompt,
        rag_collection: createForm.rag_collection || null,
        db_tools: [],
        request_tools: createForm.request_tools || [],
        llm_provider: createForm.llm_provider || null,
        model_name: createForm.model_name || null,
        max_turns: createForm.max_turns,
      },
    });
  };

  const handleDeleteDialogue = (dialogueId) => {
    setDeleteConfirmDialog({ open: true, dialogueId });
  };

  const handleDeleteConfirm = () => {
    deleteMutation.mutate(deleteConfirmDialog.dialogueId);
  };

  const handleSelectDialogue = (dialogue) => {
    setSelectedDialogue(dialogue);
    resetConversation();
  };

  return (
    <ModuleShell
      title="Dialogue"
      helpText="Dialogues: named configurations with system prompt, optional RAG collection, and HTTP request tools for multi-turn chat. Conversations: two-model scripted back-and-forth."
      subtitleBar={
        <Tabs
          value={tabValue}
          onChange={(e, newValue) => setTabValue(newValue)}
          variant="scrollable"
          scrollButtons="auto"
        >
          <Tab label="Dialogues" sx={{ fontSize: '0.65rem', minHeight: 36, py: 0 }} />
          <Tab label="Conversations" sx={{ fontSize: '0.65rem', minHeight: 36, py: 0 }} />
        </Tabs>
      }
    >
      <Box sx={{ flex: 1, minHeight: 0, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
      {tabValue === 0 && (
        <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', pb: 1 }}>
          <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 1, flexShrink: 0 }}>
            <Button
              variant="contained"
              size="small"
              startIcon={<AddIcon />}
              onClick={() => setOpenCreateDialog(true)}
            >
              New Dialogue
            </Button>
          </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 1, flexShrink: 0 }}>
          Failed to load dialogues
        </Alert>
      )}

      <Grid container spacing={2} sx={{ flex: 1, minHeight: 0, overflow: 'hidden', alignItems: 'stretch', '& > .MuiGrid-item': { display: 'flex', flexDirection: 'column', minHeight: 0 } }}>
        <Grid item xs={12} md={4} sx={{ display: 'flex', flexDirection: 'column', minHeight: 0 }}>
          <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', pr: 0.25 }}>
          <Grid container spacing={1.25}>
            {dialogues.map((dialogue) => (
              <Grid item xs={12} key={dialogue.id}>
                <Card
                  sx={{
                    boxShadow: 2,
                    cursor: 'pointer',
                    border:
                      selectedDialogue && selectedDialogue.id === dialogue.id
                        ? '2px solid #1976d2'
                        : '1px solid rgba(0,0,0,0.12)',
                  }}
                  onClick={() => handleSelectDialogue(dialogue)}
                >
                  <CardContent sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <Box sx={{ flex: 1, pr: 1 }}>
                      <Typography variant="h6">{dialogue.name}</Typography>
                      {dialogue.description && (
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                          {dialogue.description}
                        </Typography>
                      )}
                      <Box sx={{ mt: 1, display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                        {dialogue.rag_collection && (
                          <Chip size="small" label={`RAG: ${dialogue.rag_collection}`} color="primary" variant="outlined" />
                        )}
                        {dialogue.request_tools && dialogue.request_tools.length > 0 && (
                          <Chip size="small" label={`Request Tools: ${dialogue.request_tools.length}`} color="info" variant="outlined" />
                        )}
                        {dialogue.llm_provider && (
                          <Chip size="small" label={`Provider: ${dialogue.llm_provider}`} variant="outlined" />
                        )}
                        {dialogue.model_name && (
                          <Chip size="small" label={`Model: ${dialogue.model_name}`} variant="outlined" />
                        )}
                        <Chip size="small" label={`Max Turns: ${dialogue.max_turns}`} variant="outlined" />
                      </Box>
                    </Box>
                    <Box sx={{ display: 'flex', gap: 0.5 }}>
                      <IconButton
                        size="small"
                        color="primary"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleEditDialogue(dialogue);
                        }}
                      >
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton
                        size="small"
                        color="error"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteDialogue(dialogue.id);
                        }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Box>
                  </CardContent>
                </Card>
              </Grid>
            ))}
            {!isLoading && dialogues.length === 0 && (
              <Grid item xs={12}>
                <Typography variant="body2" color="text.secondary">
                  No dialogues yet. Create one to get started.
                </Typography>
              </Grid>
            )}
          </Grid>
          </Box>
        </Grid>

        {/* Conversation Panel — fills viewport; transcript scrolls inside */}
        <Grid item xs={12} md={8} sx={{ display: 'flex', flexDirection: 'column', minHeight: 0 }}>
          <Card sx={{ boxShadow: 2, flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', height: '100%' }}>
            <CardContent sx={{ p: 1.75, flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', '&:last-child': { pb: 1.75 } }}>
              <Typography variant="subtitle2" gutterBottom sx={{ flexShrink: 0, fontWeight: 700 }}>
                Conversation
              </Typography>
              {selectedDialogue ? (
                <>
                  <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 0.35, flexShrink: 0 }}>
                    {selectedDialogue.name}
                  </Typography>
                  <Typography
                    variant="caption"
                    color="text.secondary"
                    sx={{
                      mb: 1,
                      whiteSpace: 'pre-wrap',
                      maxHeight: 56,
                      overflowY: 'auto',
                      flexShrink: 0,
                      border: '1px solid rgba(255,255,255,0.12)',
                      borderRadius: 1,
                      p: 0.75,
                      display: 'block',
                    }}
                  >
                    {selectedDialogue.system_prompt}
                  </Typography>

                  {/* Conversation History */}
                  <Box
                    sx={{
                      flex: 1,
                      minHeight: 0,
                      overflow: 'hidden',
                      display: 'flex',
                      flexDirection: 'column',
                      bgcolor: 'background.paper',
                      border: '1px solid',
                      borderColor: 'primary.main',
                      borderOpacity: 0.3,
                      borderRadius: 1,
                    }}
                  >
                    <Box
                      sx={{
                        flex: 1,
                        minHeight: 0,
                        overflowY: 'auto',
                        p: 1.25,
                        '&::-webkit-scrollbar': { width: '8px' },
                        '&::-webkit-scrollbar-track': { bgcolor: 'background.default', borderRadius: '4px' },
                        '&::-webkit-scrollbar-thumb': {
                          bgcolor: 'primary.main',
                          bgcolorOpacity: 0.5,
                          borderRadius: '4px',
                          '&:hover': { bgcolor: 'primary.light' },
                        },
                      }}
                    >
                    {conversationHistory.length === 0 ? (
                      <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', mt: 4 }}>
                        Start a conversation by typing a message below
                      </Typography>
                    ) : (
                      conversationHistory.map((msg, idx) => (
                        <Box key={idx} sx={{ mb: 2 }}>
                          <Paper
                            elevation={0}
                            sx={{
                              p: 1.5,
                              bgcolor: msg.role === 'user' ? 'primary.main' : 'background.default',
                              border: msg.role === 'assistant' ? '1px solid' : 'none',
                              borderColor: msg.role === 'assistant' ? 'primary.main' : 'transparent',
                              borderOpacity: msg.role === 'assistant' ? 0.3 : 1,
                              ml: msg.role === 'user' ? 'auto' : 0,
                              mr: msg.role === 'user' ? 0 : 'auto',
                              maxWidth: '80%',
                              borderRadius: 2,
                            }}
                          >
                            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                              {msg.role === 'user' ? 'You' : 'AI'}
                            </Typography>
                            <Box
                              sx={{
                                '& p': { margin: 0, marginBottom: 1 },
                                '& p:last-child': { marginBottom: 0 },
                                '& ul, & ol': { marginTop: 0.5, marginBottom: 0.5, paddingLeft: 2 },
                                '& code': { 
                                  backgroundColor: msg.role === 'user' ? 'rgba(255, 255, 255, 0.2)' : 'rgba(0, 0, 0, 0.05)',
                                  padding: '2px 4px',
                                  borderRadius: '3px',
                                  fontSize: '0.9em',
                                },
                                '& pre': {
                                  backgroundColor: msg.role === 'user' ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.05)',
                                  padding: 1,
                                  borderRadius: '4px',
                                  overflow: 'auto',
                                  '& code': { backgroundColor: 'transparent', padding: 0 },
                                },
                              }}
                            >
                              <ReactMarkdown>{msg.content}</ReactMarkdown>
                            </Box>
                          </Paper>
                        </Box>
                      ))
                    )}
                    {isWaitingForResponse && (
                      <Box sx={{ textAlign: 'center', py: 2 }}>
                        <Typography variant="body2" color="text.secondary">
                          AI is thinking...
                        </Typography>
                      </Box>
                    )}
                    </Box>
                  </Box>

                  {conversationMeta && (
                    <Box sx={{ mt: 1, mb: 0.75, display: 'flex', gap: 0.75, flexWrap: 'wrap', flexShrink: 0 }}>
                      <Chip size="small" label={`Turn: ${conversationMeta.turn_number}/${conversationMeta.max_turns}`} />
                      {conversationMeta.model_used && (
                        <Chip size="small" label={`Model: ${conversationMeta.model_used}`} />
                      )}
                      {conversationMeta.rag_collection_used && (
                        <Chip size="small" label={`RAG: ${conversationMeta.rag_collection_used}`} />
                      )}
                      {isComplete && (
                        <Chip size="small" label="Complete" color="success" />
                      )}
                    </Box>
                  )}

                  {conversationError && (
                    <Alert severity="error" sx={{ mt: 0.75, mb: 0.75, flexShrink: 0 }}>
                      {conversationError}
                    </Alert>
                  )}

                  {/* Message Input */}
                  <Box sx={{ display: 'flex', gap: 1, mt: 'auto', pt: 1, flexShrink: 0 }}>
                    <TextField
                      fullWidth
                      size="small"
                      label={conversationId ? "Your response" : "Start conversation"}
                      value={currentMessage}
                      onChange={(e) => setCurrentMessage(e.target.value)}
                      onKeyPress={(e) => {
                        if (e.key === 'Enter' && !e.shiftKey && !isWaitingForResponse) {
                          e.preventDefault();
                          handleSendMessage();
                        }
                      }}
                      disabled={isWaitingForResponse || isComplete}
                      multiline
                      rows={2}
                    />
                    <Button
                      variant="contained"
                      size="small"
                      startIcon={<SendIcon />}
                      onClick={handleSendMessage}
                      disabled={!currentMessage.trim() || isWaitingForResponse || isComplete}
                      sx={{ alignSelf: 'flex-end', flexShrink: 0 }}
                    >
                      {isWaitingForResponse ? 'Sending...' : 'Send'}
                    </Button>
                  </Box>

                  {isComplete && (
                    <Alert severity="info" sx={{ mt: 1, flexShrink: 0 }} icon={false}>
                      Conversation complete. Start a new conversation to continue.
                    </Alert>
                  )}
                </>
              ) : (
                <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', mt: 4 }}>
                  Select a dialogue from the list to start a conversation.
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Edit Dialogue Dialog */}
      <Dialog
        open={openEditDialog}
        onClose={() => {
          setOpenEditDialog(false);
          setEditingDialogueId(null);
          setCreateForm({
            name: '',
            description: '',
            system_prompt: '',
            rag_collection: '',
            request_tools: [],
            llm_provider: '',
            model_name: '',
            max_turns: 5,
          });
        }}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Edit Dialogue</DialogTitle>
        <DialogContent>
          {updateMutation.isError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Failed to update dialogue. Please check your inputs and try again.
            </Alert>
          )}
          <Box sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="Name"
              value={createForm.name}
              onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Description"
              value={createForm.description}
              onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
              multiline
              rows={2}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="System Prompt / Instructions"
              value={createForm.system_prompt}
              onChange={(e) => setCreateForm({ ...createForm, system_prompt: e.target.value })}
              multiline
              rows={6}
              sx={{ mb: 2 }}
              placeholder="Describe how the AI should behave for this dialogue..."
            />
            <Autocomplete
              freeSolo
              options={collections.map((col) => col.name)}
              value={createForm.rag_collection}
              onChange={(event, newValue) => {
                setCreateForm({ ...createForm, rag_collection: newValue || '' });
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="RAG Collection (optional)"
                  helperText="Name of an existing RAG collection to use as context"
                  sx={{ mb: 2 }}
                />
              )}
            />
            <Autocomplete
              multiple
              options={requestTools.map((tool) => ({ id: tool.id, label: tool.name }))}
              getOptionLabel={(option) => typeof option === 'string' ? option : option.label}
              value={(createForm.request_tools || []).map(id => {
                const tool = requestTools.find(t => t.id === id);
                return tool ? { id: tool.id, label: tool.name } : null;
              }).filter(Boolean)}
              onChange={(event, newValue) => {
                setCreateForm({ ...createForm, request_tools: newValue.map(v => v.id) });
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Request Tools (optional)"
                  helperText="Select request tools to use in this dialogue"
                  sx={{ mb: 2 }}
                />
              )}
              renderTags={(value, getTagProps) =>
                value.map((option, index) => (
                  <Chip
                    label={option.label}
                    {...getTagProps({ index })}
                    key={option.id}
                    size="small"
                  />
                ))
              }
            />
            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel>LLM Provider (optional)</InputLabel>
              <Select
                value={createForm.llm_provider}
                label="LLM Provider (optional)"
                onChange={(e) => setCreateForm({ ...createForm, llm_provider: e.target.value })}
              >
                <MenuItem value="">
                  <em>Use default provider</em>
                </MenuItem>
                {providersData?.providers?.map((provider) => (
                  <MenuItem key={provider} value={provider}>
                    {provider}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <Autocomplete
              freeSolo
              options={models.map((model) => model.name)}
              value={createForm.model_name}
              onChange={(event, newValue) => {
                setCreateForm({ ...createForm, model_name: newValue || '' });
              }}
              onInputChange={(event, newInputValue) => {
                setCreateForm({ ...createForm, model_name: newInputValue });
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Model Name (optional)"
                  helperText="Choose from available models or enter any model name"
                  sx={{ mb: 2 }}
                />
              )}
            />
            <TextField
              fullWidth
              type="number"
              label="Max Turns"
              value={createForm.max_turns}
              onChange={(e) => setCreateForm({ ...createForm, max_turns: parseInt(e.target.value) || 5 })}
              inputProps={{ min: 1, max: 10 }}
              helperText="Maximum number of conversation turns (1-10)"
              sx={{ mb: 2 }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenEditDialog(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={handleUpdate}
            disabled={updateMutation.isLoading || !createForm.name.trim() || !createForm.system_prompt.trim()}
          >
            {updateMutation.isLoading ? 'Updating...' : 'Update'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Create Dialogue Dialog */}
      <Dialog open={openCreateDialog} onClose={() => setOpenCreateDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>Create Dialogue</DialogTitle>
        <DialogContent>
          {createMutation.isError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Failed to create dialogue. Please check your inputs and try again.
            </Alert>
          )}
          <Box sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="Name"
              value={createForm.name}
              onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Description"
              value={createForm.description}
              onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
              multiline
              rows={2}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="System Prompt / Instructions"
              value={createForm.system_prompt}
              onChange={(e) => setCreateForm({ ...createForm, system_prompt: e.target.value })}
              multiline
              rows={6}
              sx={{ mb: 2 }}
              placeholder="Describe how the AI should behave for this dialogue..."
            />
            <Autocomplete
              freeSolo
              options={collections.map((col) => col.name)}
              value={createForm.rag_collection}
              onChange={(event, newValue) => {
                setCreateForm({ ...createForm, rag_collection: newValue || '' });
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="RAG Collection (optional)"
                  helperText="Name of an existing RAG collection to use as context"
                  sx={{ mb: 2 }}
                />
              )}
            />
            <Autocomplete
              multiple
              options={requestTools.map((tool) => ({ id: tool.id, label: tool.name }))}
              getOptionLabel={(option) => typeof option === 'string' ? option : option.label}
              value={(createForm.request_tools || []).map(id => {
                const tool = requestTools.find(t => t.id === id);
                return tool ? { id: tool.id, label: tool.name } : null;
              }).filter(Boolean)}
              onChange={(event, newValue) => {
                setCreateForm({ ...createForm, request_tools: newValue.map(v => v.id) });
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Request Tools (optional)"
                  helperText="Select request tools to use in this dialogue"
                  sx={{ mb: 2 }}
                />
              )}
              renderTags={(value, getTagProps) =>
                value.map((option, index) => (
                  <Chip
                    label={option.label}
                    {...getTagProps({ index })}
                    key={option.id}
                    size="small"
                  />
                ))
              }
            />
            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel>LLM Provider (optional)</InputLabel>
              <Select
                value={createForm.llm_provider}
                label="LLM Provider (optional)"
                onChange={(e) => setCreateForm({ ...createForm, llm_provider: e.target.value })}
              >
                <MenuItem value="">
                  <em>Use default provider</em>
                </MenuItem>
                {providersData?.providers?.map((provider) => (
                  <MenuItem key={provider} value={provider}>
                    {provider}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <Autocomplete
              freeSolo
              options={models.map((model) => model.name)}
              value={createForm.model_name}
              onChange={(event, newValue) => {
                setCreateForm({ ...createForm, model_name: newValue || '' });
              }}
              onInputChange={(event, newInputValue) => {
                setCreateForm({ ...createForm, model_name: newInputValue });
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Model Name (optional)"
                  helperText="Choose from available models or enter any model name"
                  sx={{ mb: 2 }}
                />
              )}
            />
            <TextField
              fullWidth
              type="number"
              label="Max Turns"
              value={createForm.max_turns}
              onChange={(e) => setCreateForm({ ...createForm, max_turns: parseInt(e.target.value) || 5 })}
              inputProps={{ min: 1, max: 10 }}
              helperText="Maximum number of conversation turns (1-10, default: 5)"
              sx={{ mb: 2 }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenCreateDialog(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={handleCreate}
            disabled={createMutation.isLoading || !createForm.name.trim() || !createForm.system_prompt.trim()}
          >
            {createMutation.isLoading ? 'Creating...' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog
        open={deleteConfirmDialog.open}
        onClose={() => setDeleteConfirmDialog({ open: false, dialogueId: null })}
      >
        <DialogTitle>Confirm Delete</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete this dialogue? This action cannot be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteConfirmDialog({ open: false, dialogueId: null })}>
            Cancel
          </Button>
          <Button onClick={handleDeleteConfirm} color="error" variant="contained">
            Delete
          </Button>
        </DialogActions>
      </Dialog>
        </Box>
      )}

      {tabValue === 1 && (
        <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', width: '100%' }}>
          <Conversation />
        </Box>
      )}
      </Box>
    </ModuleShell>
  );
};

export default Dialogue;

