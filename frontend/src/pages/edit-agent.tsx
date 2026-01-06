import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChevronLeft, Loader2, Trash2, Download, File } from "lucide-react";
import { FileUploader } from "@/components/FileUploader";
import { getAgentClient } from "@/store/agents";
import { GetAgentsRequest } from "../../proto/chatservice";
import { toast } from "sonner";
import { getJWTToken } from "@/lib/auth";
import { getUIConfig } from "@/lib/config";

type TabType = "agent" | "files";

interface AgentFile {
    id: string;
    agent_id: string;
    docs_id: string;
    file_name: string;
    file_path: string;
    file_size: number;
    created_at: string;
}

export function EditAgentPage() {
    const navigate = useNavigate();
    const { agentId } = useParams<{ agentId: string }>();
    const [isLoading, setIsLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const [agentFiles, setAgentFiles] = useState<AgentFile[]>([]);
    const [loadingFiles, setLoadingFiles] = useState(true);
    const [activeTab, setActiveTab] = useState<TabType>("agent");
    
    const [formData, setFormData] = useState({
        name: "",
        description: "",
        systemPrompt: "",
        provider: "openai",
        model: "gpt-4o",
    });

    useEffect(() => {
        if (agentId) {
            loadAgentData();
            loadAgentFiles();
        }
    }, [agentId]);

    const loadAgentData = async () => {
        try {
            setIsLoading(true);
            const response = await getAgentClient().GetAgents(
                GetAgentsRequest.fromObject({}),
                {}
            );
            
            const agent = response.agents.find(a => a.id === agentId);
            if (agent) {
                setFormData({
                    name: agent.name,
                    description: agent.description,
                    systemPrompt: agent.system_prompt,
                    provider: agent.provider,
                    model: agent.model,
                });
            } else {
                toast.error("Agent not found");
                navigate("/");
            }
        } catch (error) {
            console.error("Failed to load agent", error);
            toast.error("Failed to load agent");
        } finally {
            setIsLoading(false);
        }
    };

    const loadAgentFiles = async () => {
        try {
            setLoadingFiles(true);
            const config = getUIConfig();
            if (!config) {
                console.error("UI config not loaded");
                return;
            }

            const token = getJWTToken();
            const url = `${config.API_UPLOAD_URL}/agents/files-list/${agentId}`;
            const response = await fetch(url, {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });
            
            if (response.ok) {
                const data = await response.json();
                setAgentFiles(data.files || []);
            } else {
                console.error("Failed to load agent files:", response.status, response.statusText);
            }
        } catch (error) {
            console.error("Failed to load agent files", error);
        } finally {
            setLoadingFiles(false);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSaving(true);
        try {
            // TODO: Implement update agent API
            toast.info("Update agent API not yet implemented");
            // For now just show success and go back
            setTimeout(() => {
                navigate("/");
            }, 1000);
        } catch (error) {
            console.error("Failed to update agent", error);
            toast.error("Failed to update agent");
        } finally {
            setIsSaving(false);
        }
    };

    const handleFileUploadComplete = () => {
        loadAgentFiles();
    };

    const handleDeleteFile = async (docsId: string) => {
        if (!confirm("Are you sure you want to delete this file?")) {
            return;
        }

        try {
            const config = getUIConfig();
            if (!config) {
                toast.error("Configuration not loaded");
                return;
            }

            const token = getJWTToken();
            const url = `${config.API_UPLOAD_URL}/agents/files/delete`;
            const response = await fetch(url, {
                method: "DELETE",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify({
                    agent_id: agentId,
                    docs_id: docsId,
                }),
            });

            if (response.ok) {
                toast.success("File deleted");
                loadAgentFiles();
            } else {
                const errorText = await response.text();
                console.error("Failed to delete file:", response.status, errorText);
                toast.error("Failed to delete file");
            }
        } catch (error) {
            console.error("Failed to delete file", error);
            toast.error("Failed to delete file");
        }
    };

    const formatFileSize = (bytes: number) => {
        if (bytes < 1024) return bytes + " B";
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
        return (bytes / (1024 * 1024)).toFixed(1) + " MB";
    };

    const buildFileTree = (files: AgentFile[]) => {
        const tree: any = {};
        files.forEach((file) => {
            const parts = file.file_path.split("/");
            let current = tree;
            
            parts.forEach((part, idx) => {
                if (idx === parts.length - 1) {
                    current[part] = file;
                } else {
                    if (!current[part]) {
                        current[part] = {};
                    }
                    current = current[part];
                }
            });
        });
        return tree;
    };

    const renderFileTree = (node: any, path = "", depth = 0) => {
        return Object.entries(node).map(([key, value]: [string, any]) => {
            if (value.docs_id) {
                // It's a file
                return (
                    <div
                        key={value.id}
                        className="flex items-center justify-between p-2 hover:bg-muted rounded-md group"
                        style={{ marginLeft: `${depth * 20}px` }}
                    >
                        <div className="flex items-center gap-2 flex-1 min-w-0">
                            <File className="h-4 w-4 shrink-0 text-muted-foreground" />
                            <span className="text-sm truncate">{key}</span>
                            <span className="text-xs text-muted-foreground">
                                ({formatFileSize(value.file_size)})
                            </span>
                        </div>
                        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100">
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-7 w-7 p-0"
                                onClick={() => {
                                    const config = getUIConfig();
                                    if (config) {
                                        window.open(`${config.API_UPLOAD_URL}/agents/files/${value.docs_id}`, "_blank");
                                    }
                                }}
                                title="Download"
                            >
                                <Download className="h-3 w-3" />
                            </Button>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-7 w-7 p-0 text-red-600 hover:text-red-700"
                                onClick={() => handleDeleteFile(value.docs_id)}
                                title="Delete"
                            >
                                <Trash2 className="h-3 w-3" />
                            </Button>
                        </div>
                    </div>
                );
            } else {
                // It's a folder
                return (
                    <div key={path + key} style={{ marginLeft: `${depth * 20}px` }}>
                        <div className="flex items-center gap-2 p-2 text-sm font-medium text-muted-foreground">
                            📁 {key}
                        </div>
                        {renderFileTree(value, path + key + "/", depth + 1)}
                    </div>
                );
            }
        });
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-screen">
                <Loader2 className="h-8 w-8 animate-spin" />
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full w-full">
            {/* Header - Fixed */}
            <div className="flex-shrink-0 p-6 border-b border-border bg-card">
                <div className="max-w-4xl mx-auto w-full">
                    <div className="flex items-center space-x-4">
                        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
                            <ChevronLeft className="h-4 w-4" />
                        </Button>
                        <h1 className="text-2xl font-bold">Edit Agent</h1>
                    </div>
                </div>
            </div>

            {/* Tabs */}
            <div className="flex-shrink-0 border-b border-border bg-card">
                <div className="max-w-4xl mx-auto w-full">
                    <div className="flex gap-1 px-6">
                        <button
                            className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                                activeTab === "agent"
                                    ? "border-primary text-primary"
                                    : "border-transparent text-muted-foreground hover:text-foreground"
                            }`}
                            onClick={() => setActiveTab("agent")}
                        >
                            Agent
                        </button>
                        <button
                            className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                                activeTab === "files"
                                    ? "border-primary text-primary"
                                    : "border-transparent text-muted-foreground hover:text-foreground"
                            }`}
                            onClick={() => setActiveTab("files")}
                        >
                            Files {agentFiles.length > 0 && `(${agentFiles.length})`}
                        </button>
                    </div>
                </div>
            </div>

            {/* Scrollable Content */}
            <div className="flex-1 overflow-y-auto min-h-0 bg-background">
                <div className="max-w-4xl mx-auto w-full p-6 pb-12">
                    {activeTab === "agent" && (
                        <form onSubmit={handleSubmit} className="space-y-6 bg-card p-6 rounded-lg border border-border">
                            <div className="space-y-2">
                                <label className="text-sm font-medium">Name</label>
                                <Input
                                    required
                                    placeholder="e.g. Coding Assistant"
                                    value={formData.name}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                />
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium">Description</label>
                                <Input
                                    placeholder="Brief description of what this agent does"
                                    value={formData.description}
                                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                />
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium">System Prompt</label>
                                <textarea
                                    className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                    placeholder="You are a helpful assistant..."
                                    value={formData.systemPrompt}
                                    onChange={(e) => setFormData({ ...formData, systemPrompt: e.target.value })}
                                />
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">Provider</label>
                                    <select
                                        className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                        value={formData.provider}
                                        onChange={(e) => setFormData({ ...formData, provider: e.target.value })}
                                    >
                                        <option value="openai">OpenAI</option>
                                        <option value="anthropic">Anthropic</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">Model</label>
                                    <select
                                        className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                        value={formData.model}
                                        onChange={(e) => setFormData({ ...formData, model: e.target.value })}
                                    >
                                        <option value="gpt-4o">GPT-4o</option>
                                        <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
                                        <option value="claude-3-opus-20240229">Claude 3 Opus</option>
                                        <option value="claude-3-sonnet-20240229">Claude 3 Sonnet</option>
                                    </select>
                                </div>
                            </div>

                            <div className="pt-4 flex justify-end gap-2">
                                <Button variant="outline" onClick={() => navigate("/")}>
                                    Cancel
                                </Button>
                                <Button type="submit" disabled={isSaving}>
                                    {isSaving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                    Save Changes
                                </Button>
                            </div>
                        </form>
                    )}

                    {activeTab === "files" && (
                        <div className="space-y-4 bg-card p-6 rounded-lg border border-border">
                            <div className="space-y-2">
                                <h2 className="text-xl font-semibold">Agent Files</h2>
                                <p className="text-sm text-muted-foreground">
                                    Files and folders that this agent can access during conversations.
                                </p>
                            </div>

                            {loadingFiles ? (
                                <div className="flex items-center justify-center p-8">
                                    <Loader2 className="h-6 w-6 animate-spin" />
                                </div>
                            ) : agentFiles.length > 0 ? (
                                <div className="space-y-2 border rounded-md p-4 max-h-96 overflow-y-auto">
                                    <div className="flex items-center justify-between mb-2">
                                        <span className="text-sm font-medium">
                                            {agentFiles.length} file(s)
                                        </span>
                                    </div>
                                    {renderFileTree(buildFileTree(agentFiles))}
                                </div>
                            ) : (
                                <div className="text-center text-muted-foreground p-8 border rounded-md border-dashed">
                                    No files uploaded yet
                                </div>
                            )}

                            <div className="pt-4 border-t">
                                <h3 className="text-sm font-medium mb-3">Upload Additional Files</h3>
                                <FileUploader
                                    uploadUrl="/agents/upload"
                                    agentId={agentId}
                                    onCompleteUpload={handleFileUploadComplete}
                                />
                            </div>

                            <div className="flex justify-end gap-2">
                                <Button variant="outline" onClick={() => navigate("/")}>
                                    Done
                                </Button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

