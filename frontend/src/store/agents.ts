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
    SortedChatClient,
} from "../../proto/chatservice";
import { createAuthenticatedClientOptions } from "../lib/auth";
import { getUIConfig } from "../lib/config";
import type { ClientReadableStream } from "grpc-web";

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
    data: AgentMessage[];
    loading: boolean;
    error: string | null;
}>({
    data: [],
    loading: false,
    error: null,
});

export const $isAgentStreaming = atom<boolean>(false);
export const $agentStreamingMessage = atom<string>("");
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
    // Temporary ID and sequence until refresh? Or just optimistic update
    userMsg.id = "temp-" + Date.now();

    $agentMessages.set({
        ...$agentMessages.get(),
        data: [...currentMessages, userMsg]
    });

    $isAgentStreaming.set(true);
    $agentStreamingMessage.set("");

    try {
        agentStream = getAgentClient().AgentChat(
            AgentChatRequest.fromObject({
                session_id: sessionId,
                message: message
            }),
            {}
        );

        let assistantContent = "";

        agentStream.on("data", (res: AgentChatResponse) => {
            if (res.has_message) {
                assistantContent += res.message;
                $agentStreamingMessage.set(assistantContent);
            }
            // TODO: Handle thinking, tool_call, tool_result
        });

        agentStream.on("end", () => {
            const assistantMsg = new AgentMessage();
            assistantMsg.role = "assistant";
            assistantMsg.content = assistantContent;
            assistantMsg.type = "text";
            assistantMsg.id = "temp-resp-" + Date.now();

            $agentMessages.set({
                ...$agentMessages.get(),
                data: [...$agentMessages.get().data, assistantMsg],
                loading: false,
                error: null
            });
            $isAgentStreaming.set(false);
            $agentStreamingMessage.set("");
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
