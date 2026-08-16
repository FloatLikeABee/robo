import React from 'react';
import {
  Modal,
  View,
  Text,
  TouchableOpacity,
  ScrollView,
  StyleSheet,
  Platform,
  Linking,
} from 'react-native';
import Markdown from 'react-native-markdown-display';
import useAppStore from '../store/appStore';

export default function DocViewerModal({ visible, doc, apiBase, onClose }) {
  const { theme } = useAppStore();
  if (!doc) return null;

  const t = (doc.type || '').toLowerCase();
  const fileUrl =
    doc.id && apiBase ? `${apiBase}/docs/${encodeURIComponent(doc.id)}/file` : '';

  const mdStyles = {
    body: { color: theme.colors.textPrimary, fontSize: 14, lineHeight: 22 },
    heading1: { color: theme.colors.textPrimary, marginTop: 10, fontSize: 20, fontWeight: '700' },
    heading2: { color: theme.colors.textPrimary, marginTop: 8, fontSize: 17, fontWeight: '700' },
    paragraph: { marginBottom: 10, color: theme.colors.textPrimary },
    code_inline: {
      backgroundColor: 'rgba(255,255,255,0.08)',
      paddingHorizontal: 4,
      borderRadius: 4,
      fontFamily: Platform.select({ ios: 'Menlo', android: 'monospace', default: 'monospace' }),
    },
    fence: {
      backgroundColor: 'rgba(0,0,0,0.35)',
      padding: 10,
      borderRadius: 8,
      marginBottom: 10,
    },
    code_block: {
      backgroundColor: 'rgba(0,0,0,0.35)',
      padding: 10,
      borderRadius: 8,
      marginBottom: 10,
      fontFamily: Platform.select({ ios: 'Menlo', android: 'monospace', default: 'monospace' }),
    },
    link: { color: theme.colors.primary },
  };

  const openInBrowser = () => {
    if (!fileUrl) return;
    void Linking.openURL(fileUrl);
  };

  return (
    <Modal visible={visible} animationType="fade" transparent onRequestClose={onClose}>
      <View style={styles.backdrop}>
        <View
          style={[
            styles.panel,
            { backgroundColor: theme.colors.bgSecondary, borderColor: theme.colors.borderSubtle },
          ]}
        >
          <View style={[styles.header, { borderBottomColor: theme.colors.borderSubtle }]}>
            <Text style={[styles.title, { color: theme.colors.textPrimary }]} numberOfLines={2}>
              {doc.title || 'Document'}
            </Text>
            <TouchableOpacity onPress={onClose} hitSlop={{ top: 12, bottom: 12, left: 12, right: 12 }}>
              <Text style={{ color: theme.colors.textSecondary, fontSize: 18 }}>✕</Text>
            </TouchableOpacity>
          </View>
          <ScrollView style={styles.scroll} contentContainerStyle={styles.scrollInner}>
            {t === 'pdf' || t === 'image' ? (
              <View>
                <Text style={[styles.plain, { color: theme.colors.textSecondary }]}>
                  {doc.ai_summary || (t === 'pdf' ? 'PDF document.' : 'Image.')}
                </Text>
                {fileUrl ? (
                  <TouchableOpacity
                    style={[styles.openBtn, { borderColor: theme.colors.primary, backgroundColor: 'rgba(91, 140, 255, 0.15)' }]}
                    onPress={openInBrowser}
                    activeOpacity={0.8}
                  >
                    <Text style={[styles.openBtnText, { color: theme.colors.primary }]}>
                      {t === 'pdf' ? 'Open PDF in browser' : 'Open in browser'}
                    </Text>
                  </TouchableOpacity>
                ) : null}
              </View>
            ) : (
              <Markdown style={mdStyles}>{doc.content || doc.ai_summary || '(No body)'}</Markdown>
            )}
          </ScrollView>
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
    paddingHorizontal: 14,
  },
  panel: {
    maxHeight: '90%',
    borderRadius: 16,
    borderWidth: 1,
    overflow: 'hidden',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 14,
    borderBottomWidth: 1,
  },
  title: {
    flex: 1,
    fontSize: 17,
    fontWeight: '700',
    paddingRight: 8,
  },
  scroll: {
    maxHeight: 480,
  },
  scrollInner: {
    padding: 14,
    paddingBottom: 24,
  },
  plain: {
    fontSize: 14,
    lineHeight: 22,
    marginBottom: 16,
  },
  openBtn: {
    alignSelf: 'flex-start',
    paddingVertical: 12,
    paddingHorizontal: 18,
    borderRadius: 12,
    borderWidth: 1,
  },
  openBtnText: {
    fontSize: 14,
    fontWeight: '700',
  },
});
