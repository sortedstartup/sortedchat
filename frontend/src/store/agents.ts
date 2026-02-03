import { atom, map, onMount } from "nanostores";
import { toast } from "sonner";
import {
    Agent,
    AgentChatRequest,
    AgentChatResponse,
    AgentMessage,
    AgentServiceClient,
    CreateAgentRequest,
    UpdateAgentRequest,
    CreateSessionRequest,
    GetAgentMessagesRequest,
    GetAgentsRequest,
    GetSessionsRequest,
    Session,
    ContentEvent,
    ToolCall,
    ToolResult,
    MCPServer,
} from "../../proto/chatservice";
import { createAuthenticatedClientOptions } from "../lib/auth";
import { getUIConfig } from "../lib/config";
import type { ClientReadableStream } from "grpc-web";

// Stream event types matching the new proto structure
export interface StreamEvent {
    type: 'text' | 'thinking' | 'tool_call' | 'tool_result' | 'error' | 'image' | 'video' | 'audio' | 'write_file';
    timestamp: number;
    // For content events (text, thinking, image, etc.)
    text?: string;
    phase?: string;
    model?: string;
    url?: string;
    mimeType?: string;
    // For tool_call events
    toolCallId?: string;
    toolName?: string;
    argumentsJson?: string;
    // For tool_result events
    resultJson?: string;
    success?: boolean;
    errorMessage?: string;
    durationMs?: number;
    // For write_file events
    fileName?: string;
    filePath?: string;
    fileContent?: string;
    fileSize?: number;
    // For structured error events
    errorType?: number;
    errorCode?: number;
}

// Streaming state per session
export interface SessionStreamingState {
    isStreaming: boolean;
    message: string;
    events: StreamEvent[];
    stream: ClientReadableStream<AgentChatResponse> | null;
}

let _agentClient: AgentServiceClient | undefined;

export function getAgentClient(): AgentServiceClient {
    if (!_agentClient) {
        const config = getUIConfig();
        if (!config) {
            throw new Error("UI config not loaded, cannot initialize chat client.");
        }

        _agentClient = new AgentServiceClient(
            config.API_URL,
            {},
            createAuthenticatedClientOptions()
        );
    }
    return _agentClient;
}

// Stores
export const $agents = atom<Agent[]>([]);
export const $currentAgentId = atom<string>("");

export const $sessions = map<Record<string, Session[]>>({});
export const $currentSessionId = atom<string>("");

export const $agentMessages = atom<{
    data: Array<AgentMessage & { streamEvents?: StreamEvent[] }>;
    loading: boolean;
    error: string | null;
}>({
    data: [],
    loading: false,
    error: null,
});

// Map of streaming states keyed by sessionId to support concurrent chats
export const $streamingStates = map<Record<string, SessionStreamingState>>({});

// Helper to update streaming state for a session
function updateSessionStreaming(sessionId: string, updates: Partial<SessionStreamingState>) {
    const currentState = $streamingStates.get()[sessionId] || {
        isStreaming: false,
        message: "",
        events: [],
        stream: null
    };
    $streamingStates.setKey(sessionId, { ...currentState, ...updates });
}

// Actions

export const getAgents = async () => {
    try {
        const response = await getAgentClient().GetAgents(
            GetAgentsRequest.fromObject({}),
            {}
        );
        $agents.set(response.agents || []);
    } catch (error) {
        console.error("Failed to fetch agents:", error);
        toast.error("Failed to fetch agents");
    }
};

export const createAgent = async (name: string, description: string, systemPrompt: string, model: string, provider: string, mcpServers?: MCPServer[]) => {
    try {
        // Properly serialize MCP servers
        const serializedMcpServers = (mcpServers || []).map(server => server.toObject());
        
        const response = await getAgentClient().CreateAgent(
            CreateAgentRequest.fromObject({
                name,
                description,
                system_prompt: systemPrompt,
                model,
                provider,
                local_tools: [],
                mcp_servers: serializedMcpServers
            }),
            {}
        );
        toast.success("Agent created successfully");
        await getAgents();
        return response.agent_id;
    } catch (error) {
        console.error("Failed to create agent:", error);
        toast.error("Failed to create agent");
        throw error;
    }
};

export const updateAgent = async (agentId: string, name: string, description: string, systemPrompt: string, model: string, provider: string, mcpServers?: MCPServer[]) => {
    try {
        // Properly serialize MCP servers
        const serializedMcpServers = (mcpServers || []).map(server => server.toObject());
        
        const response = await getAgentClient().UpdateAgent(
            UpdateAgentRequest.fromObject({
                agent_id: agentId,
                name,
                description,
                system_prompt: systemPrompt,
                model,
                provider,
                mcp_servers: serializedMcpServers
            }),
            {}
        );
        toast.success("Agent updated successfully");
        await getAgents();
        return response.message;
    } catch (error) {
        console.error("Failed to update agent:", error);
        toast.error("Failed to update agent");
        throw error;
    }
};

export const getSessions = async (agentId: string) => {
    if (!agentId) return;
    try {
        const response = await getAgentClient().GetSessions(
            GetSessionsRequest.fromObject({
                agent_id: agentId
            }),
            {}
        );
        $sessions.setKey(agentId, response.sessions || []);
    } catch (error) {
        $sessions.setKey(agentId, []);
    }
};

