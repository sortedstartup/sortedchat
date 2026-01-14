import React from 'react';
import 'katex/dist/katex.min.css';
import katex from 'katex';

interface KaTeXRendererProps {
  formula: string;
}

const KaTeXRenderer: React.FC<KaTeXRendererProps> = ({ formula }) => {
  let html = '';
  try {
    html = katex.renderToString(formula);
  } catch (error) {
    html = '<span style="color: red;">Invalid formula</span>';
  }

  return (
    <span
      className="katex-container inline-block align-middle my-0"
      style={{
        isolation: 'isolate',
        contain: 'layout style',
      }}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
};

export default KaTeXRenderer;