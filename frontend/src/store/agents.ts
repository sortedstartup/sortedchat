import { atom, map, onMount } from "nanostores";
import { toast } from "sonner";
import {
    Agent,
    AgentChatRequest,
    AgentChatResponse,
    AgentMessage,
    AgentServiceClient,
    CreateAgentRequest,
    CreateSessionRequest,
    GetAgentMessagesRequest,
    GetAgentsRequest,
    GetSessionsRequest,
    Session,
    ContentEvent,
    ToolCall,
    ToolResult,
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

export const $isAgentStreaming = atom<boolean>(false);
export const $agentStreamingMessage = atom<string>("");
export const $agentStreamingEvents = atom<StreamEvent[]>([]);
export let agentStream: ClientReadableStream<AgentChatResponse> | null = null;

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

export const createAgent = async (name: string, description: string, systemPrompt: string, model: string, provider: string) => {
    try {
        const response = await getAgentClient().CreateAgent(
            CreateAgentRequest.fromObject({
                name,
                description,
                system_prompt: systemPrompt,
                model,
                provider,
                local_tools: [] // TODO: Add support for tools
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
        // console.error("Failed to fetch sessions:", error);
        // Suppress error for now as backend might return unimplemented
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
        // If we have a sessions list, refresh it
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
            error: "Failed to fetch messages" // (error as Error).message
        });
    }
};

export const sendAgentMessage = async (sessionId: string, message: string) => {
    if (!sessionId || !message.trim()) return;

    const currentMessages = $agentMessages.get().data;
    const userMsg = new AgentMessage();
    userMsg.role = "user";
    userMsg.content = message;
    userMsg.type = "text";
    userMsg.id = "temp-" + Date.now();

    $agentMessages.set({
        ...$agentMessages.get(),
        data: [...currentMessages, userMsg]
    });

    $isAgentStreaming.set(true);
    $agentStreamingMessage.set("");
    $agentStreamingEvents.set([]);

    try {
        agentStream = getAgentClient().AgentChat(
            AgentChatRequest.fromObject({
                session_id: sessionId,
                message: message
            }),
            {}
        );

        let assistantContent = "";
        const events: StreamEvent[] = [];

        agentStream.on("data", (res: AgentChatResponse) => {
            const responseType = res.response;
            console.log("Agent stream data:", responseType, res.toObject());
            
            // Handle ContentEvent (text, thinking, image, video, audio)
            if (res.has_content) {
                const content = res.content;
                const eventType = content.type as StreamEvent['type'];
                
                if (eventType === 'text') {
                    // Accumulate text content
                    assistantContent += content.text;
                    $agentStreamingMessage.set(assistantContent);
                }
                
                // Add event for all content types
                events.push({
                    type: eventType,
                    timestamp: Date.now(),
                    text: content.text,
                    phase: content.phase,
                    model: content.model,
                    url: content.url,
                    mimeType: content.mime_type,
                });
                $agentStreamingEvents.set([...events]);
            }
            
            // Handle ToolCall
            if (res.has_tool_call) {
                const toolCall = res.tool_call;
                events.push({
                    type: 'tool_call',
                    timestamp: Date.now(),
                    toolCallId: toolCall.id,
                    toolName: toolCall.name,
                    argumentsJson: toolCall.arguments_json,
                });
                $agentStreamingEvents.set([...events]);
            }
            
            // Handle ToolResult
            if (res.has_tool_result) {
                const toolResult = res.tool_result;
                events.push({
                    type: 'tool_result',
                    timestamp: Date.now(),
                    toolCallId: toolResult.id,
                    toolName: toolResult.name,
                    resultJson: toolResult.result_json,
                    success: toolResult.success,
                    errorMessage: toolResult.error_message,
                    durationMs: Number(toolResult.duration_ms),
                });
                $agentStreamingEvents.set([...events]);
            }
            
            // Handle Error
            if (res.has_error) {
                events.push({
                    type: 'error',
                    timestamp: Date.now(),
                    text: res.error,
                });
                $agentStreamingEvents.set([...events]);
            }
        });

        agentStream.on("end", () => {
            // Create assistant message object with proper fields
            const msgWithEvents = {
                id: "temp-resp-" + Date.now(),
                session_id: sessionId,
                sequence_number: 0,
                role: "assistant",
                content: assistantContent,
                type: "text",
                tool_name: "",
                tool_call_id: "",
                tool_args: "",
                streamEvents: events
            };

            // Add message to history FIRST, before clearing streaming state
            $agentMessages.set({
                ...$agentMessages.get(),
                data: [...$agentMessages.get().data, msgWithEvents],
                loading: false,
                error: null
            });
            
            // Now clear streaming state after message is in history
            $agentStreamingMessage.set("");
            $agentStreamingEvents.set([]);
            $isAgentStreaming.set(false);
        });

        agentStream.on("error", (err) => {
            console.error("Agent chat stream error:", err);
            toast.error("Error in agent chat");
            $isAgentStreaming.set(false);
        });

    } catch (error) {
        console.error("Failed to send message:", error);
        toast.error("Failed to send message");
        $isAgentStreaming.set(false);
    }
};


// Listeners
// $currentAgentId.listen((agentId) => {
//     if (agentId) {
//         getSessions(agentId);
//     }
// });

$currentSessionId.listen((sessionId) => {
    if (sessionId) {
        getAgentMessages(sessionId);
    } else {
        $agentMessages.set({ data: [], loading: false, error: null });
    }
});

// Init
onMount($agents, () => {
    getAgents();
});
