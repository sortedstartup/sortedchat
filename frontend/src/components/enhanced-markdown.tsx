import { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Copy, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import MermaidChart from '@/components/mermaid-chart';

interface EnhancedMarkdownProps {
  children: string;
}

const CodeComponent = ({ inline, className, children }: any) => {
  const [copied, setCopied] = useState(false);

  const codeString = String(children || '');
  const match = /language-(\w+)/.exec(className || '');
  const language = match ? match[1] : 'text';

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(codeString);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy text:', err);
    }
  };

  const isInline = inline || codeString.length < 40;
  if (isInline) {
    return (
      <code className="bg-gray-100 dark:bg-gray-800 text-rose-600 dark:text-rose-400 px-2 py-1 rounded text-sm font-mono border">
        {children}
      </code>
    );
  }

  return (
    <div className="relative group my-2">
      {language !== 'text' && (
        <div className="absolute top-2 left-3 z-20 text-xs bg-blue-600 text-white px-2 py-1 rounded-md font-mono">
          {language}
        </div>
      )}

      <Button
        size="sm"
        variant="ghost"
        className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity z-20 h-8 w-8 p-0 bg-gray-800/80 hover:bg-gray-700/80"
        onClick={copyToClipboard}
      >
        {copied ? <Check className="h-4 w-4 text-green-400" /> : <Copy className="h-4 w-4 text-gray-300" />}
      </Button>

      <SyntaxHighlighter
        style={oneDark}
        language={language}
        PreTag="div"
        className="text-sm leading-tight border border-gray-700 rounded-lg"
        customStyle={{
          margin: 0,
          paddingTop: language !== 'text' ? '2rem' : '1rem',
          paddingBottom: '1rem',
          paddingLeft: '1rem',
          paddingRight: '1rem',
        }}
      >
        {codeString}
      </SyntaxHighlighter>
    </div>
  );
};

export function EnhancedMarkdown({ children }: EnhancedMarkdownProps) {
  return (
    <div className="prose prose-sm max-w-none dark:prose-invert">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          code: ({ className, children, ...props }) => {
            const codeString = String(children || '');
            const match = /language-(\w+)/.exec(className || '');
            console.log("language: ", children);
            const language = match ? match[1] : 'text';

            if ( language === 'mermaid') {
              return <MermaidChart chart={codeString} className="my-4 flex justify-center" />;
            }

            return <CodeComponent className={className} children={children} {...props} />;
          },
          p: (props) => <p className="mb-1 last:mb-0 " {...props} />,
          ul: (props) => <ul className="mb-1 ml-8 list-disc space-y-0" {...props} />,
          ol: (props) => <ol className="mb-1 ml-8 list-decimal space-y-0" {...props} />,
          li: (props) => <li className="mb-1 leading-normal pl-2" {...props} />,
          h1: (props) => <h1 className="text-xl font-bold mb-0 mt-0 first:mt-0" {...props} />,
          h2: (props) => <h2 className="text-lg font-bold mb-0 mt-0 first:mt-0" {...props} />,
          h3: (props) => <h3 className="text-base font-bold mb-0 mt-0 first:mt-0" {...props} />,
          blockquote: (props) => (
            <blockquote className="border-l-4 border-blue-500 bg-blue-50/50 dark:bg-blue-950/50 pl-4 py-2 my-0 rounded-r italic" {...props} />
          ),
          a: (props) => (
            <a className="text-blue-600 hover:text-blue-800 underline" target="_blank" rel="noopener noreferrer" {...props} />
          ),
          table: (props) => (
            <div className="overflow-x-auto my-3">
              <table className="border-collapse border border-gray-700 dark:border-gray-500 w-full text-sm">
                {props.children}
              </table>
            </div>
          ),
          thead: (props) => (
            <thead className="bg-gray-200 dark:bg-gray-800">{props.children}</thead>
          ),
          th: (props) => (
            <th className="border border-gray-700 dark:border-gray-500 px-3 py-1 text-left font-semibold">
              {props.children}
            </th>
          ),
          tbody: (props) => <tbody>{props.children}</tbody>,
          td: (props) => (
            <td className="border border-gray-700 dark:border-gray-500 px-3 py-1">
              {props.children}
            </td>
          ),
          tr: (props) => (
            <tr className="odd:bg-gray-100 even:bg-gray-50 dark:odd:bg-gray-900 dark:even:bg-gray-800">
              {props.children}
            </tr>
          ),
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
}