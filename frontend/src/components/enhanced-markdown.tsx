import React, { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus, vs } from 'react-syntax-highlighter/dist/esm/styles/prism';

import { Button } from '@/components/ui/button';
import { Copy, Check } from 'lucide-react';
import { cn } from '@/lib/utils';

interface EnhancedMarkdownProps {
  children: string;
  className?: string;
  showCopyMessage?: boolean;
}

interface CopyButtonProps {
  text: string;
  size?: 'sm' | 'default' | 'lg' | 'icon';
  className?: string;
  iconOnly?: boolean;
}

const CopyButton: React.FC<CopyButtonProps> = ({
  text,
  size = 'icon',
  className,
  iconOnly = false,
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy text: ', err);
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = text;
      document.body.appendChild(textArea);
      textArea.select();
      try {
        document.execCommand('copy');
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } catch (fallbackErr) {
        console.error('Fallback copy failed: ', fallbackErr);
      }
      document.body.removeChild(textArea);
    }
  };

  return (
    <Button
      variant="ghost"
      size={size}
      onClick={handleCopy}
      className={cn(
        "transition-all duration-200 relative",
        size === 'icon' ? "h-6 w-6 p-1" : "h-8 w-8 p-1.5",
        copied
          ? "text-green-600 bg-green-50 dark:bg-green-900/20"
          : "text-gray-500 hover:text-gray-700 hover:bg-gray-100 dark:hover:bg-gray-700",
        className
      )}
      title={copied ? "Copied!" : "Copy to clipboard"}
    >
      <div
        className={cn(
          "transition-all duration-200",
          copied ? "scale-110" : "scale-100"
        )}
      >
        {copied ? (
          <Check
            className={cn(
              size === 'icon' ? "h-3 w-3" : "h-4 w-4",
              "animate-pulse"
            )}
          />
        ) : (
          <Copy className={size === 'icon' ? "h-3 w-3" : "h-4 w-4"} />
        )}
      </div>
      {!iconOnly && copied && (
        <span className="ml-1 text-xs animate-pulse">Copied!</span>
      )}
    </Button>
  );
};

const CodeBlock: React.FC<any> = ({
  node,
  inline,
  className,
  children,
  ...props
}) => {
  const [showOverlay, setShowOverlay] = useState(false);
  const [isDark, setIsDark] = useState(false);
  
  // Detect dark mode
  React.useEffect(() => {
    const checkDarkMode = () => {
      const isDarkMode = document.documentElement.classList.contains('dark') || 
                        window.matchMedia('(prefers-color-scheme: dark)').matches;
      setIsDark(isDarkMode);
    };
    
    checkDarkMode();
    
    // Watch for theme changes
    const observer = new MutationObserver(checkDarkMode);
    observer.observe(document.documentElement, { 
      attributes: true, 
      attributeFilter: ['class'] 
    });
    
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', checkDarkMode);
    
    return () => {
      observer.disconnect();
      mediaQuery.removeEventListener('change', checkDarkMode);
    };
  }, []);
  
  const match = /language-(\w+)/.exec(className || '');
  const rawLanguage = match ? match[1] : '';
  
  // Map common language aliases to proper syntax highlighter names
  const languageMap: { [key: string]: string } = {
    'js': 'javascript',
    'ts': 'typescript',
    'jsx': 'javascript',
    'tsx': 'typescript',
    'py': 'python',
    'rb': 'ruby',
    'sh': 'bash',
    'yml': 'yaml',
    'md': 'markdown',
    'golang': 'go',
  };
  
  const language = languageMap[rawLanguage] || rawLanguage || 'text';
  const codeContent = String(children).replace(/\n$/, '');
  
  // Choose theme based on dark mode
  const syntaxTheme = isDark ? vscDarkPlus : vs;

  if (inline) {
    return (
      <code
        className="relative bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono border"
        {...props}
      >
        {children}
      </code>
    );
  }

  return (
    <>
      <div className="relative group my-4 border border-gray-200 dark:border-gray-700 rounded-md overflow-hidden shadow-sm">
        <div className="flex items-center justify-between bg-gray-100 dark:bg-gray-800 px-3 py-2 border-b border-gray-200 dark:border-gray-700">
          <span className="text-xs font-medium text-gray-600 dark:text-gray-400 uppercase tracking-wide">
            {rawLanguage || language || 'Code'}
          </span>
          <div className="flex items-center space-x-2">
            <CopyButton
              text={codeContent}
              className="opacity-0 group-hover:opacity-100 transition-all duration-200 transform translate-x-1 group-hover:translate-x-0"
            />
            <button
              onClick={() => setShowOverlay(true)}
              className="opacity-0 group-hover:opacity-100 transition-all duration-200 text-xs text-blue-600 hover:underline"
              aria-label="View full code"
            >
              View
            </button>
          </div>
        </div>
        <SyntaxHighlighter
          language={language}
          style={syntaxTheme}
          PreTag="div"
          className="!p-0 !m-0 overflow-x-auto text-sm leading-relaxed"
          showLineNumbers={false}
          wrapLines={true}
          wrapLongLines={true}
          customStyle={{
            margin: 0,
            padding: '1rem',
            backgroundColor: isDark ? '#1e1e1e' : '#f8f8f8',
            fontSize: '14px',
            lineHeight: '1.5',
          }}
        >
          {codeContent}
        </SyntaxHighlighter>
      </div>

      {showOverlay && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-70 p-8"
          role="dialog"
          aria-modal="true"
        >
          <div className="bg-white dark:bg-gray-900 p-6 rounded-lg max-w-3xl w-full relative">
            <button
              onClick={() => setShowOverlay(false)}
              className="absolute top-2 right-2 text-gray-800 dark:text-gray-200 text-2xl font-bold"
              aria-label="Close"
            >
              ×
            </button>
            <CopyButton text={codeContent} size="default" iconOnly={false} className="mb-4" />
            <SyntaxHighlighter
              language={language}
              style={syntaxTheme}
              PreTag="div"
              className="overflow-x-auto max-h-[70vh] text-sm leading-normal !rounded-md"
              showLineNumbers={true}
              wrapLines={true}
              wrapLongLines={true}
              customStyle={{
                fontSize: '14px',
                lineHeight: '1.5',
                borderRadius: '6px',
                backgroundColor: isDark ? '#1e1e1e' : '#f8f8f8',
              }}
            >
              {codeContent}
            </SyntaxHighlighter>
          </div>
        </div>
      )}
    </>
  );
};

