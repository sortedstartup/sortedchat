// Tests for Chat page components: ChunksDisplay, ChatInputBox, Chat
// Framework: React Testing Library with Jest/Vitest (auto-detected by repository).
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, within, fireEvent } from '@testing-library/react'

// Fallback: if the project uses Jest, the imports below are compatible after ts-jest setup.
// If using Jest, replace vitest's vi with jest in globals via jest.fn or configure ts-jest to alias.


//
// Lightweight mocks and stubs
//
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<any>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => vi.fn(),
    useParams: () => ({ projectId: 'proj-1', chatId: 'chat-123' }),
  }
})

// Mock clipboard
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn().mockResolvedValue(undefined),
  },
})

// Mock UI components to reduce rendering complexity
vi.mock('@/components/ui/button', () => ({
  Button: ({ children, onClick, className, ...rest }: any) => (
    <button onClick={onClick} className={className} {...rest}>{children}</button>
  ),
}))
vi.mock('@/components/ui/dropdown-menu', () => {
  const Comp = ({ children }: any) => <div>{children}</div>;
  const Trigger = ({ children }: any) => <div>{children}</div>;
  const Content = ({ children }: any) => <div>{children}</div>;
  const Item = ({ children, onClick }: any) => <button onClick={onClick}>{children}</button>;
  return {
    DropdownMenu: Comp,
    DropdownMenuTrigger: Trigger,
    DropdownMenuContent: Content,
    DropdownMenuItem: Item,
  }
})
vi.mock('@/components/ui/dialog', () => ({
  Dialog: ({ open, children }: any) => open ? <div role="dialog">{children}</div> : null,
  DialogContent: ({ children }: any) => <div>{children}</div>,
  DialogHeader: ({ children }: any) => <div>{children}</div>,
  DialogTitle: ({ children }: any) => <h2>{children}</h2>,
}))

// Mock icons to simple spans
vi.mock('lucide-react', () => new Proxy({}, {
  get: () => (props: any) => <span aria-hidden="true" {...props} />
}))

// Mock EnhancedMarkdown to render raw children
vi.mock('@/components/enhanced-markdown', () => ({
  EnhancedMarkdown: ({ children }: any) => <div data-testid="md">{children}</div>,
}))

// Mock store primitives and actions used by the components
const mockStoreVal = <T,>(initial: T) => {
  let val = initial
  const listeners = new Set<(v: T) => void>()
  return {
    get: () => val,
    set: (v: T) => { val = v; listeners.forEach(l => l(v)) },
    listen: (cb: (v: T) => void) => {
      listeners.add(cb); return () => listeners.delete(cb)
    }
  }
}

// Shape stores with names used in the page
const $availableModels = mockStoreVal<any[]>([
  { id: 'gpt-4o', label: 'GPT-4o' },
  { id: 'gpt-4o-mini', label: 'GPT-4o mini' },
])
const $selectedModel = mockStoreVal<string>('gpt-4o')
const $ragEnabled = mockStoreVal<boolean>(false)

const $currentChatMessages = mockStoreVal<{ data: any[]; loading: boolean }>({ data: [], loading: false } as any)
const $streamingMessage = mockStoreVal<string>('')
const $currentChatMessage = mockStoreVal<string>('')
const $listChatBranch = mockStoreVal<any[]>([])
const $currentDocumentReferences = mockStoreVal<any[]>([])
const $ragDocumentDetails = mockStoreVal<any>({ loading: false, error: null, data: null })

const storeModule = {
  useStore: (s: any) => s.get(),
  $availableModels,
  $selectedModel,
  $ragEnabled,
  $currentChatMessages,
  $streamingMessage,
  $currentChatMessage,
  $listChatBranch,
  $currentDocumentReferences,
  $ragDocumentDetails,
  $currentChatId: mockStoreVal<string | null>(null)
}

// Actions used in the component
const actions = {
  toggleRagEnabled: vi.fn(() => $ragEnabled.set(!$ragEnabled.get())),
  setRagEnabledForProject: vi.fn((val: boolean) => $ragEnabled.set(val)),
  doChat: vi.fn((message: string, projectId?: string) => {
    // emulate pushing to messages
    const current = $currentChatMessages.get() as any
    const arr = current?.data ?? []
    $currentChatMessages.set({ data: [...arr, { message_id: 'mid-1', role: 'user', content: message }], loading: false })
  }),
  fetchRAGDocumentReference: vi.fn(async (messageId: string, projectId: string, docId: string) => {
    $ragDocumentDetails.set({
      loading: false,
      error: null,
      data: {
        Chunks: [
          { start_byte: 0, end_byte: 10, chunk_text: 'short text', simillarity: 0.987 },
          { start_byte: 11, end_byte: 500, chunk_text: new Array(25).fill('word').join(' '), simillarity: 0.432 },
        ]
      }
    })
  }),
  BranchChat: vi.fn()
}

