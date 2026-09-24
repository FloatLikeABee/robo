import VisualMarkdown from './lib/VisualMarkdown';

const sqlIntroStrip = /Here's the SQL query based on your request:\n\n/g;

export default function ChatMarkdown({ text }) {
  const cleaned = (text || '').replace(sqlIntroStrip, '');
  return (
    <div className="response-text chat-markdown">
      <VisualMarkdown text={cleaned} />
    </div>
  );
}
