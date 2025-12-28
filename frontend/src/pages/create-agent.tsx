
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createAgent } from "@/store/agents";
import { ChevronLeft, Loader2 } from "lucide-react";

export function CreateAgentPage() {
    const navigate = useNavigate();
    const [isLoading, setIsLoading] = useState(false);
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
            await createAgent(
                formData.name,
                formData.description,
                formData.systemPrompt,
                formData.model,
                formData.provider
            );
            // Navigate back or to the list, let's go back for now or maybe to the new agent?
            // createAgent returns agent_id but we might just want to go back to home or let user pick.
            // For now, let's go back to home which will show the new agent in sidebar.
            navigate("/");
        } catch (error) {
            console.error("Failed to create agent", error);
            // Toast is handled in store
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="flex flex-col h-screen w-full bg-background p-6">
            <div className="max-w-2xl mx-auto w-full space-y-8">
                <div className="flex items-center space-x-4">
                    <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
                        <ChevronLeft className="h-4 w-4" />
                    </Button>
                    <h1 className="text-2xl font-bold">Create New Agent</h1>
                </div>

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

                    <div className="pt-4 flex justify-end">
                        <Button type="submit" disabled={isLoading}>
                            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Create Agent
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    );
}
