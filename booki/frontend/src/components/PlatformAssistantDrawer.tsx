import { PlatformChatDrawer } from '@robo/platform-chat/react';
import '@robo/platform-chat/chat-drawer.css';
import { apiUrl } from '../api/client';
import { useAuth } from '../store/auth';

const SUGGESTIONS = [
  'P&L this month',
  'Trial balance as of today',
  'Analyze my Flow Log spending this month',
  'Analyze account 1000 last month',
  'Calculate (12500-8300)/12500',
  'List customers',
];

type Props = {
  open: boolean;
  onClose: () => void;
};

export function PlatformAssistantDrawer({ open, onClose }: Props) {
  const accessToken = useAuth((s) => s.accessToken);

  return (
    <PlatformChatDrawer
      open={open}
      onClose={onClose}
      title="Morph Booki AI"
      chatEndpoint={apiUrl('/api/v1/assistant/chat')}
      getHeaders={() => {
        const headers: Record<string, string> = {};
        if (accessToken) headers.Authorization = `Bearer ${accessToken}`;
        return headers;
      }}
      welcomeMessage="Hi! I am your **Morph Booki AI assistant**. I can help with ledger analytics, Flow Log money notes, customers, bookings, and calculations."
      suggestions={SUGGESTIONS}
      progressContext={{ app: 'booki' }}
    />
  );
}
