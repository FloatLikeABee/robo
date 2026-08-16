// Theme based on design.md
export const darkTheme = {
  colors: {
    bgPrimary: '#0A0F1C',
    bgSecondary: '#12182A',
    surfaceGlass: 'rgba(255,255,255,0.05)',
    borderSubtle: 'rgba(255,255,255,0.08)',
    textPrimary: '#FFFFFF',
    textSecondary: '#A8B2D1',
    // Accent gradient (we'll use the primary color for simplicity, but note it's a gradient in design)
    primary: '#5B8CFF', // Start of gradient
    secondary: '#9B6DFF', // Middle of gradient
    tertiary: '#00D4FF', // End of gradient
  },
  // We can add more tokens like radii, spacing, etc. as needed
  radii: {
    xs: 8,
    sm: 12,
    md: 16,
    lg: 20,
    xl: 24,
  },
  spacing: {
    '1': 4,
    '2': 8,
    '3': 12,
    '4': 16,
    '5': 24,
    '6': 32,
  },
};

export const lightTheme = {
  colors: {
    bgPrimary: '#F7F9FC',
    bgSecondary: '#FFFFFF',
    surfaceGlass: 'rgba(0,0,0,0.05)', // Adjust for light mode if needed
    borderSubtle: 'rgba(0,0,0,0.08)',
    textPrimary: '#0A0F1C',
    textSecondary: '#5C6A8A',
    primary: '#5B8CFF',
    secondary: '#9B6DFF',
    tertiary: '#00D4FF',
  },
  radii: {
    xs: 8,
    sm: 12,
    md: 16,
    lg: 20,
    xl: 24,
  },
  spacing: {
    '1': 4,
    '2': 8,
    '3': 12,
    '4': 16,
    '5': 24,
    '6': 32,
  },
};

// We'll export a function to get the theme based on mode
export const getTheme = (mode) => (mode === 'dark' ? darkTheme : lightTheme);