import React, { useState, useRef, useEffect } from 'react';
import { View, TextInput, StyleSheet, Animated, Dimensions, Platform } from 'react-native';

const { width } = Dimensions.get('window');

export default function NeuralSearchBar({
  placeholder = 'Search...',
  value,
  onChangeText,
  onSubmit,
  onFocus,
  onBlur,
  autoFocus = false
}) {
  const [isFocused, setIsFocused] = useState(false);
  const borderAnim = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    if (isFocused) {
      Animated.loop(
        Animated.sequence([
          Animated.timing(borderAnim, {
            toValue: 1,
            duration: 1000,
            useNativeDriver: false,
          }),
          Animated.timing(borderAnim, {
            toValue: 0,
            duration: 1000,
            useNativeDriver: false,
          }),
        ])
      ).start();
    } else {
      borderAnim.setValue(0);
    }
  }, [isFocused]);

  const handleFocus = () => {
    setIsFocused(true);
    onFocus && onFocus();
  };

  const handleBlur = () => {
    setIsFocused(false);
    onBlur && onBlur();
  };

  const handleSubmit = (text) => {
    onSubmit && onSubmit(text);
    handleBlur();
  };

  return (
    <View style={styles.container}>
      <View style={[styles.searchWrapper, isFocused && styles.searchWrapperFocused]}>
        <View style={styles.searchIcon}>
          <Text style={styles.searchIconText}>🔍</Text>
        </View>
        <TextInput
          style={styles.input}
          placeholder={placeholder}
          value={value}
          onChangeText={onChangeText}
          onSubmitEditing={(e) => handleSubmit(e.nativeEvent.text)}
          placeholderTextColor="rgba(168, 178, 209, 0.5)"
          autoFocus={autoFocus}
          onFocus={handleFocus}
          onBlur={handleBlur}
          returnKeyType="search"
        />
      </View>
      {isFocused && (
        <Animated.View
          style={[
            styles.glowBorder,
            {
              opacity: borderAnim.interpolate({
                inputRange: [0, 1],
                outputRange: [0.3, 0.8],
              }),
            },
          ]}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    marginHorizontal: 16,
    marginVertical: 12,
  },
  searchWrapper: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(255,255,255,0.05)',
    borderRadius: 24,
    paddingHorizontal: 20,
    height: 48,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.08)',
  },
  searchWrapperFocused: {
    borderColor: 'rgba(91, 140, 255, 0.5)',
    backgroundColor: 'rgba(255,255,255,0.08)',
  },
  searchIcon: {
    marginRight: 10,
  },
  searchIconText: {
    fontSize: 16,
  },
  input: {
    flex: 1,
    color: '#FFFFFF',
    fontSize: 16,
    height: '100%',
  },
  glowBorder: {
    position: 'absolute',
    top: -2,
    left: -2,
    right: -2,
    bottom: -2,
    borderRadius: 26,
    backgroundColor: 'transparent',
    borderWidth: 2,
    borderColor: '#5B8CFF',
    shadowColor: '#5B8CFF',
    shadowOffset: { width: 0, height: 0 },
    shadowOpacity: 0.5,
    shadowRadius: 10,
    elevation: 5,
  },
});
