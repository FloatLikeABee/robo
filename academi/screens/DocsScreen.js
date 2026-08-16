import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  Image,
  Dimensions,
  AppState,
  Alert,
  Vibration,
} from 'react-native';
import { useRoute, useNavigation, useFocusEffect } from '@react-navigation/native';
import NeuralSearchBar from '../components/NeuralSearchBar';
import LearnAnalysisModal from '../components/LearnAnalysisModal';
import DocViewerModal from '../components/DocViewerModal';
import useAppStore from '../store/appStore';
import {
  getAcademiApiBaseUrl,
  ensureAcademiSession,
  getAcademiToken,
  fetchDocsBrief,
  fetchDocById,
  postLearnAnalysis,
  createMarkdownDoc,
} from '../services/academiApi';

const { width } = Dimensions.get('window');

const MOCK_DOCS = [
  {
    id: '1',
    title: 'Quantum Computing Fundamentals',
    type: 'pdf',
    thumbnail: 'https://via.placeholder.com/150',
    tags: ['#quantum', '#computing', '#physics'],
    aiSummary: 'Introduction to quantum bits, superposition, and quantum gates.',
    uploadedBy: 'Alex Chen',
    uploadedAt: '2 days ago',
    size: '2.4 MB',
  },
  {
    id: '2',
    title: 'Neural Networks Diagram',
    type: 'image',
    thumbnail: 'https://via.placeholder.com/150',
    tags: ['#ai', '#machinelearning', '#diagram'],
    aiSummary: 'Visual representation of a multi-layer neural network architecture.',
    uploadedBy: 'Samira Patel',
    uploadedAt: '5 days ago',
    size: '1.2 MB',
  },
  {
    id: '3',
    title: 'Research Methodology',
    type: 'video',
    thumbnail: 'https://via.placeholder.com/150',
    tags: ['#research', '#methods', '#education'],
    aiSummary: 'Overview of qualitative and quantitative research methodologies.',
    uploadedBy: 'Jordan Kim',
    uploadedAt: '1 week ago',
    size: '45.6 MB',
  },
  {
    id: '4',
    title: 'Linear Algebra Notes',
    type: 'markdown',
    thumbnail: 'https://via.placeholder.com/150',
    tags: ['#math', '#algebra', '#notes'],
    aiSummary: 'Comprehensive notes on vector spaces and matrix operations.',
    uploadedBy: 'Taylor Wong',
    uploadedAt: '3 days ago',
    size: '890 KB',
  },
  {
    id: '5',
    title: 'Organic Chemistry Reactions',
    type: 'pdf',
    thumbnail: 'https://via.placeholder.com/150',
    tags: ['#chemistry', '#organic', '#reactions'],
    aiSummary: 'Chart of common organic chemistry reactions and mechanisms.',
    uploadedBy: 'Maya Singh',
    uploadedAt: '4 days ago',
    size: '3.1 MB',
  },
  {
    id: '6',
    title: 'Data Structures in Python',
    type: 'markdown',
    thumbnail: 'https://via.placeholder.com/150',
    tags: ['#cs', '#python', '#algorithms'],
    aiSummary: 'Reference guide for common data structures and their implementations.',
    uploadedBy: 'Ryan Lee',
    uploadedAt: '1 week ago',
    size: '1.5 MB',
  },
];

const CATEGORY_FILTERS = ['All', 'PDF', 'Image', 'Video', 'Document'];

function isDocLearnable(doc) {
  const t = (doc.type || '').toLowerCase();
  return ['markdown', 'text', 'pdf', 'image'].includes(t);
}

function mapApiDoc(d) {
  const thumb = d.thumbnail && String(d.thumbnail).trim()
    ? d.thumbnail
    : 'https://via.placeholder.com/150';
  return {
    id: String(d.id),
    title: d.title || 'Untitled',
    type: (d.type || 'document').toLowerCase(),
    thumbnail: thumb,
    tags: Array.isArray(d.tags) ? d.tags : [],
    aiSummary: d.ai_summary || '',
    uploadedBy: 'You',
    uploadedAt: '',
    size: d.size || '—',
  };
}

