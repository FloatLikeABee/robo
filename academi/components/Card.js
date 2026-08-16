import React from 'react';
import { View, Text, StyleSheet, Image, TouchableOpacity } from 'react-native';
import { darkTheme } from '../themes/theme';

// Variants: community, doc, guide
export default function Card({ 
  variant = 'community', 
  title, 
  content, 
  imageUri, 
  tags = [], 
  onPress,
  showActions = true
}) {
  const theme = darkTheme;

  return (
    <TouchableOpacity 
      style={[styles.container, styles[variant]]} 
      onPress={onPress}
      activeOpacity={0.8}
    >
      {imageUri && (
        <Image 
          source={{ uri: imageUri }} 
          style={styles.image} 
          resizeMode="cover"
        />
      )}
      <View style={styles.content}>
        {title && <Text style={styles.title}>{title}</Text>}
        {content && <Text style={styles.body}>{content}</Text>}
        {tags.length > 0 && (
          <View style={styles.tagsContainer}>
            {tags.map((tag, index) => (
              <Text key={index} style={styles.tag}>
                {tag}
              </Text>
            ))}
          </View>
        )}
        {showActions && (
          <View style={styles.actionsContainer}>
            <View style={styles.actionButton}>
              <Text style={styles.actionText}>Save</Text>
            </View>
            <View style={styles.actionButton}>
              <Text style={styles.actionText}>Share</Text>
            </View>
          </View>
        )}
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    borderRadius: 16,
    overflow: 'hidden',
    backgroundColor: darkTheme.colors.surfaceGlass,
    borderWidth: 1,
    borderColor: darkTheme.colors.borderSubtle,
    margin: 8,
  },
  // Variant-specific styles
  community: {},
  doc: {},
  guide: {},
  
  image: {
    height: 120,
    width: '100%',
  },
  content: {
    padding: 16,
  },
  title: {
    color: darkTheme.colors.textPrimary,
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 8,
  },
  body: {
    color: darkTheme.colors.textSecondary,
    fontSize: 14,
    marginBottom: 12,
    lineHeight: 20,
  },
  tagsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginBottom: 12,
  },
  tag: {
    backgroundColor: 'rgba(91, 140, 255, 0.2)',
    borderRadius: 12,
    paddingHorizontal: 10,
    paddingVertical: 4,
    marginRight: 8,
    marginBottom: 8,
    fontSize: 12,
    color: darkTheme.colors.primary,
  },
  actionsContainer: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 8,
  },
  actionButton: {
    backgroundColor: darkTheme.colors.bgSecondary,
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 12,
  },
  actionText: {
    color: darkTheme.colors.textPrimary,
    fontSize: 12,
    fontWeight: '500',
  },
});