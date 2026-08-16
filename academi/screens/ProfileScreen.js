import React, { useState } from 'react';
import {
  View, Text, StyleSheet, FlatList, TouchableOpacity,
  Image, Switch, ScrollView
} from 'react-native';
import Button from '../components/Button';
import TagChip from '../components/TagChip';
import useAppStore from '../store/appStore';

const userData = {
  name: "Alex Chen",
  email: "alex@academi.app",
  avatar: "https://via.placeholder.com/100",
  stats: {
    streak: 12,
    contributions: 45,
    savedDocs: 28,
  },
};

const settingsItems = [
  { id: 1, title: "Appearance", subtitle: "Theme, icon, and colors", icon: "🎨" },
  { id: 2, title: "Notifications", subtitle: "Manage alerts and messages", icon: "🔔" },
  { id: 3, title: "Privacy & Security", subtitle: "Account protection and data", icon: "🔒" },
  { id: 4, title: "AI Personalization", subtitle: "Tone, depth, and response style", icon: "🤖" },
  { id: 5, title: "Data & Storage", subtitle: "Manage downloads and cache", icon: "💾" },
  { id: 6, title: "Help & Support", subtitle: "Get assistance and report issues", icon: "❓" },
  { id: 7, title: "About Academi", subtitle: "Version, licenses, and team", icon: "ℹ️" },
];

export default function ProfileScreen() {
  const { isDarkMode, toggleTheme } = useAppStore();

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      <View style={styles.profileHeader}>
        <Image
          source={{ uri: userData.avatar }}
          style={styles.avatar}
        />
        <View style={styles.profileInfo}>
          <Text style={styles.profileName}>{userData.name}</Text>
          <Text style={styles.profileEmail}>{userData.email}</Text>
        </View>
        <TouchableOpacity style={styles.editButton} activeOpacity={0.7}>
          <Text style={styles.editText}>Edit</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.statsContainer}>
        <View style={styles.statCard}>
          <Text style={styles.statIcon}>🔥</Text>
          <Text style={styles.statValue}>{userData.stats.streak}</Text>
          <Text style={styles.statLabel}>Streak</Text>
        </View>
        <View style={styles.statCard}>
          <Text style={styles.statIcon}>💬</Text>
          <Text style={styles.statValue}>{userData.stats.contributions}</Text>
          <Text style={styles.statLabel}>Posts</Text>
        </View>
        <View style={styles.statCard}>
          <Text style={styles.statIcon}>📚</Text>
          <Text style={styles.statValue}>{userData.stats.savedDocs}</Text>
          <Text style={styles.statLabel}>Saved</Text>
        </View>
      </View>

      <View style={styles.themeSection}>
        <View style={styles.themeToggleRow}>
          <View style={styles.themeToggleLeft}>
            <Text style={styles.themeTitle}>Dark Mode</Text>
            <Text style={styles.themeSubtitle}>{isDarkMode ? 'Dark theme active' : 'Light theme active'}</Text>
          </View>
          <Switch
            trackColor={{ false: '#767577', true: '#5B8CFF' }}
            thumbColor={isDarkMode ? '#0A0F1C' : '#FFFFFF'}
            ios_backgroundColor="#3e3e3e"
            value={isDarkMode}
            onValueChange={toggleTheme}
          />
        </View>
      </View>

      <View style={styles.settingsSection}>
        <Text style={styles.sectionTitle}>Settings</Text>
        {settingsItems.map((item) => (
          <TouchableOpacity
            key={item.id}
            style={styles.settingsItem}
            activeOpacity={0.6}
          >
            <View style={styles.settingsItemLeft}>
              <Text style={styles.settingsIcon}>{item.icon}</Text>
              <View style={styles.settingsItemText}>
                <Text style={styles.settingsItemTitle}>{item.title}</Text>
                <Text style={styles.settingsItemSubtitle}>{item.subtitle}</Text>
              </View>
            </View>
            <Text style={styles.settingsChevron}>›</Text>
          </TouchableOpacity>
        ))}
      </View>

      <TouchableOpacity style={styles.signOutContainer} activeOpacity={0.6}>
        <Text style={styles.signOutText}>Sign Out</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0F1C',
  },
  profileHeader: {
    padding: 24,
    paddingTop: 48,
    flexDirection: 'row',
    alignItems: 'center',
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255,255,255,0.08)',
  },
  avatar: {
    width: 64,
    height: 64,
    borderRadius: 32,
    marginRight: 16,
    borderWidth: 2,
    borderColor: 'rgba(91, 140, 255, 0.5)',
  },
  profileInfo: {
    flex: 1,
  },
  profileName: {
    color: '#FFFFFF',
    fontSize: 20,
    fontWeight: '600',
  },
  profileEmail: {
    color: '#A8B2D1',
    fontSize: 14,
  },
  editButton: {
    paddingVertical: 8,
    paddingHorizontal: 16,
    borderRadius: 20,
    backgroundColor: 'rgba(91, 140, 255, 0.15)',
  },
  editText: {
    color: '#5B8CFF',
    fontSize: 14,
    fontWeight: '600',
  },
  statsContainer: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    padding: 24,
  },
  statCard: {
    alignItems: 'center',
    backgroundColor: 'rgba(255,255,255,0.03)',
    borderRadius: 16,
    padding: 20,
    flex: 1,
    marginHorizontal: 4,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.06)',
  },
  statIcon: {
    fontSize: 28,
    marginBottom: 8,
  },
  statValue: {
    color: '#FFFFFF',
    fontSize: 24,
    fontWeight: '700',
    marginBottom: 4,
  },
  statLabel: {
    color: '#A8B2D1',
    fontSize: 12,
    textTransform: 'uppercase',
    letterSpacing: 1,
  },
  themeSection: {
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  themeToggleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: 16,
    backgroundColor: 'rgba(255,255,255,0.03)',
    borderRadius: 16,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.06)',
  },
  themeToggleLeft: {
    flex: 1,
  },
  themeTitle: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  themeSubtitle: {
    color: '#A8B2D1',
    fontSize: 12,
    marginTop: 2,
  },
  settingsSection: {
    marginTop: 16,
  },
  sectionTitle: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
    paddingHorizontal: 20,
    paddingVertical: 12,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  settingsItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 14,
    paddingHorizontal: 20,
  },
  settingsItemLeft: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
  },
  settingsIcon: {
    fontSize: 20,
    marginRight: 16,
    width: 28,
    textAlign: 'center',
  },
  settingsItemText: {
    flex: 1,
  },
  settingsItemTitle: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '500',
  },
  settingsItemSubtitle: {
    color: '#A8B2D1',
    fontSize: 12,
    marginTop: 2,
  },
  settingsChevron: {
    color: 'rgba(255,255,255,0.2)',
    fontSize: 22,
  },
  signOutContainer: {
    alignItems: 'center',
    padding: 32,
    marginTop: 16,
    marginBottom: 40,
  },
  signOutText: {
    color: '#FF6B6B',
    fontSize: 16,
    fontWeight: '600',
  },
});
