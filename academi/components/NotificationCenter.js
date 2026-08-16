import React, { useState } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, FlatList } from 'react-native';
import useAppStore from '../store/appStore';

const mockNotifications = [
  {
    id: 1,
    type: 'ai_response',
    title: 'AI Response Ready',
    description: 'Your quantum entanglement explanation is ready',
    timestamp: '2m ago',
    read: false,
    icon: '🤖',
  },
  {
    id: 2,
    type: 'community_reply',
    title: 'New Reply',
    description: 'Alex commented on your post about calculus',
    timestamp: '15m ago',
    read: false,
    icon: '💬',
  },
  {
    id: 3,
    type: 'doc_recommendation',
    title: 'Doc Recommendation',
    description: 'Based on your interests: "Advanced Linear Algebra"',
    timestamp: '1h ago',
    read: true,
    icon: '📚',
  },
  {
    id: 4,
    type: 'guide_progress',
    title: 'Guide Reminder',
    description: 'Continue your "Learn Calculus in 7 Days" guide - Day 3',
    timestamp: '3h ago',
    read: true,
    icon: '🧠',
  },
];

export default function NotificationCenter({ visible, onClose }) {
  const { markNotificationRead } = useAppStore();
  const unreadCount = mockNotifications.filter(n => !n.read).length;

  if (!visible) return null;

  const renderNotification = ({ item }) => (
    <TouchableOpacity
      style={[styles.item, !item.read && styles.itemUnread]}
      activeOpacity={0.6}
      onPress={() => markNotificationRead(item.id)}
    >
      <Text style={styles.icon}>{item.icon}</Text>
      <View style={styles.content}>
        <View style={styles.titleRow}>
          <Text style={styles.title}>{item.title}</Text>
          <Text style={styles.timestamp}>{item.timestamp}</Text>
        </View>
        <Text style={styles.description}>{item.description}</Text>
      </View>
      {!item.read && <View style={styles.unreadDot} />}
    </TouchableOpacity>
  );

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Notifications</Text>
        {unreadCount > 0 && (
          <View style={styles.unreadBadge}>
            <Text style={styles.unreadBadgeText}>{unreadCount}</Text>
          </View>
        )}
        <TouchableOpacity onPress={onClose} style={styles.closeButton}>
          <Text style={styles.closeText}>✕</Text>
        </TouchableOpacity>
      </View>

      <FlatList
        data={mockNotifications}
        keyExtractor={item => item.id.toString()}
        renderItem={renderNotification}
        contentContainerStyle={styles.listContent}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyIcon}>🔔</Text>
            <Text style={styles.emptyText}>No notifications yet</Text>
          </View>
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0F1C',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 20,
    paddingTop: 48,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255,255,255,0.06)',
  },
  headerTitle: {
    color: '#FFFFFF',
    fontSize: 22,
    fontWeight: '700',
    flex: 1,
  },
  unreadBadge: {
    backgroundColor: '#5B8CFF',
    borderRadius: 12,
    minWidth: 24,
    height: 24,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 6,
    marginRight: 8,
  },
  unreadBadgeText: {
    color: '#FFFFFF',
    fontSize: 12,
    fontWeight: '700',
  },
  closeButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: 'rgba(255,255,255,0.05)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  closeText: {
    color: '#A8B2D1',
    fontSize: 14,
  },
  listContent: {
    paddingBottom: 40,
  },
  item: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255,255,255,0.04)',
  },
  itemUnread: {
    backgroundColor: 'rgba(91, 140, 255, 0.05)',
  },
  icon: {
    fontSize: 24,
    marginRight: 14,
    width: 32,
    textAlign: 'center',
  },
  content: {
    flex: 1,
  },
  titleRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  title: {
    color: '#FFFFFF',
    fontSize: 15,
    fontWeight: '600',
    flex: 1,
  },
  timestamp: {
    color: '#A8B2D1',
    fontSize: 11,
    marginLeft: 12,
  },
  description: {
    color: '#A8B2D1',
    fontSize: 13,
    lineHeight: 18,
  },
  unreadDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#5B8CFF',
    marginLeft: 8,
    marginTop: 6,
  },
  emptyContainer: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 80,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: 16,
  },
  emptyText: {
    color: '#A8B2D1',
    fontSize: 16,
  },
});
