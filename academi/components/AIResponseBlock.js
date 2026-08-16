import React, { useState, useRef, useEffect } from 'react';
import { View, Text, StyleSheet, Animated, TouchableOpacity, Dimensions } from 'react-native';
import useAppStore from '../store/appStore';

const { width } = Dimensions.get('window');

export default function AIResponseBlock({
  text,
  sources = [],
  expanded = false,
  onToggleExpand,
  onSave,
  onShare,
  onCopy,
  showActions = true
}) {
  const { isDarkMode } = useAppStore();
  const copyAnim = useRef(new Animated.Value(0)).current;

  const handleCopy = () => {
    Animated.sequence([
      Animated.timing(copyAnim, { toValue: 1, duration: 200, useNativeDriver: true }),
      Animated.timing(copyAnim, { toValue: 0, duration: 200, useNativeDriver: true }),
    ]).start();
    onCopy && onCopy();
  };

  return (
    <View style={styles.container}>
      <View style={styles.aiHeader}>
        <View style={styles.aiDotContainer}>
          <View style={styles.aiDot} />
          <View style={styles.aiDotGlow} />
        </View>
        <Text style={styles.aiLabel}>Academi AI</Text>
      </View>

      <View style={styles.responseBody}>
        <Text style={styles.responseText}>{text}</Text>
      </View>

      {sources.length > 0 && (
        <View style={styles.sourcesSection}>
          <TouchableOpacity style={styles.sourcesHeader} onPress={onToggleExpand}>
            <Text style={styles.sourcesTitle}>Sources ({sources.length})</Text>
            <Text style={styles.sourcesToggleIcon}>{expanded ? '−' : '+'}</Text>
          </TouchableOpacity>

          {expanded && (
            <View style={styles.sourcesList}>
              {sources.map((source, index) => (
                <View key={index} style={styles.sourceItem}>
                  <Text style={styles.sourceIcon}>
                    {source.type === 'web' ? '🌐' : source.type === 'doc' ? '📄' : '🎬'}
                  </Text>
                  <View style={styles.sourceInfo}>
                    <Text style={styles.sourceName}>{source.title}</Text>
                    <Text style={styles.sourceType}>{source.type}</Text>
                  </View>
                </View>
              ))}
            </View>
          )}
        </View>
      )}

      {showActions && (
        <View style={styles.actionBar}>
          <TouchableOpacity style={styles.actionItem} onPress={onSave} activeOpacity={0.6}>
            <Text style={styles.actionIcon}>🔖</Text>
            <Text style={styles.actionLabel}>Save</Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.actionItem} onPress={onShare} activeOpacity={0.6}>
            <Text style={styles.actionIcon}>↗️</Text>
            <Text style={styles.actionLabel}>Share</Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.actionItem} onPress={handleCopy} activeOpacity={0.6}>
            <Text style={styles.actionIcon}>📋</Text>
            <Text style={styles.actionLabel}>Copy</Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.actionItem} onPress={() => {}} activeOpacity={0.6}>
            <Text style={styles.actionIcon}>💬</Text>
            <Text style={styles.actionLabel}>Cite</Text>
          </TouchableOpacity>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    marginBottom: 16,
  },
  aiHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  aiDotContainer: {
    width: 20,
    height: 20,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 8,
  },
  aiDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#5B8CFF',
    position: 'absolute',
  },
  aiDotGlow: {
    width: 20,
    height: 20,
    borderRadius: 10,
    backgroundColor: 'rgba(91, 140, 255, 0.15)',
    position: 'absolute',
  },
  aiLabel: {
    color: '#A8B2D1',
    fontSize: 12,
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  responseBody: {
    backgroundColor: 'rgba(255,255,255,0.04)',
    borderRadius: 16,
    padding: 16,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.06)',
  },
  responseText: {
    color: '#FFFFFF',
    fontSize: 15,
    lineHeight: 22,
  },
  sourcesSection: {
    marginTop: 8,
    backgroundColor: 'rgba(255,255,255,0.02)',
    borderRadius: 12,
    overflow: 'hidden',
  },
  sourcesHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 12,
  },
  sourcesTitle: {
    color: '#A8B2D1',
    fontSize: 13,
    fontWeight: '600',
  },
  sourcesToggleIcon: {
    color: '#5B8CFF',
    fontSize: 16,
    fontWeight: '700',
  },
  sourcesList: {
    padding: 8,
    paddingBottom: 4,
  },
  sourceItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 10,
    backgroundColor: 'rgba(255,255,255,0.03)',
    borderRadius: 8,
    marginBottom: 4,
  },
  sourceIcon: {
    fontSize: 16,
    marginRight: 10,
  },
  sourceInfo: {
    flex: 1,
  },
  sourceName: {
    color: '#FFFFFF',
    fontSize: 13,
    fontWeight: '500',
  },
  sourceType: {
    color: '#A8B2D1',
    fontSize: 11,
    textTransform: 'capitalize',
  },
  actionBar: {
    flexDirection: 'row',
    marginTop: 8,
    justifyContent: 'space-around',
    paddingVertical: 8,
  },
  actionItem: {
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 12,
  },
  actionIcon: {
    fontSize: 16,
    marginBottom: 2,
  },
  actionLabel: {
    color: '#A8B2D1',
    fontSize: 11,
    fontWeight: '500',
  },
});
