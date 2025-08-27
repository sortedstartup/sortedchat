import React, { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Pre, highlight } from 'codehike/code';
import { Copy, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface EnhancedMarkdownProps {
  children: string;
  showCopyMessage?: boolean;
  preferInlineCode?: boolean; // New prop to prefer inline code display
}

// Custom code component using Code Hike's direct components
const CodeComponent = ({ 
  inline,
  className, 
  children, 
  showCopyMessage,
  preferInlineCode,
  ...props 
}: { 
  inline?: boolean;
  className?: string; 
  children?: React.ReactNode; 
  showCopyMessage?: boolean;
  preferInlineCode?: boolean;
  [key: string]: any;
}) => {
  const [copiedStates, setCopiedStates] = useState<{ [key: string]: boolean }>({});
  const [highlightedCode, setHighlightedCode] = useState<any>(null);
  
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

  const codeString = String(children || '').replace(/\n$/, '');
  const codeId = `code-${Math.random().toString(36).substr(2, 9)}`;
  
  // Extract language from className
  const match = /language-(\w+)/.exec(className || '');
  const language = match ? match[1] : 'text';

  // Check if this should be displayed inline
  const shouldDisplayInline = inline || (preferInlineCode && codeString.split('\n').length <= 3);
  
  // Enhanced inline code for short snippets
  if (shouldDisplayInline) {
    const isMultiLine = codeString.includes('\n');
    
    if (isMultiLine && codeString.split('\n').length <= 3) {
      // Multi-line inline code (up to 3 lines)
      return (
        <div className="inline-block relative group mx-1">
          <div className="bg-gray-100 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 font-mono text-sm">
            {language && language !== 'text' && (
              <div className="text-xs text-blue-600 dark:text-blue-400 font-semibold mb-2">
                {language}
              </div>
            )}
            <pre className="text-rose-600 dark:text-rose-400 m-0 whitespace-pre-wrap">
              {codeString}
            </pre>
          </div>
          {showCopyMessage && (
            <Button
              size="sm"
              variant="ghost"
              className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity h-6 w-6 p-0 bg-gray-200/80 hover:bg-gray-300/80 dark:bg-gray-700/80 dark:hover:bg-gray-600/80 rounded"
              onClick={() => copyToClipboard(codeString, codeId)}
            >
              {copiedStates[codeId] ? (
                <Check className="h-3 w-3 text-green-600 dark:text-green-400" />
              ) : (
                <Copy className="h-3 w-3 text-gray-600 dark:text-gray-300" />
              )}
            </Button>
          )}
        </div>
      );
    }
    
    // Single-line inline code
    return (
      <span className="relative group">
        <code className="bg-gray-100 dark:bg-gray-800 text-rose-600 dark:text-rose-400 px-2 py-1 mx-1 rounded text-sm font-mono border border-gray-300 dark:border-gray-600" {...props}>
          {language && language !== 'text' && (
            <span className="text-xs text-blue-600 dark:text-blue-400 font-semibold mr-2">
              {language}:
            </span>
          )}
          {children}
        </code>
        {showCopyMessage && (
          <Button
            size="sm"
            variant="ghost"
            className="absolute -top-1 -right-1 opacity-0 group-hover:opacity-100 transition-opacity h-5 w-5 p-0 bg-gray-200/90 hover:bg-gray-300/90 dark:bg-gray-700/90 dark:hover:bg-gray-600/90 rounded"
            onClick={() => copyToClipboard(codeString, codeId)}
          >
            {copiedStates[codeId] ? (
              <Check className="h-3 w-3 text-green-600 dark:text-green-400" />
            ) : (
              <Copy className="h-3 w-3 text-gray-600 dark:text-gray-300" />
            )}
          </Button>
        )}
      </span>
    );
  }

  // Use Code Hike's highlight function for syntax highlighting (block code)
  React.useEffect(() => {
    const highlightCode = async () => {
      try {
        const highlighted = await highlight(
          { value: codeString, lang: language, meta: '' },
          'github-dark-dimmed'
        );
        setHighlightedCode(highlighted);
      } catch (err) {
        console.error('Code highlighting error:', err);
        // Fallback to plain code
        setHighlightedCode({ code: codeString, style: {} });
      }
    };

    highlightCode();
  }, [codeString, language]);

  if (!highlightedCode) {
    return (
      <div className="bg-gray-900 text-gray-100 p-4 rounded-lg border border-gray-700 my-4">
        <pre className="font-mono text-sm overflow-x-auto">
          <code>{codeString}</code>
        </pre>
      </div>
    );
  }

  return (
    <div className="relative group my-2">
      {language !== 'text' && (
        <div className="absolute top-3 left-3 z-20 text-xs bg-blue-600 text-white px-2 py-1 rounded font-mono shadow-sm">
          {language}
        </div>
      )}
      
      {showCopyMessage && (
        <Button
          size="sm"
          variant="ghost"
          className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity z-20 h-7 w-7 p-0 bg-gray-800/80 hover:bg-gray-700/80 rounded"
          onClick={() => copyToClipboard(codeString, codeId)}
        >
          {copiedStates[codeId] ? (
            <Check className="h-3.5 w-3.5 text-green-400" />
          ) : (
            <Copy className="h-3.5 w-3.5 text-gray-300" />
          )}
        </Button>
      )}
      
      {/* Use Code Hike's Pre component with tighter spacing */}
      <Pre 
        code={highlightedCode} 
        style={highlightedCode.style} 
        className="pt-8 pb-3 px-3 rounded-lg border border-gray-700 shadow-sm"
      />
    </div>
  );
};

export function EnhancedMarkdown({ 
  children, 
  showCopyMessage = false, 
  preferInlineCode = false 
}: EnhancedMarkdownProps) {
  return (
    <div className="prose prose-sm max-w-none dark:prose-invert">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          code: (props) => (
            <CodeComponent 
              {...props} 
              showCopyMessage={showCopyMessage} 
              preferInlineCode={preferInlineCode}
            />
          ),
          p: (props) => (
            <p className="mb-1 last:mb-0 leading-relaxed" {...props} />
          ),
          ul: (props) => (
            <ul className="mb-2 mt-1 ml-4 list-disc space-y-0" {...props} />
          ),
          ol: (props) => (
            <ol className="mb-2 mt-1 ml-4 list-decimal space-y-0" {...props} />
          ),
          li: (props) => (
            <li className="leading-normal mb-0" {...props} />
          ),
          h1: (props) => (
            <h1 className="text-xl font-bold mb-1 mt-3 first:mt-0" {...props} />
          ),
          h2: (props) => (
            <h2 className="text-lg font-bold mb-1 mt-3 first:mt-0" {...props} />
          ),
          h3: (props) => (
            <h3 className="text-base font-bold mb-1 mt-2 first:mt-0" {...props} />
          ),
          blockquote: (props) => (
            <blockquote className="border-l-4 border-blue-500 bg-blue-50/50 dark:bg-blue-950/50 pl-4 py-2 my-2 rounded-r italic" {...props} />
          ),
          a: (props) => (
            <a className="text-blue-600 hover:text-blue-800 underline" target="_blank" rel="noopener noreferrer" {...props} />
          ),
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
}