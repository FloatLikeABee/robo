import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

const sqlIntroStrip = /Here's the SQL query based on your request:\n\n/g;

export default function ChatMarkdown({ text }) {
  const cleaned = (text || '').replace(sqlIntroStrip, '');
  return (
    <div className="response-text chat-markdown">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          a: ({ href, children, ...props }) => (
            <a href={href} target="_blank" rel="noopener noreferrer" {...props}>
              {children}
            </a>
          ),
        }}
      >
        {cleaned}
      </ReactMarkdown>
    </div>
  );
}
