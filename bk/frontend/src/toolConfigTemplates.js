/**
 * Help text + example JSON for Tool Manager configuration dialog.
 * Keys match backend tool ids (see src/tools.py register_tool first argument).
 */

export const TOOL_CONFIG_HELP = {
  default:
    'Most tools use an empty object {}. Only tools that need credentials or API settings require JSON keys here.',
  web_search: 'No settings required. Use {} unless you add custom keys later.',
  wikipedia: 'No settings required.',
  calculator: 'No settings required.',
  financial: 'No API keys required (uses yfinance). Leave {} unless you extend the backend.',
  equalizer: 'Uses the default LLM pipeline. No JSON fields required for basic use.',
  crawler:
    'Optional crawler defaults (if supported by your backend version). Often {}. Example includes placeholder URL and collection naming.',
  email:
    'SMTP credentials for sending mail. Use app passwords for Gmail/Outlook. Ports: 587 (TLS) or 465 (SSL).',
  document_reader: 'Optional timeouts or provider overrides. Usually {}.',
  youtube_summarizer: 'No settings required.',
  academic_search: 'No settings required.',
  mind_map: 'No settings required.',
  debate_analyzer: 'No settings required.',
  first_principles: 'No settings required.',
  image_generator: 'No settings required for Pollinations defaults.',
  story_generator: 'No settings required.',
  task_planner: 'No settings required.',
  multi_agent: 'No settings required.',
  browser_automation: 'Optional: headless mode, viewport, or storage state paths if your backend supports them.',
  custom: 'Depends on your custom tool implementation. Ask the author which keys are required.',
};

export const TOOL_CONFIG_TEMPLATE = {
  email: {
    smtp_server: 'smtp.example.com',
    smtp_port: 587,
    smtp_username: 'your_user',
    smtp_password: 'your_app_password',
  },
  crawler: {
    default_url: 'https://example.com',
    collection_name: 'my_crawl',
    max_pages: 20,
  },
  browser_automation: {
    headless: true,
    default_timeout_ms: 30000,
  },
};

export function getToolConfigHelp(toolId) {
  return TOOL_CONFIG_HELP[toolId] || TOOL_CONFIG_HELP.default;
}

export function getToolConfigTemplateObject(toolId) {
  if (TOOL_CONFIG_TEMPLATE[toolId]) return { ...TOOL_CONFIG_TEMPLATE[toolId] };
  return {};
}
