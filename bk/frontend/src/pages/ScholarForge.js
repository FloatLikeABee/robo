import React, { useState, useMemo, useCallback } from 'react';
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
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  Divider,
  CircularProgress,
  LinearProgress,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  PlayArrow as RunIcon,
  Download as DownloadIcon,
  ExpandMore as ExpandMoreIcon,
  CloudUpload as UploadIcon,
  CheckCircle as CheckIcon,
} from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import api, {
  exportScholarForgePdf,
  getScholarForgeDownloadUrl,
  streamScholarForge,
} from '../services/api';

const WORKFLOW_STEPS = [
  'Setup',
  'Prepare',
  'Clarify',
  'Structure',
  'Generate',
  'Download',
];

const STATUS_STEP = {
  draft: 0,
  preparing: 1,
  clarifying: 2,
  structuring: 3,
  planning: 4,
  generating: 4,
  completed: 5,
  failed: 0,
};

const MIN_PROMPT_WORDS = 100;

const darkScrollSx = {
  overflowY: 'auto',
  overflowX: 'hidden',
  minHeight: 0,
  scrollbarWidth: 'thin',
  scrollbarColor: 'rgba(59, 130, 246, 0.55) rgba(10, 15, 26, 0.9)',
  scrollbarGutter: 'stable',
  '&::-webkit-scrollbar': { width: 8 },
  '&::-webkit-scrollbar-track': {
    bgcolor: 'rgba(10, 15, 26, 0.9)',
    borderRadius: 4,
    margin: '4px 0',
  },
  '&::-webkit-scrollbar-thumb': {
    bgcolor: 'rgba(59, 130, 246, 0.5)',
    borderRadius: 4,
    border: '2px solid rgba(10, 15, 26, 0.9)',
    '&:hover': { bgcolor: 'rgba(56, 189, 248, 0.75)' },
  },
};

/** Direct scroll cap (Articles.js pattern — avoids broken flex height chains). */
const WORKFLOW_SCROLL_MAX = 'calc(100dvh - 480px)';
const PROJECT_LIST_MAX = 'calc(100dvh - 280px)';

const flexFill = {
  flex: '1 1 0',
  minHeight: 0,
};

const shellColumn = {
  height: '100%',
  minHeight: 0,
  display: 'flex',
  flexDirection: 'column',
  overflow: 'hidden',
};

const countWords = (text) => (text || '').trim().split(/\s+/).filter(Boolean).length;

const CITATION_STYLES = ['APA 7', 'IEEE', 'Chicago', 'Harvard', 'Vancouver', 'MLA', 'Other'];

const emptyDocumentMeta = () => ({
  author_name: '',
  author_email: '',
  author_affiliation: '',
  co_authors: [],
  university: '',
  faculty: '',
  department: '',
  degree_program: '',
  degree_awarded: '',
  supervisor: '',
  co_supervisor: '',
  student_id: '',
  submission_date: '',
  location: '',
  language: 'English',
  citation_style: 'APA 7',
  keywords: [],
  abstract_word_limit: '',
  thesis_requirements_notes: '',
});

const emptyForm = () => ({
  subject: '',
  title: '',
  short_intro: '',
  detailed_prompt: '',
  recommended_sites: [],
  document_type: 'article',
  document_meta: emptyDocumentMeta(),
  materials: [],
  siteInput: '',
  coAuthorInput: '',
  keywordInput: '',
});

const isThesisDocumentType = (type) => type === 'thesis' || type === 'dissertation';

