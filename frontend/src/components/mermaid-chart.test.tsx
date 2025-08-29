/*
  Tests for MermaidChart component
  Testing framework: Vitest
  Testing library: @testing-library/react (+ @testing-library/jest-dom/vitest)
*/
import React from 'react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import '@testing-library/jest-dom/vitest';
import { render, act, cleanup } from '@testing-library/react';
import MermaidChart from './mermaid-chart';

// Prepare controllable mermaid mocks used across tests
const initialize = vi.fn();
const renderMermaid = vi.fn();

// Module mock: default export object with initialize/render plus named re-exports.
vi.mock('mermaid', () => ({
  default: { initialize, render: renderMermaid },
  initialize,
  render: renderMermaid,
}));

const nextTick = () => new Promise((r) => setTimeout(r, 0));

beforeEach(() => {
  initialize.mockReset();
  renderMermaid.mockReset();
  vi.restoreAllMocks(); // reset any previous spies
  vi.spyOn(Math, 'random').mockReturnValue(0.123456789); // deterministic id
});

afterEach(() => {
  cleanup();
});

describe('MermaidChart', () => {
  const chart = 'graph TD; A-->B;';

  it('initializes mermaid with safe defaults and merges overrides', () => {
    render(<MermaidChart chart={chart} config={{ theme: 'forest', securityLevel: 'loose' as any }} />);
    expect(initialize).toHaveBeenCalledTimes(1);
    expect(initialize).toHaveBeenCalledWith(
      expect.objectContaining({
        startOnLoad: false,
        securityLevel: 'loose',
        theme: 'forest',
      })
    );
  });

  it('renders SVG and binds functions on successful render', async () => {
    const bind = vi.fn();
    renderMermaid.mockResolvedValue({ svg: '<svg id="ok"></svg>', bindFunctions: bind });

    const { container } = render(<MermaidChart chart={chart} className="host" />);
    const host = container.querySelector('div') as HTMLDivElement;

    await act(async () => {
      await nextTick();
    });

    // ID is derived from Math.random(); verify argument and DOM updates
    const expectedId = 'm-' + (0.123456789).toString(36).slice(2);
    expect(renderMermaid).toHaveBeenCalledWith(expectedId, chart);
    expect(host.innerHTML).toContain('<svg');
    expect(bind).toHaveBeenCalledWith(host);
    // Inline style from component
    expect(host.style.width).toBe('100%');
    expect(host.style.overflow).toBe('auto');
  });

  it('shows a red error <pre> when mermaid.render rejects', async () => {
    renderMermaid.mockRejectedValueOnce(new Error('boom'));

    const { container } = render(<MermaidChart chart={chart} />);
    const host = container.querySelector('div') as HTMLDivElement;

    await act(async () => {
      await nextTick();
    });

    const pre = host.querySelector('pre');
    expect(pre).not.toBeNull();
    expect(pre!).toHaveTextContent('Mermaid error:');
    expect(pre!).toHaveTextContent('boom');
    expect(pre!.getAttribute('style') || '').toMatch(/color:\s*red/);
  });

  it('re-renders when chart changes', async () => {
    renderMermaid.mockResolvedValue({ svg: '<svg></svg>' });

    const { rerender } = render(<MermaidChart chart={'graph TD; A-->B;'} />);
    await act(async () => { await nextTick(); });
    expect(renderMermaid).toHaveBeenCalledTimes(1);

    rerender(<MermaidChart chart={'graph TD; A-->C;'} />);
    await act(async () => { await nextTick(); });
    expect(renderMermaid).toHaveBeenCalledTimes(2);
  });

  it('does not re-render if config identity changes but JSON content is the same', async () => {
    renderMermaid.mockResolvedValue({ svg: '<svg></svg>' });

    const cfg1 = { theme: 'dark' as const };
    const cfg2 = { theme: 'dark' as const }; // same JSON
    const { rerender } = render(<MermaidChart chart={chart} config={cfg1} />);
    await act(async () => { await nextTick(); });
    expect(renderMermaid).toHaveBeenCalledTimes(1);

    rerender(<MermaidChart chart={chart} config={cfg2} />);
    await act(async () => { await nextTick(); });
    expect(renderMermaid).toHaveBeenCalledTimes(1); // unchanged
  });

  it('re-renders and re-initializes when config content changes', async () => {
    renderMermaid.mockResolvedValue({ svg: '<svg></svg>' });

    const { rerender } = render(<MermaidChart chart={chart} config={{ theme: 'default' }} />);
    await act(async () => { await nextTick(); });
    expect(initialize).toHaveBeenCalledTimes(1);

    rerender(<MermaidChart chart={chart} config={{ theme: 'forest' }} />);
    await act(async () => { await nextTick(); });
    expect(initialize).toHaveBeenCalledTimes(2);
    const lastInitArg = initialize.mock.calls.at(-1)![0];
    expect(lastInitArg.theme).toBe('forest');
  });
});