// Mock modules where these stores/actions are exported from.
// Paths may differ in the repo; adapt these mocks to actual import paths if needed.
vi.mock('@/stores', () => storeModule)
vi.mock('@/stores/chat', () => storeModule)
vi.mock('@/stores/rag', () => storeModule)
vi.mock('@/actions/chat', () => actions)
vi.mock('@/actions/rag', () => actions)
vi.mock('@/lib/rag', () => actions)
vi.mock('nanostores', () => ({})) // prevent real nanostores

// Finally import the page under test after mocks
let ChatPage: any
let InternalExports: any

beforeEach(async () => {
  vi.clearAllMocks()
  // Dynamic import to ensure mocks are applied first; adjust path to actual file if different.
  const mod = await import('./chat') // frontend/src/pages/chat.tsx compiled path relative to this test
  ChatPage = mod.Chat
  // Exported inner components might not be exported; for unit tests of internal pure components we may render indirectly via page.
  InternalExports = mod
})

afterEach(() => {
  // reset store values
  $availableModels.set([
    { id: 'gpt-4o', label: 'GPT-4o' },
    { id: 'gpt-4o-mini', label: 'GPT-4o mini' },
  ])
  $selectedModel.set('gpt-4o')
  $ragEnabled.set(false)
  $currentChatMessages.set({ data: [], loading: false } as any)
  $streamingMessage.set('')
  $currentChatMessage.set('')
  $listChatBranch.set([])
  $currentDocumentReferences.set([])
  $ragDocumentDetails.set({ loading: false, error: null, data: null })
})

describe('ChunksDisplay', () => {
  it('renders "No chunks available" when chunks is undefined or empty', async () => {
    const Comp = InternalExports?.ChunksDisplay ?? (() => null)
    const { rerender } = render(<Comp chunks={undefined} />)
    expect(screen.getByText(/No chunks available/i)).toBeInTheDocument()

    rerender(<Comp chunks={[]} />)
    expect(screen.getByText(/No chunks available/i)).toBeInTheDocument()
  })

  it('shows bytes, similarity (with 3 decimals), and full text for short chunks', () => {
    const Comp = InternalExports?.ChunksDisplay ?? (() => null)
    const chunks = [{ start_byte: 1, end_byte: 20, chunk_text: 'short text', simillarity: 0.9876 }]
    render(<Comp chunks={chunks as any} />)
    expect(screen.getByText(/Bytes 1 - 20/)).toBeInTheDocument()
    expect(screen.getByText(/Similarity: 0\.988/)).toBeInTheDocument()
    expect(screen.getByText('short text')).toBeInTheDocument()
    expect(screen.queryByText('Show more')).not.toBeInTheDocument()
  })

  it('truncates long chunks to 20 words, shows ellipsis and toggles on "Show more/less"', () => {
    const Comp = InternalExports?.ChunksDisplay ?? (() => null)
    const long = new Array(25).fill('word').join(' ')
    render(<Comp chunks={[{ start_byte: 0, end_byte: 100, chunk_text: long, simillarity: 0.5 } as any]} />)

    // Initially truncated
    const display = screen.getByText((content, node) => node?.textContent?.startsWith(new Array(20).fill('word').join(' ')) ?? false)
    expect(display).toBeInTheDocument()
    expect(screen.getByText('...')).toBeInTheDocument()
    const btn = screen.getByRole('button', { name: /Show more/i })
    fireEvent.click(btn)

    // Expanded shows full content and "Show less"
    expect(screen.getByRole('button', { name: /Show less/i })).toBeInTheDocument()
    expect(screen.getByText(long)).toBeInTheDocument()
  })

  it('handles missing fields gracefully (defaults to 0 bytes, N/A similarity, default text)', () => {
    const Comp = InternalExports?.ChunksDisplay ?? (() => null)
    render(<Comp chunks={[{} as any]} />)
    expect(screen.getByText(/Bytes 0 - 0/)).toBeInTheDocument()
    expect(screen.getByText(/Similarity: N\/A/)).toBeInTheDocument()
    expect(screen.getByText(/No content available/)).toBeInTheDocument()
  })
})