export default function DocsScreen() {
  const route = useRoute();
  const navigation = useNavigation();
  const { theme, addNotification } = useAppStore();
  const [searchTerm, setSearchTerm] = useState('');
  const [viewMode, setViewMode] = useState('grid');
  const [activeCategory, setActiveCategory] = useState('All');
  const [docs, setDocs] = useState(MOCK_DOCS);

  const [learnModalVisible, setLearnModalVisible] = useState(false);
  const [learnPhase, setLearnPhase] = useState('idle');
  const [learnSheetTitle, setLearnSheetTitle] = useState('Help you learn');
  const [learnBody, setLearnBody] = useState('');
  const [learnSucceeded, setLearnSucceeded] = useState(true);
  const [learnSaving, setLearnSaving] = useState(false);

  const [docViewerVisible, setDocViewerVisible] = useState(false);
  const [docViewerDoc, setDocViewerDoc] = useState(null);

  const learnJobSeqRef = useRef(0);
  const learnJobIdRef = useRef(0);
  const learnNotifyJobIdRef = useRef(null);
  const learnFetchPendingRef = useRef(false);
  const learnModalVisibleRef = useRef(false);
  const pendingLearnAlertRef = useRef(null);
  const apiBaseRef = useRef(getAcademiApiBaseUrl());

  useEffect(() => {
    learnModalVisibleRef.current = learnModalVisible;
  }, [learnModalVisible]);

  const flushPendingLearnAlert = useCallback(() => {
    const p = pendingLearnAlertRef.current;
    if (!p) return;
    pendingLearnAlertRef.current = null;
    Alert.alert(p.title, p.msg, [
      {
        text: 'View',
        onPress: () => {
          setLearnModalVisible(true);
        },
      },
      { text: 'OK', style: 'cancel' },
    ]);
  }, []);

  const openDocFromCommunityLink = useCallback(async (docId) => {
    const id = String(docId || '').trim();
    if (!id) return;
    const apiBase = apiBaseRef.current;
    try {
      await ensureAcademiSession(apiBase);
      const tok = getAcademiToken();
      try {
        const raw = await fetchDocsBrief(apiBase, tok);
        if (Array.isArray(raw) && raw.length > 0) setDocs(raw.map(mapApiDoc));
      } catch (_) {
        /* keep current list */
      }
      const full = await fetchDocById(apiBase, tok, id);
      setDocs((prev) => {
        const mapped = mapApiDoc(full);
        const ix = prev.findIndex((d) => String(d.id) === String(id));
        if (ix >= 0) {
          const next = [...prev];
          next[ix] = mapped;
          return next;
        }
        return [...prev, mapped];
      });
      setDocViewerDoc(full);
      setDocViewerVisible(true);
    } catch (e) {
      Alert.alert('Document', e.message || 'Could not download document');
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      const raw = route.params?.openDocId;
      if (raw == null || raw === '') return;
      const id = String(raw);
      navigation.setParams({ openDocId: undefined });
      void openDocFromCommunityLink(id);
    }, [route.params?.openDocId, navigation, openDocFromCommunityLink]),
  );

  useEffect(() => {
    const sub = AppState.addEventListener('change', (s) => {
      if (s === 'active') flushPendingLearnAlert();
    });
    return () => sub.remove();
  }, [flushPendingLearnAlert]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      const apiBase = apiBaseRef.current;
      try {
        await ensureAcademiSession(apiBase);
        const tok = getAcademiToken();
        const raw = await fetchDocsBrief(apiBase, tok);
        if (cancelled || !Array.isArray(raw) || raw.length === 0) return;
        setDocs(raw.map(mapApiDoc));
      } catch {
        /* keep mock catalog when API unreachable */
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const pushLearnNotification = useCallback(
    (ok, docTitle, errMsg) => {
      const line = ok
        ? `Analysis finished for “${docTitle}”.`
        : errMsg || 'Analysis failed.';
      addNotification({
        id: Date.now(),
        type: 'learn_ready',
        title: ok ? 'Help you learn — ready' : 'Help you learn — error',
        description: line,
        timestamp: 'Just now',
        read: false,
        icon: '📚',
      });
    },
    [addNotification],
  );

  const maybeAlertLearnDone = useCallback(
    (ok, docTitle, errMsg) => {
      const title = ok ? 'Help you learn — ready' : 'Help you learn — error';
      const msg = ok ? `Analysis finished for “${docTitle}”.` : errMsg || 'Analysis failed.';
      pushLearnNotification(ok, docTitle, errMsg);

      if (AppState.currentState !== 'active') {
        pendingLearnAlertRef.current = { title, msg, ok };
        Vibration.vibrate(400);
        return;
      }

      Alert.alert(title, msg, [
        {
          text: 'View',
          onPress: () => setLearnModalVisible(true),
        },
        { text: 'OK', style: 'cancel' },
      ]);
    },
    [pushLearnNotification],
  );

  const closeLearnModal = useCallback(() => {
    if (learnFetchPendingRef.current) {
      learnNotifyJobIdRef.current = learnJobIdRef.current;
    }
    setLearnModalVisible(false);
  }, []);

  const runLearnBackground = useCallback(() => {
    closeLearnModal();
  }, [closeLearnModal]);

  const saveLearnToDocs = useCallback(async () => {
    if (!learnBody.trim() || !learnSucceeded) return;
    const apiBase = apiBaseRef.current;
    setLearnSaving(true);
    try {
      await ensureAcademiSession(apiBase);
      const tok = getAcademiToken();
      await createMarkdownDoc(apiBase, tok, {
        title: learnSheetTitle.replace(/^Learning:\s*/, 'Learning notes — ') || 'Learning analysis',
        content: learnBody,
      });
      setLearnModalVisible(false);
      setLearnPhase('idle');
      const raw = await fetchDocsBrief(apiBase, getAcademiToken());
      if (Array.isArray(raw) && raw.length > 0) setDocs(raw.map(mapApiDoc));
    } catch (e) {
      Alert.alert('Save failed', e.message || 'Could not save');
    } finally {
      setLearnSaving(false);
    }
  }, [learnBody, learnSheetTitle, learnSucceeded]);

  const runLearnForDoc = useCallback(
    async (docId) => {
      const doc = docs.find((d) => String(d.id) === String(docId));
      learnJobSeqRef.current += 1;
      const jobId = learnJobSeqRef.current;
      learnJobIdRef.current = jobId;
      learnNotifyJobIdRef.current = null;
      learnFetchPendingRef.current = true;

      const headingDocTitle = doc ? doc.title : 'Document';
      setLearnSheetTitle(doc ? `Help you learn · ${doc.title}` : 'Help you learn');
      setLearnBody('');
      setLearnPhase('analyzing');
      setLearnSucceeded(true);
      setLearnModalVisible(true);

      const finish = (ok, textOrErr) => {
        if (jobId !== learnJobSeqRef.current) return;

        learnFetchPendingRef.current = false;
        const modalHidden = !learnModalVisibleRef.current;
        const appHidden = AppState.currentState !== 'active';
        const dismissedThisJob = learnNotifyJobIdRef.current === jobId;
        const shouldNotify = appHidden || dismissedThisJob || modalHidden;

        setLearnSucceeded(ok);
        if (ok) {
          setLearnBody(textOrErr || '');
          setLearnSheetTitle(`Learning: ${headingDocTitle}`);
        } else {
          setLearnBody(textOrErr || 'Error');
          setLearnSheetTitle(`Help you learn · ${headingDocTitle}`);
        }
        setLearnPhase('result');

        if (learnNotifyJobIdRef.current === jobId) {
          learnNotifyJobIdRef.current = null;
        }

        if (shouldNotify) {
          maybeAlertLearnDone(ok, headingDocTitle, ok ? null : String(textOrErr || 'Error'));
        }
      };

      const apiBase = apiBaseRef.current;
      try {
        await ensureAcademiSession(apiBase);
        const tok = getAcademiToken();
        const data = await postLearnAnalysis(apiBase, tok, docId, { disableResearch: false });
        if (jobId !== learnJobSeqRef.current) return;
        finish(true, data.response || '');
      } catch (e) {
        if (jobId !== learnJobSeqRef.current) return;
        finish(false, e.message || 'Error');
      }
    },
    [docs, maybeAlertLearnDone],
  );

  const filteredDocs = docs.filter((doc) => {
    const matchesSearch =
      doc.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      doc.tags.some((tag) => tag.toLowerCase().includes(searchTerm.toLowerCase())) ||
      doc.aiSummary.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesCategory =
      activeCategory === 'All' || doc.type.toLowerCase() === activeCategory.toLowerCase();
    return matchesSearch && matchesCategory;
  });

  const themedStyles = stylesForTheme(theme);

  const renderGridItem = ({ item }) => (
    <View style={themedStyles.gridItem}>
      <TouchableOpacity style={themedStyles.gridCard} activeOpacity={0.7}>
        <View style={themedStyles.thumbnailWrapper}>
          <Image source={{ uri: item.thumbnail }} style={themedStyles.thumbnail} resizeMode="cover" />
          <View style={themedStyles.typeBadge}>
            <Text style={themedStyles.typeBadgeText}>{item.type.toUpperCase()}</Text>
          </View>
        </View>
        <View style={themedStyles.gridCardContent}>
          <Text style={themedStyles.gridCardTitle} numberOfLines={2}>
            {item.title}
          </Text>
          <Text style={themedStyles.gridCardSummary} numberOfLines={2}>
            {item.aiSummary}
          </Text>
          <View style={themedStyles.gridCardTags}>
            {item.tags.slice(0, 2).map((tag, index) => (
              <Text key={index} style={themedStyles.gridTag}>
                {tag}
              </Text>
            ))}
          </View>
          {isDocLearnable(item) ? (
            <TouchableOpacity
              style={themedStyles.learnBtn}
              onPress={() => runLearnForDoc(item.id)}
              activeOpacity={0.8}
            >
              <Text style={themedStyles.learnBtnText}>Help you learn</Text>
            </TouchableOpacity>
          ) : null}
        </View>
      </TouchableOpacity>
    </View>
  );

  const renderListItem = ({ item }) => (
    <View style={themedStyles.listRow}>
      <TouchableOpacity style={themedStyles.listCard} activeOpacity={0.7}>
        <View style={themedStyles.listCardLeft}>
          <View style={themedStyles.listCardIcon}>
            <Text style={themedStyles.listCardIconText}>
              {item.type === 'pdf' ? '📄' : item.type === 'image' ? '🖼️' : item.type === 'video' ? '🎬' : '📝'}
            </Text>
          </View>
          <View style={themedStyles.listCardInfo}>
            <Text style={themedStyles.listCardTitle}>{item.title}</Text>
            <Text style={themedStyles.listCardMeta}>
              {item.uploadedBy} · {item.size}
            </Text>
            <View style={themedStyles.listCardTags}>
              {item.tags.slice(0, 2).map((tag, index) => (
                <Text key={index} style={themedStyles.listTag}>
                  {tag}
                </Text>
              ))}
            </View>
          </View>
        </View>
        <Text style={themedStyles.listCardChevron}>›</Text>
      </TouchableOpacity>
      {isDocLearnable(item) ? (
        <TouchableOpacity
          style={[themedStyles.learnBtn, themedStyles.learnBtnList]}
          onPress={() => runLearnForDoc(item.id)}
          activeOpacity={0.8}
        >
          <Text style={themedStyles.learnBtnText}>Help you learn</Text>
        </TouchableOpacity>
      ) : null}
    </View>
  );

  const learnModalPhase = learnPhase === 'analyzing' ? 'analyzing' : 'result';

  return (
    <View style={themedStyles.container}>
      <View style={themedStyles.header}>
        <Text style={themedStyles.headerTitle}>Docs</Text>
        <Text style={themedStyles.headerSubtitle}>{filteredDocs.length} documents</Text>
        <View style={themedStyles.headerActions}>
          <TouchableOpacity
            style={[themedStyles.viewToggle, viewMode === 'grid' && themedStyles.viewToggleActive]}
            onPress={() => setViewMode('grid')}
            activeOpacity={0.7}
          >
            <Text style={themedStyles.viewToggleText}>⊞</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[themedStyles.viewToggle, viewMode === 'list' && themedStyles.viewToggleActive]}
            onPress={() => setViewMode('list')}
            activeOpacity={0.7}
          >
            <Text style={themedStyles.viewToggleText}>☰</Text>
          </TouchableOpacity>
        </View>
      </View>

      <NeuralSearchBar
        placeholder="Search documents, topics, authors..."
        value={searchTerm}
        onChangeText={setSearchTerm}
      />

      <View style={themedStyles.categoryBar}>
        {CATEGORY_FILTERS.map((cat) => (
          <TouchableOpacity
            key={cat}
            style={[themedStyles.categoryButton, activeCategory === cat && themedStyles.categoryButtonActive]}
            onPress={() => setActiveCategory(cat)}
            activeOpacity={0.7}
          >
            <Text
              style={[themedStyles.categoryText, activeCategory === cat && themedStyles.categoryTextActive]}
            >
              {cat}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      {viewMode === 'grid' ? (
        <FlatList
          data={filteredDocs}
          keyExtractor={(item) => item.id.toString()}
          numColumns={2}
          contentContainerStyle={themedStyles.gridContent}
          renderItem={renderGridItem}
          showsVerticalScrollIndicator={false}
        />
      ) : (
        <FlatList
          data={filteredDocs}
          keyExtractor={(item) => item.id.toString()}
          contentContainerStyle={themedStyles.listContent}
          renderItem={renderListItem}
          showsVerticalScrollIndicator={false}
        />
      )}

      <TouchableOpacity style={themedStyles.fab} activeOpacity={0.8}>
        <Text style={themedStyles.fabText}>+</Text>
      </TouchableOpacity>

      <LearnAnalysisModal
        visible={learnModalVisible}
        phase={learnModalPhase}
        title={learnSheetTitle}
        bodyText={learnBody}
        succeeded={learnSucceeded}
        saving={learnSaving}
        onClose={closeLearnModal}
        onBackground={runLearnBackground}
        onSave={saveLearnToDocs}
      />
      <DocViewerModal
        visible={docViewerVisible}
        doc={docViewerDoc}
        apiBase={apiBaseRef.current}
        onClose={() => {
          setDocViewerVisible(false);
          setDocViewerDoc(null);
        }}
      />
    </View>
  );
}

function stylesForTheme(theme) {
  const c = theme.colors;
  return StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: c.bgPrimary,
    },
    header: {
      padding: 20,
      paddingTop: 48,
      flexDirection: 'row',
      alignItems: 'flex-end',
      justifyContent: 'space-between',
    },
    headerTitle: {
      color: c.textPrimary,
      fontSize: 28,
      fontWeight: '700',
    },
    headerSubtitle: {
      color: c.textSecondary,
      fontSize: 13,
      flex: 1,
      marginLeft: 12,
    },
    headerActions: {
      flexDirection: 'row',
      gap: 8,
    },
    viewToggle: {
      width: 36,
      height: 36,
      borderRadius: 12,
      backgroundColor: 'rgba(255,255,255,0.05)',
      justifyContent: 'center',
      alignItems: 'center',
    },
    viewToggleActive: {
      backgroundColor: 'rgba(91, 140, 255, 0.2)',
    },
    viewToggleText: {
      color: c.textPrimary,
      fontSize: 18,
      fontWeight: '700',
    },
    categoryBar: {
      flexDirection: 'row',
      paddingHorizontal: 16,
      paddingVertical: 8,
    },
    categoryButton: {
      paddingHorizontal: 14,
      paddingVertical: 8,
      borderRadius: 16,
      backgroundColor: 'rgba(255,255,255,0.04)',
      marginRight: 8,
    },
    categoryButtonActive: {
      backgroundColor: 'rgba(0, 212, 255, 0.15)',
      borderWidth: 1,
      borderColor: 'rgba(0, 212, 255, 0.3)',
    },
    categoryText: {
      color: c.textSecondary,
      fontSize: 12,
      fontWeight: '500',
    },
    categoryTextActive: {
      color: c.tertiary,
      fontWeight: '600',
    },
    gridContent: {
      padding: 12,
      paddingBottom: 80,
    },
    gridItem: {
      flex: 1,
      margin: 4,
      maxWidth: (width - 32) / 2,
    },
    gridCard: {
      backgroundColor: 'rgba(255,255,255,0.03)',
      borderRadius: 16,
      overflow: 'hidden',
      borderWidth: 1,
      borderColor: 'rgba(255,255,255,0.06)',
    },
    thumbnailWrapper: {
      height: 120,
      position: 'relative',
    },
    thumbnail: {
      width: '100%',
      height: '100%',
      backgroundColor: 'rgba(255,255,255,0.05)',
    },
    typeBadge: {
      position: 'absolute',
      top: 8,
      right: 8,
      backgroundColor: 'rgba(10, 15, 28, 0.7)',
      borderRadius: 8,
      paddingHorizontal: 8,
      paddingVertical: 4,
    },
    typeBadgeText: {
      color: c.textPrimary,
      fontSize: 9,
      fontWeight: '700',
      letterSpacing: 0.5,
    },
    gridCardContent: {
      padding: 12,
    },
    gridCardTitle: {
      color: c.textPrimary,
      fontSize: 14,
      fontWeight: '600',
      marginBottom: 6,
      lineHeight: 18,
    },
    gridCardSummary: {
      color: c.textSecondary,
      fontSize: 12,
      lineHeight: 16,
      marginBottom: 8,
    },
    gridCardTags: {
      flexDirection: 'row',
      flexWrap: 'wrap',
      marginBottom: 8,
    },
    gridTag: {
      color: c.primary,
      fontSize: 10,
      marginRight: 8,
      fontWeight: '500',
    },
    learnBtn: {
      alignSelf: 'flex-start',
      marginTop: 4,
      paddingVertical: 8,
      paddingHorizontal: 12,
      borderRadius: 12,
      borderWidth: 1,
      borderColor: c.primary,
      backgroundColor: 'rgba(91, 140, 255, 0.12)',
    },
    learnBtnList: {
      alignSelf: 'stretch',
      marginHorizontal: 14,
      marginBottom: 10,
      alignItems: 'center',
    },
    learnBtnText: {
      color: c.primary,
      fontSize: 11,
      fontWeight: '700',
    },
    listContent: {
      padding: 12,
      paddingBottom: 80,
    },
    listRow: {
      marginBottom: 8,
    },
    listCard: {
      flexDirection: 'row',
      alignItems: 'center',
      backgroundColor: 'rgba(255,255,255,0.03)',
      borderRadius: 16,
      padding: 14,
      borderWidth: 1,
      borderColor: 'rgba(255,255,255,0.06)',
    },
    listCardLeft: {
      flex: 1,
      flexDirection: 'row',
      alignItems: 'center',
    },
    listCardIcon: {
      width: 40,
      height: 40,
      borderRadius: 12,
      backgroundColor: 'rgba(255,255,255,0.05)',
      justifyContent: 'center',
      alignItems: 'center',
      marginRight: 12,
    },
    listCardIconText: {
      fontSize: 20,
    },
    listCardInfo: {
      flex: 1,
    },
    listCardTitle: {
      color: c.textPrimary,
      fontSize: 15,
      fontWeight: '600',
      marginBottom: 4,
    },
    listCardMeta: {
      color: c.textSecondary,
      fontSize: 12,
      marginBottom: 6,
    },
    listCardTags: {
      flexDirection: 'row',
    },
    listTag: {
      color: c.primary,
      fontSize: 11,
      marginRight: 8,
      fontWeight: '500',
    },
    listCardChevron: {
      color: 'rgba(255,255,255,0.2)',
      fontSize: 22,
      marginLeft: 8,
    },
    fab: {
      position: 'absolute',
      bottom: 24,
      right: 24,
      width: 56,
      height: 56,
      borderRadius: 28,
      backgroundColor: c.primary,
      justifyContent: 'center',
      alignItems: 'center',
      shadowColor: c.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.4,
      shadowRadius: 8,
      elevation: 6,
    },
    fabText: {
      color: '#FFFFFF',
      fontSize: 28,
      fontWeight: '300',
    },
  });
}
