import { CHAT_THEME_DEFAULT, CHAT_THEME_KEY, chatThemeToAdminMode, resolveChatTheme } from './chatTheme';

describe('chatTheme', () => {
  beforeEach(() => {
    localStorage.removeItem(CHAT_THEME_KEY);
  });

  test('migrates legacy dark to twilight', () => {
    localStorage.setItem(CHAT_THEME_KEY, 'dark');
    expect(resolveChatTheme()).toBe(CHAT_THEME_DEFAULT);
    expect(localStorage.getItem(CHAT_THEME_KEY)).toBe(CHAT_THEME_DEFAULT);
  });

  test('twilight maps to admin dark mode', () => {
    expect(chatThemeToAdminMode('twilight')).toBe('dark');
    expect(chatThemeToAdminMode('light')).toBe('light');
  });
});