export const createSession = async (agentId: string) => {
    try {
        const response = await getAgentClient().CreateSession(
            CreateSessionRequest.fromObject({
                agent_id: agentId
            }),
            {}
        );
        toast.success("Session created");
        await getSessions(agentId);
        return response.session_id;
    } catch (error) {
        console.error("Failed to create session:", error);
        toast.error("Failed to create session");
        throw error;
    }
};

export const getAgentMessages = async (sessionId: string) => {
    if (!sessionId) return;

    $agentMessages.set({
        data: [],
        loading: true,
        error: null
    });

    try {
        const response = await getAgentClient().GetAgentMessages(
            GetAgentMessagesRequest.fromObject({
                session_id: sessionId
            }),
            {}
        );
        $agentMessages.set({
            data: response.messages || [],
            loading: false,
            error: null
        });
    } catch (error) {
        console.error("Failed to fetch messages:", error);
        $agentMessages.set({
            data: [],
            loading: false,
            error: "Failed to fetch messages"
        });
    }
};

export const sendAgentMessage = async (sessionId: string, message: string) => {
    if (!sessionId || !message.trim()) return;

    // 1. Optimistic update if viewing the same session
    if ($currentSessionId.get() === sessionId) {
        const userMsg = new AgentMessage();
        userMsg.role = "user";
        userMsg.content = message;
        userMsg.type = "text";
        userMsg.id = "temp-" + Date.now();

        $agentMessages.set({
            ...$agentMessages.get(),
            data: [...$agentMessages.get().data, userMsg]
        });
    }

    // 2. Initialize streaming state for this session
    updateSessionStreaming(sessionId, {
        isStreaming: true,
        message: "",
        events: []
    });

    try {
        const stream = getAgentClient().AgentChat(
            AgentChatRequest.fromObject({
                session_id: sessionId,
                message: message
            }),
            {}
        );

        updateSessionStreaming(sessionId, { stream });

        let accumulatedContent = "";
        const accumulatedEvents: StreamEvent[] = [];

        stream.on("data", (res: AgentChatResponse) => {
            if (res.has_content) {
                const content = res.content;
                if (content.type === 'text') {
                    accumulatedContent += content.text;
                    updateSessionStreaming(sessionId, { message: accumulatedContent });
                }
                accumulatedEvents.push({
                    type: content.type as StreamEvent['type'],
                    timestamp: Date.now(),
                    text: content.text,
                    phase: content.phase,
                    model: content.model,
                    url: content.url,
                    mimeType: content.mime_type,
                });
            } else if (res.has_tool_call) {
                accumulatedEvents.push({
                    type: 'tool_call',
                    timestamp: Date.now(),
                    toolCallId: res.tool_call.id,
                    toolName: res.tool_call.name,
                    argumentsJson: res.tool_call.arguments_json,
                });
            } else if (res.has_tool_result) {
                accumulatedEvents.push({
                    type: 'tool_result',
                    timestamp: Date.now(),
                    toolCallId: res.tool_result.id,
                    toolName: res.tool_result.name,
                    resultJson: res.tool_result.result_json,
                    success: res.tool_result.success,
                    errorMessage: res.tool_result.error_message,
                    durationMs: Number(res.tool_result.duration_ms),
                });
            } else if (res.has_error) {
                accumulatedEvents.push({
                    type: 'error',
                    timestamp: Date.now(),
                    text: res.error.message,
                    errorType: res.error.type,
                    errorCode: res.error.code,
                });
            }
            updateSessionStreaming(sessionId, { events: [...accumulatedEvents] });
        });

        stream.on("end", () => {
            // Commit to messages if active
            if ($currentSessionId.get() === sessionId) {
                const finalAssistantMsg = {
                    id: "temp-resp-" + Date.now(),
                    session_id: sessionId,
                    sequence_number: 0,
                    role: "assistant",
                    content: accumulatedContent,
                    type: "text",
                    streamEvents: [...accumulatedEvents]
                };

                $agentMessages.set({
                    ...$agentMessages.get(),
                    data: [...$agentMessages.get().data, finalAssistantMsg],
                    loading: false,
                    error: null
                });
            }
            
            // Cleanup streaming state
            updateSessionStreaming(sessionId, {
                isStreaming: false,
                message: "",
                events: [],
                stream: null
            });
        });

        stream.on("error", (err) => {
            console.error(`Stream error for ${sessionId}:`, err);
            updateSessionStreaming(sessionId, { isStreaming: false, stream: null });
            if ($currentSessionId.get() === sessionId) {
                toast.error("Error in agent chat");
            }
        });

    } catch (error) {
        console.error("Failed to initiate chat:", error);
        updateSessionStreaming(sessionId, { isStreaming: false });
        if ($currentSessionId.get() === sessionId) {
            toast.error("Failed to send message");
        }
    }
};

// Listener for session changes
$currentSessionId.listen((sessionId) => {
    if (sessionId) {
        getAgentMessages(sessionId);
    } else {
        $agentMessages.set({ data: [], loading: false, error: null });
    }
});

// Initial load
onMount($agents, () => {
    getAgents();
});