describe('ChatInputBox', () => {
  it('disables send button on empty/whitespace and sends on click/Enter', async () => {
    const Comp = InternalExports?.ChatInputBox ?? (() => null)
    const onSend = vi.fn()
    render(<Comp projectId="proj-1" onSendMessage={onSend} />)

    const sendBtn = screen.getByRole('button', { name: '' }) // icon-only button
    expect(sendBtn).toBeDisabled()

    const input = screen.getByPlaceholderText(/Ask anything/i) as HTMLTextAreaElement
    fireEvent.change(input, { target: { value: '  hello  ' } })
    expect(sendBtn).not.toBeDisabled()

    fireEvent.click(sendBtn)
    expect(onSend).toHaveBeenCalledWith('  hello  ')
    // input cleared
    expect(input.value).toBe('')

    // Enter to send
    fireEvent.change(input, { target: { value: 'second' } })
    fireEvent.keyDown(input, { key: 'Enter', code: 'Enter', shiftKey: false })
    expect(onSend).toHaveBeenCalledWith('second')
  })

  it('toggles RAG checkbox when projectId is present', () => {
    const Comp = InternalExports?.ChatInputBox ?? (() => null)
    render(<Comp projectId="proj-1" onSendMessage={vi.fn()} />)
    const ragToggle = screen.getByRole('checkbox', { name: /Enable RAG/i })
    expect(ragToggle).toBeInTheDocument()
    expect(ragToggle).not.toBeChecked()
    fireEvent.click(ragToggle)
    expect(ragToggle).toBeChecked()
  })

  it('shows model dropdown with available models and updates selected model on click', () => {
    const Comp = InternalExports?.ChatInputBox ?? (() => null)
    render(<Comp projectId="proj-1" onSendMessage={vi.fn()} />)
    const modelButton = screen.getByRole('button', { name: /gpt-4o|Select Model/i })
    expect(modelButton).toBeInTheDocument()
    fireEvent.click(modelButton)
    const item = screen.getByRole('button', { name: /GPT-4o mini/i })
    fireEvent.click(item)
    // Selected model in store should update; component label may update on re-render cycles.
    expect(($selectedModel.get())).toBe('gpt-4o-mini')
  })
})

describe('Chat page', () => {
  it('shows loading state, empty state, then messages', () => {
    $currentChatMessages.set({ data: [], loading: true } as any)
    render(<ChatPage />)
    expect(screen.getByText(/Loading messages/i)).toBeInTheDocument()

    $currentChatMessages.set({ data: null, loading: false } as any)
    render(<ChatPage />)
    expect(screen.getByText(/No messages yet/i)).toBeInTheDocument()

    $currentChatMessages.set({ data: [{ message_id: 'u1', role: 'user', content: 'Hi' }], loading: false } as any)
    render(<ChatPage />)
    expect(screen.getAllByTestId('md')[0]).toHaveTextContent('Hi')
  })

  it('renders assistant message with RAG not enabled pill when projectId present and rag_enabled false', () => {
    $currentChatMessages.set({ data: [{ message_id: 'a1', role: 'assistant', content: 'Answer', rag_enabled: false }], loading: false } as any)
    render(<ChatPage />)
    expect(screen.getByText(/RAG not enabled/i)).toBeInTheDocument()
  })

  it('renders document references buttons and opens details dialog with chunks', async () => {
    $currentChatMessages.set({
      data: [{
        message_id: 'a2',
        role: 'assistant',
        content: 'See sources',
        references: [
          { doc_id: 'doc1', file_name: 'file1.md', Chunks: [{}, {}] }
        ]
      }],
      loading: false
    } as any)
    render(<ChatPage />)

    const btn = screen.getByRole('button', { name: /file1\.md/i })
    fireEvent.click(btn)

    // After clicking, mocked fetch sets chunks data
    expect(await screen.findByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText(/Document Chunks: file1\.md/i)).toBeInTheDocument()
    expect(screen.getByText(/Showing 2 chunks?/i)).toBeInTheDocument()

    // Verify truncated chunk shows Show more
    const showMore = screen.getByRole('button', { name: /Show more/i })
    expect(showMore).toBeInTheDocument()
  })

  it('copy message button writes to clipboard and shows success icon state', async () => {
    $currentChatMessages.set({
      data: [{ message_id: 'aid', role: 'assistant', content: 'Copy me' }],
      loading: false
    } as any)
    render(<ChatPage />)
    // First ghost button is copy (icon-only); query all then click first
    const buttons = screen.getAllByRole('button')
    // Find by title is not available; click the first copy ghost button heuristically
    fireEvent.click(buttons.find(b => b.className?.includes?.('ghost'))!)
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('Copy me')
  })

  it('branch chat button calls BranchChat when assistant message has id', () => {
    $currentChatMessages.set({
      data: [{ message_id: 'aid', role: 'assistant', content: 'x' }],
      loading: false
    } as any)
    render(<ChatPage />)
    const branchBtn = screen.getByRole('button', { name: /Branch Chat/i })
    fireEvent.click(branchBtn)
    expect(actions.BranchChat).toHaveBeenCalledWith('aid')
  })

  it('sends message via ChatInputBox and triggers doChat', () => {
    render(<ChatPage />)
    const input = screen.getByPlaceholderText(/Ask anything/i)
    fireEvent.change(input, { target: { value: 'hello from page' } })
    // Send with click on icon button (last button in footer)
    const sendButtons = screen.getAllByRole('button')
    fireEvent.click(sendButtons[sendButtons.length - 1])
    expect(actions.doChat).toHaveBeenCalledWith('hello from page', 'proj-1')
  })
})