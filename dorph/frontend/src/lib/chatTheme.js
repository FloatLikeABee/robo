export const CHAT_THEME_KEY = 'skool-ai-chat-theme';
export const CHAT_THEME_DEFAULT = 'twilight';

/** Single product preset; migrates legacy dark/light keys once. */
export function resolveChatTheme() {
  try {
    const stored = localStorage.getItem(CHAT_THEME_KEY);
    if (stored === CHAT_THEME_DEFAULT) {
      return CHAT_THEME_DEFAULT;
    }
    if (stored === 'dark' || stored === 'light' || !stored) {
      localStorage.setItem(CHAT_THEME_KEY, CHAT_THEME_DEFAULT);
      return CHAT_THEME_DEFAULT;
    }
    return stored;
  } catch {
    return CHAT_THEME_DEFAULT;
  }
}

/** MUI admin palette mode (twilight maps to dark). */
export function chatThemeToAdminMode(chatTheme) {
  return chatTheme === 'light' ? 'light' : 'dark';
}
