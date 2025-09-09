import { useEffect, useRef, useMemo } from "react";
//mermaid dynamically wants dagre-d3-es at runtime so we need to import it here,
//otherwise tree shaking will remove it and mermaid will fail to render
import * as dagre from 'dagre-d3-es';
import type { MermaidConfig } from "mermaid";
import mermaid from "mermaid";

console.debug('dagre-d3-es loaded:', !!dagre);

type Props = {
  chart: string;               // Mermaid DSL
  config?: MermaidConfig;      // optional overrides
  className?: string;
};

export default function MermaidChart({ chart, config, className }: Props) {
  const ref = useRef<HTMLDivElement>(null);
  const configJson = useMemo(() => JSON.stringify(config), [config]);


  useEffect(() => {
    mermaid.initialize({
      startOnLoad: false,
      theme: 'default',
      ...config,
      securityLevel: 'strict',
    });
  
    const el = ref.current;
    if (!el) return;
  
    mermaid.parse(chart)
      .then(() => {
        // If parse succeeds, render
        const id = 'm-' + Math.random().toString(36).slice(2);
        mermaid.render(id, chart)
          .then(({ svg, bindFunctions }) => {
            el.innerHTML = svg;
            bindFunctions?.(el);
          })
          .catch((err) => {
            console.log("err", err);
            el.innerHTML = `<pre><span style="color:red">Cannot display diagram</span>\n\n${chart}</pre>`;
          });
      })
      .catch((err) => {
        console.log("err", err);
        el.innerHTML = `<pre><span style="color:red">Cannot display diagram</span>\n\n${chart}</pre>`;
      });
  }, [chart, configJson]);
  

  return <div className={className} ref={ref} style={{ width: '100%', overflow: 'auto' }} />;
}
