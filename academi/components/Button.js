import React from 'react';
import { TouchableOpacity, Text, StyleSheet, View } from 'react-native';
import { darkTheme } from '../themes/theme';

// Variants: primary, secondary, ghost, icon
export default function Button({ variant = 'primary', label, onPress, icon, disabled = false }) {
  const theme = darkTheme; // For now using dark theme as default
  
  const getButtonStyles = () => {
    switch (variant) {
      case 'primary':
        return {
          container: styles.primaryContainer,
          text: styles.primaryText
        };
      case 'secondary':
        return {
          container: styles.secondaryContainer,
          text: styles.secondaryText
        };
      case 'ghost':
        return {
          container: styles.ghostContainer,
          text: styles.ghostText
        };
      case 'icon':
        return {
          container: styles.iconContainer,
          text: styles.iconText
        };
      default:
        return {
          container: styles.primaryContainer,
          text: styles.primaryText
        };
    }
  };

  const styles = StyleSheet.create({
    // Base styles
    baseContainer: {
      padding: 12,
      borderRadius: 16,
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'center',
    },
    baseText: {
      fontSize: 16,
      fontWeight: '600', // SemiBold
    },
    
    // Primary variant
    primaryContainer: {
      backgroundColor: theme.colors.primary,
    },
    primaryText: {
      color: '#FFFFFF',
    },
    
    // Secondary variant
    secondaryContainer: {
      backgroundColor: theme.colors.bgSecondary,
      borderWidth: 1,
      borderColor: theme.colors.borderSubtle,
    },
    secondaryText: {
      color: theme.colors.textPrimary,
    },
    
    // Ghost variant
    ghostContainer: {
      backgroundColor: 'transparent',
    },
    ghostText: {
      color: theme.colors.textPrimary,
    },
    
    // Icon variant
    iconContainer: {
      padding: 12,
      borderRadius: 16,
    },
    iconText: {
      color: theme.colors.textPrimary,
    },
    
    // Disabled styles
    disabledContainer: {
      opacity: 0.5,
    },
    disabledText: {
      opacity: 0.5,
    },
  });

  const { container, text } = getButtonStyles();
  
  return (
    <TouchableOpacity
      style={[styles.baseContainer, container, disabled && styles.disabledContainer]}
      onPress={onPress}
      disabled={disabled}
    >
      {icon && <View style={{ marginRight: 8 }}>{icon}</View>}
      <Text style={[styles.baseText, text, disabled && styles.disabledText]}>
        {label}
      </Text>
    </TouchableOpacity>
  );
}