const ScholarForge = ({ suppressModuleHeader = false }) => {
  const queryClient = useQueryClient();
  const { data: projects = [], isLoading } = useQuery(
    'scholarForgeProjects',
    api.getScholarForgeProjects,
    { staleTime: 5 * 60 * 1000 },
  );

  const [selected, setSelected] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [form, setForm] = useState(emptyForm());
  const [deleteConfirm, setDeleteConfirm] = useState({ open: false, id: null });

  const [activeStep, setActiveStep] = useState(0);
  const [streamLog, setStreamLog] = useState('');
  const [streamError, setStreamError] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const [clarificationForm, setClarificationForm] = useState(null);
  const [clarifyAnswers, setClarifyAnswers] = useState({});
  const [structure, setStructure] = useState(null);
  const [finalPlan, setFinalPlan] = useState('');
  const [outputPdfId, setOutputPdfId] = useState(null);
  const [outputMdId, setOutputMdId] = useState(null);
  const [confirmNotes, setConfirmNotes] = useState('');
  const [clarifySubmitted, setClarifySubmitted] = useState(false);
  const [flowPipeline, setFlowPipeline] = useState([]);
  const [reviewReports, setReviewReports] = useState([]);
  const [currentSection, setCurrentSection] = useState(null);

  const promptWordCount = useMemo(() => countWords(form.detailed_prompt), [form.detailed_prompt]);
  const promptValid = promptWordCount >= MIN_PROMPT_WORDS;
  const thesisMode = isThesisDocumentType(form.document_type);

  const metaValid = useMemo(() => {
    const m = form.document_meta || {};
    if (!(m.author_name || '').trim()) return false;
    if (thesisMode) {
      return Boolean(
        (m.university || '').trim()
        && (m.department || '').trim()
        && (m.degree_program || '').trim()
        && (m.supervisor || '').trim(),
      );
    }
    return true;
  }, [form.document_meta, thesisMode]);

  const showClarifyForm = useMemo(
    () => Boolean(clarificationForm && !clarificationForm.sufficient && !clarifySubmitted),
    [clarificationForm, clarifySubmitted],
  );

  const resetWorkflow = useCallback(() => {
    setActiveStep(0);
    setStreamLog('');
    setStreamError('');
    setClarificationForm(null);
    setClarifyAnswers({});
    setStructure(null);
    setFinalPlan('');
    setOutputPdfId(null);
    setOutputMdId(null);
    setConfirmNotes('');
    setClarifySubmitted(false);
    setFlowPipeline([]);
    setReviewReports([]);
    setCurrentSection(null);
  }, []);

  const syncFromProject = useCallback((p) => {
    if (!p) return;
    setActiveStep(STATUS_STEP[p.status] ?? 0);
    if (p.clarification) setClarificationForm(p.clarification);
    if (p.clarification_answers) setClarifyAnswers(p.clarification_answers);
    if (p.structure) setStructure(p.structure);
    if (p.final_plan) setFinalPlan(p.final_plan);
    if (p.output_pdf_id) setOutputPdfId(p.output_pdf_id);
    if (p.output_markdown_id) setOutputMdId(p.output_markdown_id);
    const answersSaved = Object.keys(p.clarification_answers || {}).length > 0;
    const pastClarify = p.status && !['draft', 'preparing', 'clarifying'].includes(p.status);
    setClarifySubmitted(pastClarify || answersSaved);
  }, []);

  const createMutation = useMutation(api.createScholarForgeProject, {
    onSuccess: () => {
      queryClient.invalidateQueries('scholarForgeProjects');
      setOpenDialog(false);
      setForm(emptyForm());
    },
  });

  const updateMutation = useMutation(
    ({ id, payload }) => api.updateScholarForgeProject(id, payload),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('scholarForgeProjects');
        setOpenDialog(false);
        setEditingId(null);
        setForm(emptyForm());
      },
    },
  );

  const deleteMutation = useMutation(api.deleteScholarForgeProject, {
    onSuccess: () => {
      queryClient.invalidateQueries('scholarForgeProjects');
      if (selected?.id === deleteConfirm.id) setSelected(null);
    },
  });

  const buildPayload = () => {
    const m = form.document_meta || emptyDocumentMeta();
    const abstractLimit = m.abstract_word_limit === '' || m.abstract_word_limit == null
      ? null
      : Number(m.abstract_word_limit);
    return {
      subject: form.subject.trim(),
      title: form.title.trim(),
      short_intro: form.short_intro.trim(),
      detailed_prompt: form.detailed_prompt.trim(),
      recommended_sites: form.recommended_sites,
      document_type: form.document_type,
      document_meta: {
        author_name: (m.author_name || '').trim(),
        author_email: (m.author_email || '').trim() || null,
        author_affiliation: (m.author_affiliation || '').trim(),
        co_authors: m.co_authors || [],
        university: (m.university || '').trim(),
        faculty: (m.faculty || '').trim(),
        department: (m.department || '').trim(),
        degree_program: (m.degree_program || '').trim(),
        degree_awarded: (m.degree_awarded || '').trim(),
        supervisor: (m.supervisor || '').trim(),
        co_supervisor: (m.co_supervisor || '').trim() || null,
        student_id: (m.student_id || '').trim() || null,
        submission_date: (m.submission_date || '').trim(),
        location: (m.location || '').trim(),
        language: (m.language || 'English').trim() || 'English',
        citation_style: (m.citation_style || 'APA 7').trim() || 'APA 7',
        keywords: m.keywords || [],
        abstract_word_limit: Number.isFinite(abstractLimit) && abstractLimit > 0 ? abstractLimit : null,
        thesis_requirements_notes: (m.thesis_requirements_notes || '').trim(),
      },
      materials: form.materials.map(({ filename, format, content, description }) => ({
        filename,
        format,
        content,
        description: description || null,
      })),
    };
  };

  const setDocumentMeta = (key, value) => {
    setForm((f) => ({
      ...f,
      document_meta: { ...f.document_meta, [key]: value },
    }));
  };

  const handleSave = () => {
    const payload = buildPayload();
    if (editingId) updateMutation.mutate({ id: editingId, payload });
    else createMutation.mutate(payload);
  };

  const openCreate = () => {
    setEditingId(null);
    setForm(emptyForm());
    setOpenDialog(true);
  };

  const openEdit = (p) => {
    setEditingId(p.id);
    const meta = { ...emptyDocumentMeta(), ...(p.document_meta || {}) };
    setForm({
      subject: p.subject || '',
      title: p.title || '',
      short_intro: p.short_intro || '',
      detailed_prompt: p.detailed_prompt || '',
      recommended_sites: p.recommended_sites || [],
      document_type: p.document_type || 'article',
      document_meta: {
        ...meta,
        abstract_word_limit: meta.abstract_word_limit ?? '',
      },
      materials: p.materials || [],
      siteInput: '',
      coAuthorInput: '',
      keywordInput: '',
    });
    setOpenDialog(true);
  };

  const handleMaterialFiles = (e) => {
    const files = Array.from(e.target.files || []);
    files.forEach((file) => {
      const ext = (file.name.split('.').pop() || '').toLowerCase();
      const reader = new FileReader();
      reader.onload = () => {
        let format = 'txt';
        let content = reader.result;
        if (['json', 'csv', 'txt', 'md'].includes(ext)) {
          format = ext === 'md' ? 'txt' : ext;
        } else if (ext === 'pdf') {
          format = 'pdf';
          const dataUrl = reader.result;
          const b64 = typeof dataUrl === 'string' ? dataUrl.split(',')[1] : '';
          content = `base64:${b64}`;
        }
        setForm((prev) => ({
          ...prev,
          materials: [
            ...prev.materials,
            { filename: file.name, format, content, description: '' },
          ],
        }));
      };
      if (ext === 'pdf') reader.readAsDataURL(file);
      else reader.readAsText(file);
    });
    e.target.value = '';
  };

  const handleImageUpload = async (e) => {
    if (!selected) return;
    const files = Array.from(e.target.files || []);
    if (!files.length) return;
    const fd = new FormData();
    files.forEach((f) => fd.append('files', f));
    try {
      await api.uploadScholarForgeImages(selected.id, fd);
      queryClient.invalidateQueries('scholarForgeProjects');
      const refreshed = await api.getScholarForgeProject(selected.id);
      setSelected(refreshed);
    } catch (err) {
      setStreamError(err.message || 'Image upload failed');
    }
    e.target.value = '';
  };

  const appendLog = (msg) => setStreamLog((prev) => `${prev}${msg}\n`);

  const runStream = async (action, onEvent, body = {}) => {
    if (!selected) return;
    setIsStreaming(true);
    setStreamError('');
    try {
      await streamScholarForge(selected.id, action, (evt) => {
        if (evt.type === 'log') appendLog(evt.message);
        if (evt.type === 'heartbeat') appendLog(`… ${evt.message}`);
        if (evt.type === 'delta' && evt.text) appendLog(evt.text);
        if (evt.type === 'error') setStreamError(evt.message);
        if (onEvent) onEvent(evt);
      }, body);
      const refreshed = await api.getScholarForgeProject(selected.id);
      setSelected(refreshed);
      syncFromProject(refreshed);
      queryClient.invalidateQueries('scholarForgeProjects');
    } catch (err) {
      setStreamError(err.message || 'Stream failed');
    } finally {
      setIsStreaming(false);
    }
  };

  const handlePrepare = () => runStream('prepare', (evt) => {
    if (evt.type === 'session_ready') setActiveStep(2);
  });

  const handleClarify = () => runStream('clarify', (evt) => {
    if (evt.type === 'clarification') {
      setClarificationForm(evt.form);
      setClarifySubmitted(false);
      if (evt.sufficient) {
        setClarifySubmitted(true);
        setActiveStep(3);
      }
    }
  });

  const submitClarifyAnswers = async () => {
    if (!selected) return false;
    try {
      await api.submitScholarForgeClarification(selected.id, { answers: clarifyAnswers });
      setClarifySubmitted(true);
      setActiveStep(3);
      appendLog('Clarification answers saved.');
      const refreshed = await api.getScholarForgeProject(selected.id);
      setSelected(refreshed);
      queryClient.invalidateQueries('scholarForgeProjects');
      return true;
    } catch (err) {
      setStreamError(err.message || 'Failed to save clarification answers');
      return false;
    }
  };

  const handleSubmitClarify = () => submitClarifyAnswers();

  const handleStructure = async () => {
    if (showClarifyForm) {
      const ok = await submitClarifyAnswers();
      if (!ok) return;
    }
    runStream('structure', (evt) => {
      if (evt.type === 'structure') {
        setStructure(evt.structure);
        setActiveStep(3);
        appendLog('Structure plan ready — review and edit sections below.');
      }
    });
  };

  const handleSaveStructure = async () => {
    if (!selected || !structure) return;
    try {
      await api.updateScholarForgeStructure(selected.id, { structure });
      appendLog('Structure saved.');
    } catch (err) {
      setStreamError(err.message);
    }
  };

  const handleConfirmStructure = () => runStream(
    'confirm-structure',
    (evt) => {
      if (evt.type === 'plan') {
        setFinalPlan(evt.final_plan);
        setActiveStep(4);
      }
    },
    { confirmed: true, notes: confirmNotes },
  );

  const handleGenerate = () => {
    setActiveStep(4);
    setFlowPipeline([]);
    setReviewReports([]);
    setCurrentSection(null);
    runStream('generate', (evt) => {
      if (evt.type === 'section_start') {
        setCurrentSection({ id: evt.section_id, title: evt.title, index: evt.index });
        appendLog(`\n--- ${evt.title} ---\n`);
      }
      if (evt.type === 'flow_step' && evt.step) {
        setFlowPipeline((prev) => {
          const idx = prev.findIndex((s) => s.step_id === evt.step.step_id);
          if (idx >= 0) {
            const next = [...prev];
            next[idx] = evt.step;
            return next;
          }
          return [...prev, evt.step];
        });
      }
      if (evt.type === 'review_report' && evt.report) {
        setReviewReports((prev) => [
          ...prev,
          {
            sectionId: evt.section_id,
            paragraphIndex: evt.paragraph_index,
            round: evt.round,
            report: evt.report,
          },
        ]);
        const r = evt.report;
        appendLog(
          `\n[Review §${evt.paragraph_index + 1} r${evt.round}] `
          + `score=${r.quality_score} approved=${r.approved}\n${r.summary || ''}\n`,
        );
      }
      if (evt.type === 'paragraph_delta' && evt.text) appendLog(evt.text);
      if (evt.type === 'revision_delta' && evt.text) appendLog(evt.text);
      if (evt.type === 'delta' && evt.text) appendLog(evt.text);
      if (evt.type === 'done') {
        setOutputPdfId(evt.pdf_id || null);
        setOutputMdId(evt.markdown_id);
        setActiveStep(5);
      }
    });
  };

  const handleExportPdf = async () => {
    if (!selected) return;
    setStreamError(null);
    try {
      const result = await exportScholarForgePdf(selected.id);
      setOutputPdfId(result.pdf_id);
      appendLog(`PDF exported: ${result.pdf_filename}`);
    } catch (err) {
      setStreamError(err.response?.data?.detail || err.message);
    }
  };

  const renderClarificationFields = () => {
    if (!clarificationForm?.fields?.length) return null;
    return (
      <Box sx={{ mt: 0.5, pb: 1 }}>
        {clarificationForm.title && (
          <Typography sx={{ fontSize: '0.8rem', fontWeight: 600, mb: 0.5 }}>
            {clarificationForm.title}
          </Typography>
        )}
        <Typography variant="body2" sx={{ mb: 1, fontSize: '0.75rem' }}>{clarificationForm.intro}</Typography>
        {clarificationForm.fields.map((field) => (
          <Box key={field.id} sx={{ mb: 1 }}>
            {field.field_type === 'textarea' ? (
              <TextField
                fullWidth
                size="small"
                multiline
                minRows={2}
                label={field.label}
                value={clarifyAnswers[field.id] || ''}
                onChange={(e) => setClarifyAnswers((a) => ({ ...a, [field.id]: e.target.value }))}
                helperText={field.help_text}
              />
            ) : field.field_type === 'select' ? (
              <FormControl fullWidth size="small">
                <InputLabel>{field.label}</InputLabel>
                <Select
                  value={clarifyAnswers[field.id] || ''}
                  label={field.label}
                  onChange={(e) => setClarifyAnswers((a) => ({ ...a, [field.id]: e.target.value }))}
                >
                  {(field.options || []).map((o) => (
                    <MenuItem key={o} value={o}>{o}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            ) : (
              <TextField
                fullWidth
                size="small"
                label={field.label}
                value={clarifyAnswers[field.id] || ''}
                onChange={(e) => setClarifyAnswers((a) => ({ ...a, [field.id]: e.target.value }))}
                helperText={field.help_text}
              />
            )}
          </Box>
        ))}
      </Box>
    );
  };

  const renderStructureScrollContent = () => {
    if (!structure) return null;
    return (
      <>
        <TextField
          fullWidth
          size="small"
          label="Document title"
          value={structure.document_title || ''}
          onChange={(e) => setStructure((s) => ({ ...s, document_title: e.target.value }))}
          sx={{ mb: 1 }}
        />
        <TextField
          fullWidth
          size="small"
          multiline
          minRows={2}
          label="Abstract outline"
          value={structure.abstract_outline || ''}
          onChange={(e) => setStructure((s) => ({ ...s, abstract_outline: e.target.value }))}
          sx={{ mb: 1 }}
        />
        {(structure.sections || []).map((sec, idx) => (
          <Accordion key={sec.id || idx} disableGutters sx={{ mb: 0.5 }}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Typography sx={{ fontSize: '0.75rem' }}>
                {sec.order}. {sec.title}
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <TextField
                fullWidth
                size="small"
                label="Title"
                value={sec.title}
                onChange={(e) => {
                  const sections = [...structure.sections];
                  sections[idx] = { ...sec, title: e.target.value };
                  setStructure((s) => ({ ...s, sections }));
                }}
                sx={{ mb: 1 }}
              />
              <TextField
                fullWidth
                size="small"
                multiline
                minRows={2}
                label="Description"
                value={sec.description || ''}
                onChange={(e) => {
                  const sections = [...structure.sections];
                  sections[idx] = { ...sec, description: e.target.value };
                  setStructure((s) => ({ ...s, sections }));
                }}
              />
            </AccordionDetails>
          </Accordion>
        ))}
        {finalPlan && (
          <TextField
            fullWidth
            multiline
            minRows={4}
            label="Writing plan"
            value={finalPlan}
            InputProps={{ readOnly: true }}
            sx={{ mt: 1, mb: 1 }}
          />
        )}
      </>
    );
  };

  const renderStructureFooter = () => {
    if (!structure) return null;
    return (
      <Box
        sx={{
          flexShrink: 0,
          px: 1,
          py: 1,
          borderTop: '1px solid',
          borderColor: 'divider',
          bgcolor: 'rgba(15, 23, 42, 0.92)',
          display: 'flex',
          gap: 1,
          flexWrap: 'wrap',
          alignItems: 'center',
        }}
      >
        <Button size="small" variant="outlined" onClick={handleSaveStructure}>
          Save edits
        </Button>
        <TextField
          size="small"
          placeholder="Confirmation notes (optional)"
          value={confirmNotes}
          onChange={(e) => setConfirmNotes(e.target.value)}
          sx={{ flex: 1, minWidth: 160 }}
        />
        <Button size="small" variant="contained" onClick={handleConfirmStructure} disabled={isStreaming}>
          Confirm & build plan
        </Button>
      </Box>
    );
  };

  return (
    <Box sx={{ ...shellColumn, flex: 1, minHeight: 0, height: '100%', p: suppressModuleHeader ? 0 : 1 }}>
      {!suppressModuleHeader && (
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1, flexShrink: 0 }}>
          <Typography variant="h4">ScholarForge</Typography>
          <Button size="small" startIcon={<AddIcon />} variant="contained" onClick={openCreate}>
            New project
          </Button>
        </Box>
      )}
      {suppressModuleHeader && (
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 0.5, flexShrink: 0 }}>
          <Button size="small" startIcon={<AddIcon />} variant="contained" onClick={openCreate}>
            New project
          </Button>
        </Box>
      )}

      <Box
        sx={{
          ...flexFill,
          display: 'grid',
          gridTemplateRows: 'minmax(0, 1fr)',
          gridTemplateColumns: { xs: '1fr', md: 'minmax(0, 1fr) minmax(0, 2fr)' },
          gap: 1,
          overflow: 'hidden',
          '& > *': { minHeight: 0, overflow: 'hidden' },
        }}
      >
        <Box
          sx={{
            ...darkScrollSx,
            maxHeight: PROJECT_LIST_MAX,
            pr: 0.5,
          }}
        >
          {isLoading && <CircularProgress size={20} />}
          {projects.map((p) => (
            <Card
              key={p.id}
              sx={{
                mb: 0.5,
                cursor: 'pointer',
                border: selected?.id === p.id ? '1px solid' : undefined,
                borderColor: selected?.id === p.id ? 'primary.main' : undefined,
              }}
              onClick={() => {
                setSelected(p);
                resetWorkflow();
                syncFromProject(p);
              }}
            >
              <CardContent sx={{ py: 1, '&:last-child': { pb: 1 } }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography sx={{ fontSize: '0.8rem', fontWeight: 600 }}>{p.title}</Typography>
                  <Box>
                    <IconButton size="small" onClick={(e) => { e.stopPropagation(); openEdit(p); }}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={(e) => { e.stopPropagation(); setDeleteConfirm({ open: true, id: p.id }); }}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Box>
                </Box>
                <Typography sx={{ fontSize: '0.65rem', color: 'text.secondary' }}>{p.subject}</Typography>
                <Chip label={p.document_type} size="small" sx={{ mt: 0.5, height: 20, fontSize: '0.6rem' }} />
                <Chip label={p.status} size="small" sx={{ mt: 0.5, ml: 0.5, height: 20, fontSize: '0.6rem' }} />
              </CardContent>
            </Card>
          ))}
        </Box>

        <Box sx={shellColumn}>
          {!selected ? (
            <Alert severity="info">Select or create a ScholarForge project to begin.</Alert>
          ) : (
            <Box sx={{ ...shellColumn, ...flexFill }}>
              <Stepper activeStep={activeStep} alternativeLabel sx={{ mb: 1, flexShrink: 0 }}>
                {WORKFLOW_STEPS.map((label) => (
                  <Step key={label}><StepLabel sx={{ '& .MuiStepLabel-label': { fontSize: '0.6rem' } }}>{label}</StepLabel></Step>
                ))}
              </Stepper>
              {isStreaming && <LinearProgress sx={{ mb: 1, flexShrink: 0 }} />}

              <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap', mb: 1, flexShrink: 0 }}>
                <Button size="small" variant="outlined" startIcon={<RunIcon />} disabled={isStreaming} onClick={handlePrepare}>
                  1. Prepare
                </Button>
                <Button size="small" variant="outlined" disabled={isStreaming} onClick={handleClarify}>
                  2. Clarify
                </Button>
                <Button size="small" variant="outlined" disabled={isStreaming} onClick={handleStructure}>
                  3. Structure
                </Button>
                {showClarifyForm && (
                  <Chip
                    label="Submit answers below first — or Structure saves them for you"
                    size="small"
                    color="info"
                    sx={{ height: 24, fontSize: '0.6rem', maxWidth: '100%' }}
                  />
                )}
                <Button size="small" variant="contained" startIcon={<RunIcon />} disabled={isStreaming || !finalPlan} onClick={handleGenerate}>
                  4. Generate
                </Button>
                <Button component="label" size="small" variant="outlined" startIcon={<UploadIcon />}>
                  Figures
                  <input type="file" accept="image/*" multiple hidden onChange={handleImageUpload} />
                </Button>
                {outputPdfId && (
                  <Button
                    size="small"
                    component="a"
                    href={getScholarForgeDownloadUrl(outputPdfId, 'pdf')}
                    download
                    startIcon={<DownloadIcon />}
                  >
                    PDF
                  </Button>
                )}
                {outputMdId && (
                  <Button
                    size="small"
                    component="a"
                    href={getScholarForgeDownloadUrl(outputMdId, 'md')}
                    download
                    startIcon={<DownloadIcon />}
                  >
                    Markdown
                  </Button>
                )}
                {outputMdId && !outputPdfId && (
                  <Button size="small" variant="outlined" startIcon={<DownloadIcon />} onClick={handleExportPdf}>
                    Export PDF
                  </Button>
                )}
              </Box>

              <Divider sx={{ mb: 1, flexShrink: 0 }} />
              <Typography sx={{ fontSize: '0.7rem', color: 'text.secondary', mb: 0.5, flexShrink: 0 }}>
                {selected.title} · {selected.document_type} · RAG: {selected.rag_collection || 'pending'}
                {selected.document_meta?.author_name ? ` · ${selected.document_meta.author_name}` : ''}
              </Typography>
              {selected.document_meta?.supervisor && (
                <Typography sx={{ fontSize: '0.65rem', color: 'text.secondary', mb: 0.5, flexShrink: 0 }}>
                  Supervisor: {selected.document_meta.supervisor}
                  {selected.document_meta.university ? ` · ${selected.document_meta.university}` : ''}
                </Typography>
              )}

              <Box
                sx={{
                  display: 'flex',
                  flexDirection: 'column',
                  border: '1px solid',
                  borderColor: 'divider',
                  borderRadius: 1,
                  bgcolor: 'rgba(10, 15, 26, 0.4)',
                  overflow: 'hidden',
                }}
              >
                {showClarifyForm && (
                  <Typography
                    sx={{
                      flexShrink: 0,
                      px: 1,
                      py: 0.75,
                      fontSize: '0.65rem',
                      color: 'primary.light',
                      borderBottom: '1px solid',
                      borderColor: 'divider',
                    }}
                  >
                    {clarificationForm.fields?.length || 0} questions — scroll this panel for all fields
                  </Typography>
                )}
                {structure && !showClarifyForm && (
                  <Typography
                    sx={{
                      flexShrink: 0,
                      px: 1,
                      py: 0.75,
                      fontSize: '0.65rem',
                      color: 'primary.light',
                      borderBottom: '1px solid',
                      borderColor: 'divider',
                    }}
                  >
                    {(structure.sections || []).length} sections — scroll to review all chapters
                  </Typography>
                )}
                <Box
                  sx={{
                    maxHeight: WORKFLOW_SCROLL_MAX,
                    px: 1,
                    py: 1,
                    ...darkScrollSx,
                  }}
                >
                  {showClarifyForm && renderClarificationFields()}
                  {clarificationForm?.sufficient && !structure && (
                    <Alert severity="success" icon={<CheckIcon />} sx={{ mb: 1, py: 0 }}>
                      Requirements sufficient — proceed to structure.
                    </Alert>
                  )}
                  {renderStructureScrollContent()}
                </Box>
                {showClarifyForm && (
                  <Box
                    sx={{
                      flexShrink: 0,
                      px: 1,
                      py: 1,
                      borderTop: '1px solid',
                      borderColor: 'divider',
                      bgcolor: 'rgba(15, 23, 42, 0.92)',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 1,
                      flexWrap: 'wrap',
                    }}
                  >
                    <Typography sx={{ fontSize: '0.65rem', color: 'text.secondary', flex: 1, minWidth: 120 }}>
                      Done filling in? Save answers, then run Structure.
                    </Typography>
                    <Button size="small" variant="contained" onClick={handleSubmitClarify}>
                      Submit answers
                    </Button>
                    <Button size="small" variant="outlined" disabled={isStreaming} onClick={handleStructure}>
                      Submit & run Structure
                    </Button>
                  </Box>
                )}
                {structure && !showClarifyForm && renderStructureFooter()}
              </Box>

              <Typography sx={{ fontSize: '0.6rem', color: 'text.secondary', mt: 0.75, mb: 0.25, flexShrink: 0 }}>
                Agent flow (write → review → revise per paragraph)
              </Typography>
              {currentSection && (
                <Typography sx={{ fontSize: '0.65rem', color: 'primary.main', mb: 0.5, flexShrink: 0 }}>
                  Section: {currentSection.title}
                </Typography>
              )}
              {flowPipeline.length > 0 && (
                <Box sx={{ flexShrink: 0, mb: 1, display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                  {flowPipeline.slice(-12).map((step) => (
                    <Chip
                      key={step.step_id}
                      size="small"
                      label={step.label}
                      color={
                        step.status === 'done' ? 'success'
                          : step.status === 'running' ? 'primary'
                            : step.status === 'failed' ? 'error' : 'default'
                      }
                      variant={step.status === 'running' ? 'filled' : 'outlined'}
                      sx={{ fontSize: '0.6rem', height: 22 }}
                    />
                  ))}
                </Box>
              )}
              {reviewReports.length > 0 && (
                <Accordion disableGutters sx={{ flexShrink: 0, mb: 1, '&:before': { display: 'none' } }}>
                  <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Typography sx={{ fontSize: '0.7rem' }}>
                      Review reports ({reviewReports.length})
                    </Typography>
                  </AccordionSummary>
                  <AccordionDetails sx={{ pt: 0, maxHeight: 160, ...darkScrollSx }}>
                    {reviewReports.slice(-8).reverse().map((item, i) => (
                      <Box key={`${item.sectionId}-${item.paragraphIndex}-${item.round}-${i}`} sx={{ mb: 1 }}>
                        <Typography sx={{ fontSize: '0.65rem', fontWeight: 600 }}>
                          Paragraph {item.paragraphIndex + 1} · round {item.round}
                          {' · '}
                          score {item.report.quality_score}
                          {item.report.approved ? ' · approved' : ' · needs revision'}
                        </Typography>
                        {item.report.summary && (
                          <Typography sx={{ fontSize: '0.62rem', color: 'text.secondary' }}>
                            {item.report.summary}
                          </Typography>
                        )}
                        {(item.report.suggestions || []).slice(0, 3).map((s, j) => (
                          <Typography key={j} sx={{ fontSize: '0.6rem', pl: 1 }}>
                            • {s}
                          </Typography>
                        ))}
                      </Box>
                    ))}
                  </AccordionDetails>
                </Accordion>
              )}

              <Typography sx={{ fontSize: '0.6rem', color: 'text.secondary', mt: 0.75, mb: 0.25, flexShrink: 0 }}>
                Pipeline log
              </Typography>
              <Box
                sx={{
                  flexShrink: 0,
                  height: { xs: 88, md: 100 },
                  ...darkScrollSx,
                  bgcolor: 'action.hover',
                  borderRadius: 1,
                  border: '1px solid',
                  borderColor: 'divider',
                  p: 1,
                  fontFamily: 'monospace',
                  fontSize: '0.72rem',
                  whiteSpace: 'pre-wrap',
                }}
              >
                {streamLog || '(Streaming output appears here.)'}
              </Box>
              {streamError && <Alert severity="error" sx={{ mt: 1, flexShrink: 0 }}>{streamError}</Alert>}
            </Box>
          )}
        </Box>
      </Box>

      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>{editingId ? 'Edit project' : 'New ScholarForge project'}</DialogTitle>
        <DialogContent>
          {(createMutation.error || updateMutation.error) && (
            <Alert severity="error" sx={{ mb: 1 }}>
              {(createMutation.error || updateMutation.error)?.response?.data?.detail
                || (createMutation.error || updateMutation.error)?.message}
            </Alert>
          )}
          <Grid container spacing={1} sx={{ mt: 0.5 }}>
            <Grid item xs={12} sm={8}>
              <TextField fullWidth size="small" label="Title" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} />
            </Grid>
            <Grid item xs={12} sm={4}>
              <FormControl fullWidth size="small">
                <InputLabel>Type</InputLabel>
                <Select value={form.document_type} label="Type" onChange={(e) => setForm({ ...form, document_type: e.target.value })}>
                  <MenuItem value="article">Article</MenuItem>
                  <MenuItem value="thesis">Thesis</MenuItem>
                  <MenuItem value="dissertation">Dissertation</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12}>
              <TextField fullWidth size="small" label="Subject / field" value={form.subject} onChange={(e) => setForm({ ...form, subject: e.target.value })} />
            </Grid>

            <Grid item xs={12}>
              <Accordion defaultExpanded disableGutters sx={{ bgcolor: 'transparent', '&:before': { display: 'none' } }}>
                <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                  <Typography sx={{ fontSize: '0.8rem', fontWeight: 600 }}>
                    Author &amp; document metadata
                    {thesisMode ? ' (required for thesis)' : ''}
                  </Typography>
                </AccordionSummary>
                <AccordionDetails sx={{ pt: 0 }}>
                  <Grid container spacing={1}>
                    <Grid item xs={12} sm={6}>
                      <TextField
                        fullWidth
                        size="small"
                        required
                        label="Author name"
                        value={form.document_meta.author_name}
                        onChange={(e) => setDocumentMeta('author_name', e.target.value)}
                      />
                    </Grid>
                    <Grid item xs={12} sm={6}>
                      <TextField
                        fullWidth
                        size="small"
                        label="Author email"
                        value={form.document_meta.author_email}
                        onChange={(e) => setDocumentMeta('author_email', e.target.value)}
                      />
                    </Grid>
                    <Grid item xs={12} sm={6}>
                      <TextField
                        fullWidth
                        size="small"
                        label="Affiliation"
                        helperText={thesisMode ? 'Optional if university fields below are filled' : 'Department, lab, or institution'}
                        value={form.document_meta.author_affiliation}
                        onChange={(e) => setDocumentMeta('author_affiliation', e.target.value)}
                      />
                    </Grid>
                    <Grid item xs={12} sm={3}>
                      <TextField
                        fullWidth
                        size="small"
                        label="Language"
                        value={form.document_meta.language}
                        onChange={(e) => setDocumentMeta('language', e.target.value)}
                      />
                    </Grid>
                    <Grid item xs={12} sm={3}>
                      <FormControl fullWidth size="small">
                        <InputLabel>Citation style</InputLabel>
                        <Select
                          value={form.document_meta.citation_style}
                          label="Citation style"
                          onChange={(e) => setDocumentMeta('citation_style', e.target.value)}
                        >
                          {CITATION_STYLES.map((s) => (
                            <MenuItem key={s} value={s}>{s}</MenuItem>
                          ))}
                        </Select>
                      </FormControl>
                    </Grid>

                    {thesisMode && (
                      <>
                        <Grid item xs={12} sm={6}>
                          <TextField
                            fullWidth
                            size="small"
                            required
                            label="University / institution"
                            value={form.document_meta.university}
                            onChange={(e) => setDocumentMeta('university', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <TextField
                            fullWidth
                            size="small"
                            label="Faculty / school"
                            value={form.document_meta.faculty}
                            onChange={(e) => setDocumentMeta('faculty', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <TextField
                            fullWidth
                            size="small"
                            required
                            label="Department"
                            value={form.document_meta.department}
                            onChange={(e) => setDocumentMeta('department', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <TextField
                            fullWidth
                            size="small"
                            required
                            label="Degree program"
                            placeholder="e.g. MSc Chemical Engineering"
                            value={form.document_meta.degree_program}
                            onChange={(e) => setDocumentMeta('degree_program', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <TextField
                            fullWidth
                            size="small"
                            label="Degree awarded"
                            placeholder="e.g. Master of Science"
                            value={form.document_meta.degree_awarded}
                            onChange={(e) => setDocumentMeta('degree_awarded', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <TextField
                            fullWidth
                            size="small"
                            label="Student ID"
                            value={form.document_meta.student_id}
                            onChange={(e) => setDocumentMeta('student_id', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <TextField
                            fullWidth
                            size="small"
                            required
                            label="Supervisor / advisor"
                            value={form.document_meta.supervisor}
                            onChange={(e) => setDocumentMeta('supervisor', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <TextField
                            fullWidth
                            size="small"
                            label="Co-supervisor"
                            value={form.document_meta.co_supervisor}
                            onChange={(e) => setDocumentMeta('co_supervisor', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={4}>
                          <TextField
                            fullWidth
                            size="small"
                            label="Submission date"
                            placeholder="YYYY-MM-DD or Month Year"
                            value={form.document_meta.submission_date}
                            onChange={(e) => setDocumentMeta('submission_date', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={4}>
                          <TextField
                            fullWidth
                            size="small"
                            label="Location"
                            placeholder="City, Country"
                            value={form.document_meta.location}
                            onChange={(e) => setDocumentMeta('location', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12} sm={4}>
                          <TextField
                            fullWidth
                            size="small"
                            type="number"
                            label="Abstract word limit"
                            value={form.document_meta.abstract_word_limit}
                            onChange={(e) => setDocumentMeta('abstract_word_limit', e.target.value)}
                          />
                        </Grid>
                        <Grid item xs={12}>
                          <TextField
                            fullWidth
                            size="small"
                            multiline
                            minRows={3}
                            label="Institutional / thesis formatting requirements"
                            placeholder="Paste your university thesis guidelines: margins, front matter, declaration text, numbering, etc."
                            value={form.document_meta.thesis_requirements_notes}
                            onChange={(e) => setDocumentMeta('thesis_requirements_notes', e.target.value)}
                          />
                        </Grid>
                      </>
                    )}

                    <Grid item xs={12}>
                      <Box sx={{ display: 'flex', gap: 0.5, mb: 0.5 }}>
                        <TextField
                          fullWidth
                          size="small"
                          label={thesisMode ? 'Co-author (optional)' : 'Co-author'}
                          value={form.coAuthorInput}
                          onChange={(e) => setForm({ ...form, coAuthorInput: e.target.value })}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter' && form.coAuthorInput.trim()) {
                              setDocumentMeta('co_authors', [...form.document_meta.co_authors, form.coAuthorInput.trim()]);
                              setForm((f) => ({ ...f, coAuthorInput: '' }));
                            }
                          }}
                        />
                        <Button
                          size="small"
                          onClick={() => {
                            if (form.coAuthorInput.trim()) {
                              setDocumentMeta('co_authors', [...form.document_meta.co_authors, form.coAuthorInput.trim()]);
                              setForm((f) => ({ ...f, coAuthorInput: '' }));
                            }
                          }}
                        >
                          Add
                        </Button>
                      </Box>
                      {form.document_meta.co_authors.map((name, i) => (
                        <Chip
                          key={`${name}-${i}`}
                          label={name}
                          size="small"
                          onDelete={() => setDocumentMeta(
                            'co_authors',
                            form.document_meta.co_authors.filter((_, j) => j !== i),
                          )}
                          sx={{ mr: 0.5, mb: 0.5 }}
                        />
                      ))}
                    </Grid>
                    <Grid item xs={12}>
                      <Box sx={{ display: 'flex', gap: 0.5, mb: 0.5 }}>
                        <TextField
                          fullWidth
                          size="small"
                          label="Keyword"
                          value={form.keywordInput}
                          onChange={(e) => setForm({ ...form, keywordInput: e.target.value })}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter' && form.keywordInput.trim()) {
                              setDocumentMeta('keywords', [...form.document_meta.keywords, form.keywordInput.trim()]);
                              setForm((f) => ({ ...f, keywordInput: '' }));
                            }
                          }}
                        />
                        <Button
                          size="small"
                          onClick={() => {
                            if (form.keywordInput.trim()) {
                              setDocumentMeta('keywords', [...form.document_meta.keywords, form.keywordInput.trim()]);
                              setForm((f) => ({ ...f, keywordInput: '' }));
                            }
                          }}
                        >
                          Add
                        </Button>
                      </Box>
                      {form.document_meta.keywords.map((kw, i) => (
                        <Chip
                          key={`${kw}-${i}`}
                          label={kw}
                          size="small"
                          onDelete={() => setDocumentMeta(
                            'keywords',
                            form.document_meta.keywords.filter((_, j) => j !== i),
                          )}
                          sx={{ mr: 0.5, mb: 0.5 }}
                        />
                      ))}
                    </Grid>
                  </Grid>
                </AccordionDetails>
              </Accordion>
            </Grid>

            <Grid item xs={12}>
              <TextField fullWidth size="small" multiline minRows={2} label="Short intro (guides the AIs)" value={form.short_intro} onChange={(e) => setForm({ ...form, short_intro: e.target.value })} />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                size="small"
                multiline
                minRows={5}
                label={`Detailed requirements (min ${MIN_PROMPT_WORDS} words)`}
                value={form.detailed_prompt}
                onChange={(e) => setForm({ ...form, detailed_prompt: e.target.value })}
                error={form.detailed_prompt.length > 0 && !promptValid}
                helperText={`${promptWordCount} / ${MIN_PROMPT_WORDS} words`}
              />
            </Grid>
            <Grid item xs={12}>
              <Box sx={{ display: 'flex', gap: 0.5 }}>
                <TextField
                  fullWidth
                  size="small"
                  label="Recommended reference site"
                  value={form.siteInput}
                  onChange={(e) => setForm({ ...form, siteInput: e.target.value })}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && form.siteInput.trim()) {
                      setForm((f) => ({
                        ...f,
                        recommended_sites: [...f.recommended_sites, f.siteInput.trim()],
                        siteInput: '',
                      }));
                    }
                  }}
                />
                <Button
                  size="small"
                  onClick={() => {
                    if (form.siteInput.trim()) {
                      setForm((f) => ({
                        ...f,
                        recommended_sites: [...f.recommended_sites, f.siteInput.trim()],
                        siteInput: '',
                      }));
                    }
                  }}
                >
                  Add
                </Button>
              </Box>
              <Box sx={{ mt: 0.5 }}>
                {form.recommended_sites.map((s, i) => (
                  <Chip key={s} label={s} size="small" onDelete={() => setForm((f) => ({ ...f, recommended_sites: f.recommended_sites.filter((_, j) => j !== i) }))} sx={{ mr: 0.5, mb: 0.5 }} />
                ))}
              </Box>
            </Grid>
            <Grid item xs={12}>
              <Button variant="outlined" component="label" size="small" startIcon={<UploadIcon />}>
                Upload materials (txt, md, json, csv, pdf)
                <input type="file" multiple accept=".txt,.md,.json,.csv,.pdf,text/*" hidden onChange={handleMaterialFiles} />
              </Button>
              {form.materials.map((m, i) => (
                <Chip key={`${m.filename}-${i}`} label={m.filename} size="small" onDelete={() => setForm((f) => ({ ...f, materials: f.materials.filter((_, j) => j !== i) }))} sx={{ ml: 0.5 }} />
              ))}
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button
            variant="contained"
            disabled={!form.title || !form.subject || !promptValid || !metaValid || createMutation.isLoading || updateMutation.isLoading}
            onClick={handleSave}
          >
            Save
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={deleteConfirm.open} onClose={() => setDeleteConfirm({ open: false, id: null })}>
        <DialogTitle>Delete project?</DialogTitle>
        <DialogActions>
          <Button onClick={() => setDeleteConfirm({ open: false, id: null })}>Cancel</Button>
          <Button color="error" onClick={() => { deleteMutation.mutate(deleteConfirm.id); setDeleteConfirm({ open: false, id: null }); }}>
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default ScholarForge;
