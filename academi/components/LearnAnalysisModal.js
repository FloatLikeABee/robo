import React from 'react';
import {
  Modal,
  View,
  Text,
  TouchableOpacity,
  ScrollView,
  StyleSheet,
  ActivityIndicator,
  Platform,
} from 'react-native';
import Markdown from 'react-native-markdown-display';
import useAppStore from '../store/appStore';

export default function LearnAnalysisModal({
  visible,
  phase,
  title,
  bodyText,
  succeeded,
  saving,
  onClose,
  onBackground,
  onSave,
}) {
  const { theme } = useAppStore();
  const analyzing = phase === 'analyzing';

  const mdStyles = {
    body: { color: theme.colors.textPrimary, fontSize: 14, lineHeight: 22 },
    heading1: { color: theme.colors.textPrimary, marginTop: 12, marginBottom: 8, fontSize: 22, fontWeight: '700' },
    heading2: { color: theme.colors.textPrimary, marginTop: 10, marginBottom: 6, fontSize: 18, fontWeight: '700' },
    heading3: { color: theme.colors.textPrimary, marginTop: 8, marginBottom: 4, fontSize: 16, fontWeight: '600' },
    paragraph: { marginTop: 0, marginBottom: 10, color: theme.colors.textPrimary },
    bullet_list: { marginBottom: 8 },
    ordered_list: { marginBottom: 8 },
    list_item: { color: theme.colors.textPrimary, marginBottom: 4 },
    code_inline: {
      backgroundColor: 'rgba(255,255,255,0.08)',
      color: theme.colors.tertiary,
      paddingHorizontal: 4,
      borderRadius: 4,
      fontFamily: Platform.select({ ios: 'Menlo', android: 'monospace', default: 'monospace' }),
    },
    fence: {
      backgroundColor: 'rgba(0,0,0,0.35)',
      color: theme.colors.textPrimary,
      padding: 12,
      borderRadius: 8,
      marginBottom: 10,
    },
    code_block: {
      backgroundColor: 'rgba(0,0,0,0.35)',
      color: theme.colors.textPrimary,
      padding: 12,
      borderRadius: 8,
      marginBottom: 10,
      fontFamily: Platform.select({ ios: 'Menlo', android: 'monospace', default: 'monospace' }),
    },
    link: { color: theme.colors.primary },
    blockquote: {
      borderLeftWidth: 3,
      borderLeftColor: theme.colors.primary,
      paddingLeft: 10,
      marginBottom: 10,
      color: theme.colors.textSecondary,
    },
  };

  return (
    <Modal visible={visible} animationType="fade" transparent onRequestClose={onClose}>
      <View style={styles.backdrop}>
        <View style={[styles.panel, { backgroundColor: theme.colors.bgSecondary, borderColor: theme.colors.borderSubtle }]}>
          <View style={[styles.header, { borderBottomColor: theme.colors.borderSubtle }]}>
            <Text style={[styles.headerTitle, { color: theme.colors.textPrimary }]} numberOfLines={2}>
              {title || 'Help you learn'}
            </Text>
            <TouchableOpacity onPress={onClose} style={styles.closeBtn} hitSlop={{ top: 12, bottom: 12, left: 12, right: 12 }}>
              <Text style={{ color: theme.colors.textSecondary, fontSize: 18 }}>✕</Text>
            </TouchableOpacity>
          </View>

          {analyzing ? (
            <View style={styles.analyzingWrap}>
              <ActivityIndicator size="large" color={theme.colors.primary} />
              <Text style={[styles.analyzingText, { color: theme.colors.textSecondary }]}>Analyzing…</Text>
            </View>
          ) : (
            <ScrollView style={styles.scroll} contentContainerStyle={styles.scrollContent} keyboardShouldPersistTaps="handled">
              {succeeded ? (
                <Markdown style={mdStyles}>{bodyText || ''}</Markdown>
              ) : (
                <Text style={[styles.errorText, { color: theme.colors.textPrimary }]}>{bodyText || ''}</Text>
              )}
            </ScrollView>
          )}

          <View style={[styles.footer, { borderTopColor: theme.colors.borderSubtle }]}>
            {analyzing && (
              <TouchableOpacity
                style={[styles.btnSecondary, { borderColor: theme.colors.borderSubtle }]}
                onPress={onBackground}
                activeOpacity={0.75}
              >
                <Text style={[styles.btnSecondaryText, { color: theme.colors.textPrimary }]}>Run in background</Text>
              </TouchableOpacity>
            )}
            <TouchableOpacity
              style={[styles.btnSecondary, { borderColor: theme.colors.borderSubtle }]}
              onPress={onClose}
              activeOpacity={0.75}
            >
              <Text style={[styles.btnSecondaryText, { color: theme.colors.textPrimary }]}>Close</Text>
            </TouchableOpacity>
            {!analyzing && succeeded && (
              <TouchableOpacity
                style={[styles.btnPrimary, { backgroundColor: theme.colors.primary }]}
                onPress={onSave}
                disabled={saving}
                activeOpacity={0.8}
              >
                <Text style={[styles.btnPrimaryText, { color: '#FFFFFF' }]}>{saving ? 'Saving…' : 'Save to Docs'}</Text>
              </TouchableOpacity>
            )}
          </View>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    flex: 1,
    backgroundColor: 'rgba(10,15,28,0.72)',
    justifyContent: 'center',
    paddingHorizontal: 16,
  },
  panel: {
    maxHeight: '88%',
    borderRadius: 16,
    borderWidth: 1,
    overflow: 'hidden',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
  },
  headerTitle: {
    flex: 1,
    fontSize: 17,
    fontWeight: '700',
    paddingRight: 8,
  },
  closeBtn: {
    padding: 4,
  },
  analyzingWrap: {
    minHeight: 160,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 24,
    gap: 12,
  },
  analyzingText: {
    fontSize: 14,
  },
  scroll: {
    maxHeight: 420,
  },
  scrollContent: {
    padding: 16,
    paddingBottom: 24,
  },
  errorText: {
    fontSize: 14,
    lineHeight: 22,
    fontFamily: 'monospace',
  },
  footer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'flex-end',
    gap: 8,
    padding: 12,
    borderTopWidth: 1,
  },
  btnSecondary: {
    paddingVertical: 10,
    paddingHorizontal: 14,
    borderRadius: 12,
    borderWidth: 1,
  },
  btnSecondaryText: {
    fontSize: 13,
    fontWeight: '600',
  },
  btnPrimary: {
    paddingVertical: 10,
    paddingHorizontal: 16,
    borderRadius: 12,
  },
  btnPrimaryText: {
    fontSize: 13,
    fontWeight: '700',
  },
});
