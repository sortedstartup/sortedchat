
import React from 'react';
import ReactMarkdown from 'react-markdown';
import { Copy, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';

// Define proper types for ReactMarkdown component props
interface CodeProps {
  node?: any;
  inline?: boolean;
  className?: string;
  children?: React.ReactNode;
  [key: string]: any;
}

interface EnhancedMarkdownProps {
  children: string;
  showCopyMessage?: boolean;
}

export function EnhancedMarkdown({ children, showCopyMessage = false }: EnhancedMarkdownProps) {
  const [copiedStates, setCopiedStates] = React.useState<{ [key: string]: boolean }>({});

  const copyToClipboard = async (text: string, id: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedStates(prev => ({ ...prev, [id]: true }));
      setTimeout(() => {
        setCopiedStates(prev => ({ ...prev, [id]: false }));
      }, 2000);
    } catch (err) {
      console.error('Failed to copy text: ', err);
    }
  };

  return (
    <div className="prose prose-sm max-w-none dark:prose-invert">
      <ReactMarkdown
        components={{
          code: ({ node, inline, className, children, ...props }: CodeProps) => {
            const codeString = String(children).replace(/\n$/, '');
            const codeId = `code-${Math.random().toString(36).substr(2, 9)}`;

            // Treat short single-line code as inline, even if it came from ```
            const isShortSingleLine = codeString.length < 50 && !codeString.includes('\n');
            const shouldBeInline = inline || isShortSingleLine;

            // Code block detection (only for longer or multi-line code)
            if (!shouldBeInline && codeString) {
              return (
                <div className="relative group my-0">
                  {showCopyMessage && (
                    <Button
                      size="sm"
                      variant="ghost"
                      className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity z-10 h-8 w-8 p-0"
                      onClick={() => copyToClipboard(codeString, codeId)}
                    >
                      {copiedStates[codeId] ? (
                        <Check className="h-4 w-4 text-green-600" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  )}
                  <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm leading-tight border">
                    <code className="font-mono" {...props}>
                      {children}
                    </code>
                  </pre>
                </div>
              );
            }

            // Inline code (for `code` or short ```code```)
            return (
              <code
                className="bg-gray-100 dark:bg-gray-800 text-red-600 dark:text-red-400 px-1.5 py-0.5 rounded text-sm font-mono border"
                {...props}
              >
                {children}
              </code>
            );
          },
          // Clean styling for other elements with proper typing
          p: ({ children, ...props }) => (
            <p className="mb-0 last:mb-0 leading-relaxed" {...props}>
              {children}
            </p>
          ),
          ul: ({ children, ...props }) => (
            <ul className="mb-0 ml-4 list-disc" {...props}>
              {children}
            </ul>
          ),
          ol: ({ children, ...props }) => (
            <ol className="mb-0 ml-4 list-decimal" {...props}>
              {children}
            </ol>
          ),
          li: ({ children, ...props }) => (
            <li className="leading-relaxed" {...props}>
              {children}
            </li>
          ),
          h1: ({ children, ...props }) => (
            <h1 className="text-xl font-bold mb-0 mt-0 first:mt-0 text-gray-900 dark:text-gray-100" {...props}>
              {children}
            </h1>
          ),
          h2: ({ children, ...props }) => (
            <h2 className="text-lg font-bold mb-0 mt-0 first:mt-0 text-gray-900 dark:text-gray-100" {...props}>
              {children}
            </h2>
          ),
          h3: ({ children, ...props }) => (
            <h3 className="text-base font-bold mb-0 mt-0 first:mt-0 text-gray-900 dark:text-gray-100" {...props}>
              {children}
            </h3>
          ),
          blockquote: ({ children, ...props }) => (
            <blockquote className="border-l-4 border-blue-500 bg-blue-50 dark:bg-blue-950 pl-4 py-2 my-0 rounded-r" {...props}>
              {children}
            </blockquote>
          ),
          strong: ({ children, ...props }) => (
            <strong className="font-semibold text-gray-900 dark:text-gray-100" {...props}>
              {children}
            </strong>
          ),
          table: ({ children, ...props }) => (
            <div className="overflow-x-auto my-1">
              <table className="min-w-full border border-gray-300 dark:border-gray-600" {...props}>
                {children}
              </table>
            </div>
          ),
          th: ({ children, ...props }) => (
            <th className="border border-gray-300 dark:border-gray-600 px-3 py-2 bg-gray-100 dark:bg-gray-800 font-semibold text-left" {...props}>
              {children}
            </th>
          ),
          td: ({ children, ...props }) => (
            <td className="border border-gray-300 dark:border-gray-600 px-3 py-2" {...props}>
              {children}
            </td>
          ),
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
}