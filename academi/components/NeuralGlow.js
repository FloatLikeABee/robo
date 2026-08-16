import React, { useState, useEffect } from 'react';
import { View, StyleSheet, Animated } from 'react-native';

export default function NeuralGlow({ isActive = false, size = 8 }) {
  const glowOpacity = new Animated.Value(0.3);

  useEffect(() => {
    if (isActive) {
      const animate = () => {
        Animated.sequence([
          Animated.timing(glowOpacity, {
            toValue: 1,
            duration: 800,
            useNativeDriver: true,
          }),
          Animated.timing(glowOpacity, {
            toValue: 0.3,
            duration: 800,
            useNativeDriver: true,
          }),
        ]).start(() => animate());
      };
      animate();
    } else {
      Animated.timing(glowOpacity, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }).start();
    }
  }, [isActive]);

  if (!isActive) return null;

  return (
    <View style={styles.container}>
      <Animated.View
        style={[
          styles.glow,
          {
            width: size * 2,
            height: size * 2,
            borderRadius: size,
            opacity: glowOpacity,
            backgroundColor: 'rgba(91, 140, 255, 0.6)',
          },
        ]}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 4,
  },
  glow: {
    shadowColor: '#5B8CFF',
    shadowOffset: { width: 0, height: 0 },
    shadowOpacity: 0.8,
    shadowRadius: 8,
    elevation: 5,
  },
});
