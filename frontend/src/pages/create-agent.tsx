
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createAgent } from "@/store/agents";
import { ChevronLeft, Loader2 } from "lucide-react";
import { FileUploader } from "@/components/FileUploader";
import { ModelSelector } from "@/components/ModelSelector";

export function CreateAgentPage() {
    const navigate = useNavigate();
    const [isLoading, setIsLoading] = useState(false);
    const [createdAgentId, setCreatedAgentId] = useState<string | null>(null);
    const [uploadedFilesCount, setUploadedFilesCount] = useState(0);
    const [formData, setFormData] = useState({
        name: "",
        description: "",
        systemPrompt: "",
        provider: "openai",
        model: "gpt-4o",
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        try {
            const agentId = await createAgent(
                formData.name,
                formData.description,
                formData.systemPrompt,
                formData.model,
                formData.provider
            );
            // Store the agent ID to enable file uploads
            setCreatedAgentId(agentId);
        } catch (error) {
            console.error("Failed to create agent", error);
            // Toast is handled in store
        } finally {
            setIsLoading(false);
        }
    };

    const handleFileUploadComplete = (files: any[]) => {
        // files array only contains successfully uploaded files now
        setUploadedFilesCount(prev => prev + files.length);
    };

    const handleFinish = () => {
        navigate("/");
    };

    return (
        <div className="flex flex-col h-full w-full">
            {/* Header - Fixed */}
            <div className="flex-shrink-0 p-6 border-b border-border bg-card">
                <div className="max-w-2xl mx-auto w-full">
                    <div className="flex items-center space-x-4">
                        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
                            <ChevronLeft className="h-4 w-4" />
                        </Button>
                        <h1 className="text-2xl font-bold">Create New Agent</h1>
                    </div>
                </div>
            </div>

            {/* Scrollable Content */}
            <div className="flex-1 overflow-y-auto min-h-0 bg-background">
                <div className="max-w-2xl mx-auto w-full space-y-8 p-6 pb-12">

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

                    <div className="space-y-2">
                        <label className="text-sm font-medium">Model</label>
                        <ModelSelector
                            selectedModelId={formData.model}
                            onSelectModel={(modelId, provider) => {
                                setFormData({ ...formData, model: modelId, provider });
                            }}
                            className="w-full"
                        />
                    </div>

                    <div className="pt-4 flex justify-end">
                        <Button type="submit" disabled={isLoading || createdAgentId !== null}>
                            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            {createdAgentId ? "Agent Created ✓" : "Create Agent"}
                        </Button>
                    </div>
                </form>

                {createdAgentId && (
                    <div className="space-y-4 bg-card p-6 rounded-lg border border-border">
                        <div className="space-y-2">
                            <h2 className="text-xl font-semibold">Upload Files (Optional)</h2>
                            <p className="text-sm text-muted-foreground">
                                Upload files or folders that your agent can access during conversations.
                                You can skip this step and add files later.
                            </p>
                        </div>

                        <FileUploader
                            uploadUrl="/agents/upload"
                            agentId={createdAgentId}
                            onCompleteUpload={handleFileUploadComplete}
                        />

                        {uploadedFilesCount > 0 && (
                            <p className="text-sm text-green-600">
                                ✓ {uploadedFilesCount} file(s) uploaded successfully
                            </p>
                        )}

                        <div className="pt-4 flex justify-end gap-2">
                            <Button variant="outline" onClick={handleFinish}>
                                Skip & Finish
                            </Button>
                            <Button onClick={handleFinish} disabled={uploadedFilesCount === 0}>
                                Finish
                            </Button>
                        </div>
                    </div>
                )}
                </div>
            </div>
        </div>
    );
}
