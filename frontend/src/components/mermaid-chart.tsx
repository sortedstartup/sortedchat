import { useEffect, useRef } from 'react';
import mermaid, { type MermaidConfig } from 'mermaid';
import { useMemo } from 'react';

type Props = {
  chart: string;               // Mermaid DSL
  config?: MermaidConfig;      // optional overrides
  className?: string;
};

export default function MermaidChart({ chart, config, className }: Props) {
  const ref = useRef<HTMLDivElement>(null);
  const configJson = useMemo(() => JSON.stringify(config), [config]);


  useEffect(() => {
    // Initialize once per render with safe defaults
    mermaid.initialize({
      startOnLoad: false,
      securityLevel: 'strict',  // safer; disables raw HTML in labels
      theme: 'default',
      ...config,
    });

    const el = ref.current;
    if (!el) return;

    // Render fresh each time
    const id = 'm-' + Math.random().toString(36).slice(2);
    mermaid
      .render(id, chart)
      .then(({ svg, bindFunctions }) => {
        el.innerHTML = svg;
        bindFunctions?.(el);
      })
      .catch((err) => {
        el.innerHTML = `<pre style="color:red">Mermaid error:\n${String(err)}</pre>`;
      });
  }, [chart, configJson]);

  return <div className={className} ref={ref} style={{ width: '100%', overflow: 'auto' }} />;
}
