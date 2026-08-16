import React from 'react';
import { View, StyleSheet } from 'react-native';

export default function TagChip({ label, active = false, variant = 'default' }) {
  const getVariantStyle = () => {
    switch (variant) {
      case 'glow':
        return styles.glowVariant;
      case 'active':
        return styles.activeVariant;
      default:
        return styles.defaultVariant;
    }
  };

  return (
    <View style={[styles.chip, getVariantStyle(), active && styles.activeChip]}>
      <Text style={[styles.label, active && styles.activeLabel]}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  chip: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 20,
    marginRight: 8,
    marginBottom: 8,
  },
  defaultVariant: {
    backgroundColor: 'rgba(255,255,255,0.08)',
  },
  activeVariant: {
    backgroundColor: 'rgba(91, 140, 255, 0.3)',
  },
  glowVariant: {
    backgroundColor: 'rgba(91, 140, 255, 0.15)',
    borderWidth: 1,
    borderColor: 'rgba(91, 140, 255, 0.5)',
  },
  activeChip: {
    transform: [{ scale: 1.05 }],
  },
  label: {
    fontSize: 13,
    fontWeight: '500',
    color: '#A8B2D1',
  },
  activeLabel: {
    color: '#FFFFFF',
    fontWeight: '600',
  },
});