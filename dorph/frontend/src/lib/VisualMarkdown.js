import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import MermaidBlock from './MermaidBlock';
import PixelGridBlock from './PixelGridBlock';
import { EnlargeImg } from './visualLightbox';

function fenceLang(className) {
  const m = /language-([^\s]+)/.exec(className || '');
  return (m && m[1] ? m[1] : '').toLowerCase();
}

function CodeBlock({ className, children, ...props }) {
  const lang = fenceLang(className);
  const text = String(children || '').replace(/\n$/, '');
  if (lang === 'mermaid') return <MermaidBlock source={text} />;
  if (lang === 'pixel') return <PixelGridBlock source={text} />;
  return (
    <code className={className} {...props}>
      {children}
    </code>
  );
}

export default function VisualMarkdown({ text }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        a: ({ href, children, ...props }) => (
          <a href={href} target="_blank" rel="noopener noreferrer" {...props}>
            {children}
          </a>
        ),
        img: ({ src, alt }) => <EnlargeImg src={src} alt={alt || ''} />,
        code: CodeBlock,
      }}
    >
      {text || ''}
    </ReactMarkdown>
  );
}