// Optionally you can keep your paragraph component
const ParagraphWithInlineCode: React.FC<any> = ({ children, ...props }) => {
  return (
    <p className="mb-4 leading-relaxed" {...props}>
      {children}
    </p>
  );
};

export const EnhancedMarkdown: React.FC<EnhancedMarkdownProps> = ({
  children,
  className,
  showCopyMessage = false,
}) => {
  // Optional: function to strip markdown to plain text for full message copy button
  const stripMarkdown = (markdown: string): string => {
    return markdown
      .replace(/``````/g, (match) => {
        const lines = match.split('\n');
        return lines.slice(1, -1).join('\n');
      })
      .replace(/`([^`]+)`/g, '$1')
      .replace(/^#{1,6}\s+/gm, '')
      .replace(/\*\*([^\*]+)\*\*/g, '$1')
      .replace(/\*([^\*]+)\*/g, '$1')
      .replace(/__([^_]+)__/g, '$1')
      .replace(/_([^_]+)_/g, '$1')
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
      .replace(/^\s*[-*+]\s+/gm, '')
      .replace(/^\s*\d+\.\s+/gm, '')
      .replace(/^>\s+/gm, '')
      .replace(/\n\s*\n/g, '\n')
      .trim();
  };

  const plainText = stripMarkdown(children);

  return (
    <div className={cn('relative group', className)}>
      {/* Copy entire message button */}
      {showCopyMessage && (
        <div className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-all duration-200 transform translate-x-1 group-hover:translate-x-0 z-10">
          <div className="bg-white/95 dark:bg-gray-800/95 backdrop-blur-sm rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-1">
            <CopyButton
              text={plainText}
              size="sm"
              className="hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
            />
          </div>
        </div>
      )}

      <div className="prose prose-sm max-w-none dark:prose-invert">
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          components={{
            code: CodeBlock,
            p: ParagraphWithInlineCode,
            h1: ({ children }) => (
              <h1 className="text-2xl font-bold mb-4 mt-6">{children}</h1>
            ),
            h2: ({ children }) => (
              <h2 className="text-xl font-bold mb-3 mt-5">{children}</h2>
            ),
            h3: ({ children }) => (
              <h3 className="text-lg font-bold mb-2 mt-4">{children}</h3>
            ),
            h4: ({ children }) => (
              <h4 className="text-base font-bold mb-2 mt-3">{children}</h4>
            ),
            ul: ({ children }) => (
              <ul className="list-disc list-inside mb-4 space-y-1">{children}</ul>
            ),
            ol: ({ children }) => (
              <ol className="list-decimal list-inside mb-4 space-y-1">{children}</ol>
            ),
            li: ({ children }) => <li className="ml-4">{children}</li>,
            blockquote: ({ children }) => (
              <blockquote className="border-l-4 border-gray-300 dark:border-gray-600 pl-4 italic my-4 text-gray-700 dark:text-gray-300">
                {children}
              </blockquote>
            ),
            table: ({ children }) => (
              <div className="overflow-x-auto my-4">
                <table className="min-w-full border border-gray-200 dark:border-gray-700">
                  {children}
                </table>
              </div>
            ),
            th: ({ children }) => (
              <th className="border border-gray-200 dark:border-gray-700 px-3 py-2 bg-gray-50 dark:bg-gray-800 font-medium text-left">
                {children}
              </th>
            ),
            td: ({ children }) => (
              <td className="border border-gray-200 dark:border-gray-700 px-3 py-2">{children}</td>
            ),
          }}
        >
          {children}
        </ReactMarkdown>
      </div>
    </div>
  );
};
