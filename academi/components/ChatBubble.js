import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { darkTheme } from '../themes/theme';

export default function ChatBubble({ 
  text, 
  isUser = false, 
  hasSource = false, 
  expanded = false,
  onSourceToggle,
  onSave,
  onShare
}) {
  return (
    <View style={[styles.container, isUser && styles.userBubble, !isUser && styles.aiBubble]}>
      <Text style={styles.text}>{text}</Text>
      {hasSource && (
        <View style={styles.sourceContainer}>
          {!expanded ? (
            <View style={styles.sourcePreview}>
              <Text style={styles.sourceText}>Source preview...</Text>
            </View>
          ) : (
            <View style={styles.sourceExpanded}>
              <Text style={styles.sourceText}>Full source details...</Text>
            </View>
          )}
          <View style={styles.sourceActions}>
            <TouchableOpacity style={styles.actionButton} onPress={onSourceToggle} activeOpacity={0.6}>
              <Text style={styles.actionText}>{!expanded ? 'Expand' : 'Collapse'}</Text>
            </TouchableOpacity>
            <TouchableOpacity style={styles.actionButton} onPress={onSave} activeOpacity={0.6}>
              <Text style={styles.actionText}>Save</Text>
            </TouchableOpacity>
            <TouchableOpacity style={styles.actionButton} onPress={onShare} activeOpacity={0.6}>
              <Text style={styles.actionText}>Share</Text>
            </TouchableOpacity>
          </View>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    padding: 14,
    borderRadius: 16,
    maxWidth: '85%',
    marginVertical: 6,
    marginHorizontal: 12,
  },
  userBubble: {
    backgroundColor: '#5B8CFF',
    alignSelf: 'flex-end',
  },
  aiBubble: {
    backgroundColor: 'rgba(255,255,255,0.04)',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.06)',
    alignSelf: 'flex-start',
  },
  text: {
    color: '#FFFFFF',
    fontSize: 15,
    lineHeight: 22,
  },
  sourceContainer: {
    marginTop: 10,
  },
  sourcePreview: {
    backgroundColor: 'rgba(255,255,255,0.05)',
    padding: 10,
    borderRadius: 10,
  },
  sourceExpanded: {
    backgroundColor: 'rgba(255,255,255,0.05)',
    padding: 12,
    borderRadius: 10,
  },
  sourceText: {
    color: '#A8B2D1',
    fontSize: 13,
    lineHeight: 18,
  },
  sourceActions: {
    flexDirection: 'row',
    marginTop: 8,
    gap: 8,
  },
  actionButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 12,
    backgroundColor: 'rgba(255,255,255,0.06)',
  },
  actionText: {
    color: '#A8B2D1',
    fontSize: 12,
    fontWeight: '500',
  },
});