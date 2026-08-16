/** Guided interview answers used to simulate the user's own email voice. */
export type VoiceProfile = {
  display_name: string;
  role_context: string;
  greeting: string;
  sign_off: string;
  tone: string;
  length: string;
  formality: string;
  common_phrases: string;
  avoid: string;
  delay_style: string;
  sample_emails: string;
};

export const emptyVoiceProfile = (): VoiceProfile => ({
  display_name: '',
  role_context: '',
  greeting: '',
  sign_off: '',
  tone: '',
  length: '',
  formality: '',
  common_phrases: '',
  avoid: '',
  delay_style: '',
  sample_emails: '',
});

export type VoiceQuestion = {
  key: keyof VoiceProfile;
  title: string;
  hint: string;
  placeholder: string;
  multiline?: boolean;
  choices?: string[];
};

export const VOICE_QUESTIONS: VoiceQuestion[] = [
  {
    key: 'display_name',
    title: 'What name should replies use for you?',
    hint: 'First name, full name, or how you usually sign emails.',
    placeholder: 'e.g. Alex Chen',
  },
  {
    key: 'role_context',
    title: 'What should the AI know about your role?',
    hint: 'Job title, team, or context that shapes how you write.',
    placeholder: 'e.g. Product manager at Morph; I coordinate schedules with schools',
    multiline: true,
  },
  {
    key: 'greeting',
    title: 'How do you usually open an email?',
    hint: 'Pick the closest option, or type your own.',
    placeholder: 'Your greeting style…',
    choices: ['Hi {name},', 'Hello {name},', 'Hey {name},', 'Jump straight in — no greeting', 'Match whatever the sender used'],
  },
  {
    key: 'sign_off',
    title: 'How do you usually sign off?',
    hint: 'Closing line + name, or “none”.',
    placeholder: 'e.g. Thanks,\nAlex',
    multiline: true,
    choices: ['Thanks,\n{name}', 'Best,\n{name}', 'Cheers,\n{name}', 'Just my name', 'No sign-off'],
  },
  {
    key: 'tone',
    title: 'What tone should replies sound like?',
    hint: 'This is the main voice signal for the AI.',
    placeholder: 'Describe your tone…',
    choices: ['Warm and friendly', 'Professional but approachable', 'Casual and conversational', 'Direct and terse', 'Formal / corporate'],
  },
  {
    key: 'length',
    title: 'How long are your typical replies?',
    hint: 'Keeps the AI from writing essays (or being too curt).',
    placeholder: 'Reply length preference…',
    choices: ['1–3 short sentences', 'One short paragraph', 'Thorough when needed, otherwise brief', 'Detailed and complete'],
  },
  {
    key: 'formality',
    title: 'How formal are you with names and wording?',
    hint: 'Affects greetings and word choice.',
    placeholder: 'Formality preference…',
    choices: ['First names; relaxed wording', 'Match the other person’s formality', 'Stay fairly formal', 'Very formal'],
  },
  {
    key: 'common_phrases',
    title: 'Phrases you often use',
    hint: 'List a few natural lines so the AI can sound like you.',
    placeholder: 'e.g. “Sounds good — I’ll confirm by EOD.” / “Happy to help.”',
    multiline: true,
  },
  {
    key: 'avoid',
    title: 'What should replies never do?',
    hint: 'Words, habits, or styles that don’t sound like you.',
    placeholder: 'e.g. Don’t say “As an AI…”; avoid slang; don’t over-apologize',
    multiline: true,
  },
  {
    key: 'delay_style',
    title: 'If you need time before answering fully, how do you reply?',
    hint: 'Helps when the inbox item needs a holding response.',
    placeholder: 'e.g. “Got it — I’ll check and get back to you tomorrow.”',
    multiline: true,
  },
  {
    key: 'sample_emails',
    title: 'Paste 1–2 real emails you’ve written (optional)',
    hint: 'Best signal for tone. Redact anything sensitive.',
    placeholder: 'Paste sample email text here…',
    multiline: true,
  },
];

export function parseVoiceProfile(raw?: string | null): VoiceProfile {
  const base = emptyVoiceProfile();
  if (!raw || !String(raw).trim()) return base;
  try {
    const parsed = JSON.parse(String(raw)) as Partial<VoiceProfile>;
    return { ...base, ...parsed };
  } catch {
    return base;
  }
}

export function serializeVoiceProfile(profile: VoiceProfile): string {
  return JSON.stringify(profile);
}

export function voiceProfileProgress(profile: VoiceProfile): { answered: number; total: number } {
  const total = VOICE_QUESTIONS.length;
  let answered = 0;
  for (const q of VOICE_QUESTIONS) {
    if (String(profile[q.key] ?? '').trim()) answered += 1;
  }
  return { answered, total };
}

/** Compile questionnaire answers into the reply style guide used by the agent. */
export function compileVoiceStyleGuide(profile: VoiceProfile): string {
  const name = profile.display_name.trim() || 'the mailbox owner';
  const lines: string[] = [
    `Write every reply as ${name} — in the first person, as if they typed it themselves.`,
    'Do not mention AI, automation, or that you are an agent.',
  ];

  if (profile.role_context.trim()) {
    lines.push(`Role / context: ${profile.role_context.trim()}`);
  }
  if (profile.tone.trim()) {
    lines.push(`Tone: ${profile.tone.trim()}`);
  }
  if (profile.length.trim()) {
    lines.push(`Length: ${profile.length.trim()}`);
  }
  if (profile.formality.trim()) {
    lines.push(`Formality: ${profile.formality.trim()}`);
  }
  if (profile.greeting.trim()) {
    lines.push(`Greeting style: ${profile.greeting.trim().replaceAll('{name}', 'the recipient’s first name when known')}`);
  }
  if (profile.sign_off.trim()) {
    lines.push(
      `Sign-off: ${profile.sign_off.trim().replaceAll('{name}', name)}`,
    );
  }
  if (profile.common_phrases.trim()) {
    lines.push(`Phrases they naturally use:\n${profile.common_phrases.trim()}`);
  }
  if (profile.avoid.trim()) {
    lines.push(`Never do / never say:\n${profile.avoid.trim()}`);
  }
  if (profile.delay_style.trim()) {
    lines.push(`When they need more time before a full answer:\n${profile.delay_style.trim()}`);
  }
  if (profile.sample_emails.trim()) {
    lines.push(`Example emails written by them (imitate this voice closely):\n${profile.sample_emails.trim()}`);
  }

  return lines.join('\n\n');
}
