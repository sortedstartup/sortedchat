import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Plus, Trash2, ChevronDown, ChevronUp, Power } from "lucide-react";
import { MCPServer } from "../../proto/chatservice";

interface MCPServerFormProps {
    servers: MCPServer[];
    onChange: (servers: MCPServer[]) => void;
}

export function MCPServerForm({ servers, onChange }: MCPServerFormProps) {
    const [expandedServers, setExpandedServers] = useState<Set<number>>(new Set([0]));

    const addServer = () => {
        const newServer = MCPServer.fromObject({
            server_name: "",
            is_enabled: true,
            stdio: {
                command: "",
                arguments: [],
                environment_variables: {}
            }
        });
        onChange([...servers, newServer]);
        setExpandedServers(new Set([...expandedServers, servers.length]));
    };

    const removeServer = (index: number) => {
        onChange(servers.filter((_, i) => i !== index));
        const newExpanded = new Set(expandedServers);
        newExpanded.delete(index);
        setExpandedServers(newExpanded);
    };

    const updateServer = (index: number, updates: any) => {
        const updated = [...servers];
        const server = updated[index];
        const currentData = server.toObject();
        
        // Deep merge updates with current data
        const newData: any = { ...currentData };
        
        // Handle nested updates (stdio, http)
        if (updates.stdio) {
            newData.stdio = { ...(currentData.stdio || {}), ...updates.stdio };
        }
        if (updates.http) {
            newData.http = { ...(currentData.http || {}), ...updates.http };
        }
        
        // Handle top-level updates
        Object.keys(updates).forEach(key => {
            if (key !== 'stdio' && key !== 'http') {
                newData[key] = updates[key];
            }
        });
        
        // Create new server instance from merged data
        updated[index] = MCPServer.fromObject(newData);
        
        onChange([...updated]);
    };

    const toggleExpanded = (index: number) => {
        const newExpanded = new Set(expandedServers);
        if (newExpanded.has(index)) {
            newExpanded.delete(index);
        } else {
            newExpanded.add(index);
        }
        setExpandedServers(newExpanded);
    };

    const addArgument = (serverIndex: number) => {
        const server = servers[serverIndex].toObject();
        const args = server.stdio?.arguments || [];
        updateServer(serverIndex, {
            stdio: {
                ...server.stdio,
                arguments: [...args, ""]
            }
        });
    };

    const updateArgument = (serverIndex: number, argIndex: number, value: string) => {
        const server = servers[serverIndex].toObject();
        const newArgs = [...(server.stdio?.arguments || [])];
        newArgs[argIndex] = value;
        updateServer(serverIndex, {
            stdio: { ...server.stdio, arguments: newArgs }
        });
    };

    const removeArgument = (serverIndex: number, argIndex: number) => {
        const server = servers[serverIndex].toObject();
        const args = (server.stdio?.arguments || []).filter((_, i) => i !== argIndex);
        updateServer(serverIndex, {
            stdio: { ...server.stdio, arguments: args }
        });
    };

    const addEnvVar = (serverIndex: number) => {
        const server = servers[serverIndex].toObject();
        const envVars = server.stdio?.environment_variables || {};
        const newKey = `VAR_${Object.keys(envVars).length + 1}`;
        updateServer(serverIndex, {
            stdio: {
                ...server.stdio,
                environment_variables: { ...envVars, [newKey]: "" }
            }
        });
    };

    const updateEnvVar = (serverIndex: number, oldKey: string, newKey: string, value: string) => {
        const server = servers[serverIndex].toObject();
        const newEnvVars = { ...(server.stdio?.environment_variables || {}) };
        if (oldKey !== newKey) {
            delete newEnvVars[oldKey];
        }
        newEnvVars[newKey] = value;
        updateServer(serverIndex, {
            stdio: { ...server.stdio, environment_variables: newEnvVars }
        });
    };

    const removeEnvVar = (serverIndex: number, key: string) => {
        const server = servers[serverIndex].toObject();
        const newEnvVars = { ...(server.stdio?.environment_variables || {}) };
        delete newEnvVars[key];
        updateServer(serverIndex, {
            stdio: { ...server.stdio, environment_variables: newEnvVars }
        });
    };

    const addHeader = (serverIndex: number) => {
        const server = servers[serverIndex].toObject();
        const headers = server.http?.headers || {};
        const newKey = `Header-${Object.keys(headers).length + 1}`;
        updateServer(serverIndex, {
            http: {
                ...server.http,
                headers: { ...headers, [newKey]: "" }
            }
        });
    };

    const updateHeader = (serverIndex: number, oldKey: string, newKey: string, value: string) => {
        const server = servers[serverIndex].toObject();
        const newHeaders = { ...(server.http?.headers || {}) };
        if (oldKey !== newKey) {
            delete newHeaders[oldKey];
        }
        newHeaders[newKey] = value;
        updateServer(serverIndex, {
            http: { ...server.http, headers: newHeaders }
        });
    };

    const removeHeader = (serverIndex: number, key: string) => {
        const server = servers[serverIndex].toObject();
        const newHeaders = { ...(server.http?.headers || {}) };
        delete newHeaders[key];
        updateServer(serverIndex, {
            http: { ...server.http, headers: newHeaders }
        });
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div>
                    <label className="text-sm font-medium">MCP Servers</label>
                    <p className="text-xs text-muted-foreground">
                        Configure Model Context Protocol servers for extended capabilities
                    </p>
                </div>
                <Button type="button" variant="outline" size="sm" onClick={addServer}>
                    <Plus className="h-4 w-4 mr-1" />
                    Add Server
                </Button>
            </div>

            {servers.length === 0 ? (
                <div className="text-center text-muted-foreground text-sm border border-dashed rounded-md p-4">
                    No MCP servers configured. Click "Add Server" to get started.
                </div>
            ) : (
                <div className="space-y-3">
                    {servers.map((server, serverIndex) => {
                        const isExpanded = expandedServers.has(serverIndex);
                        const serverData = server.toObject();
                        const hasStdio = !!serverData.stdio;
                        
                        return (
                            <div key={serverIndex} className="border rounded-lg p-4 space-y-3 bg-card">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-3 flex-1">
                                        <Badge 
                                            variant={serverData.is_enabled ? "default" : "secondary"}
                                            className="cursor-pointer"
                                            onClick={() => updateServer(serverIndex, { is_enabled: !serverData.is_enabled })}
                                        >
                                            <Power className="h-3 w-3 mr-1" />
                                            {serverData.is_enabled ? "Enabled" : "Disabled"}
                                        </Badge>
                                        <Input
                                            placeholder="Server name"
                                            value={serverData.server_name}
                                            onChange={(e) =>
                                                updateServer(serverIndex, { server_name: e.target.value })
                                            }
                                            className="flex-1"
                                        />
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            onClick={() => toggleExpanded(serverIndex)}
                                        >
                                            {isExpanded ? (
                                                <ChevronUp className="h-4 w-4" />
                                            ) : (
                                                <ChevronDown className="h-4 w-4" />
                                            )}
                                        </Button>
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            onClick={() => removeServer(serverIndex)}
                                            className="text-red-600 hover:text-red-700"
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </div>

                                {isExpanded && (
                                    <div className="space-y-3 pt-2 border-t">
                                        <div className="space-y-2">
                                            <label className="text-xs font-medium">Transport Type</label>
                                            <div className="flex gap-4">
                                                <label className="flex items-center gap-2">
                                                    <input
                                                        type="radio"
                                                        checked={hasStdio}
                                                        onChange={() =>
                                                            updateServer(serverIndex, { 
                                                                stdio: { command: "", arguments: [], environment_variables: {} },
                                                                http: undefined 
                                                            })
                                                        }
                                                    />
                                                    <span className="text-sm">STDIO</span>
                                                </label>
                                                <label className="flex items-center gap-2">
                                                    <input
                                                        type="radio"
                                                        checked={!hasStdio}
                                                        onChange={() =>
                                                            updateServer(serverIndex, { 
                                                                http: { url: "", headers: {}, timeout_seconds: 30 },
                                                                stdio: undefined 
                                                            })
                                                        }
                                                    />
                                                    <span className="text-sm">HTTP</span>
                                                </label>
                                            </div>
                                        </div>

                                        {hasStdio ? (
                                            <>
                                                <div className="space-y-2">
                                                    <label className="text-xs font-medium">Command</label>
                                                    <Input
                                                        placeholder="e.g., npx"
                                                        value={serverData.stdio?.command || ""}
                                                        onChange={(e) =>
                                                            updateServer(serverIndex, { 
                                                                stdio: { ...serverData.stdio, command: e.target.value }
                                                            })
                                                        }
                                                    />
                                                </div>

                                                <div className="space-y-2">
                                                    <div className="flex items-center justify-between">
                                                        <label className="text-xs font-medium">Arguments</label>
                                                        <Button
                                                            type="button"
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() => addArgument(serverIndex)}
                                                        >
                                                            <Plus className="h-3 w-3 mr-1" />
                                                            Add
                                                        </Button>
                                                    </div>
                                                    {(serverData.stdio?.arguments || []).map((arg, argIndex) => (
                                                        <div key={argIndex} className="flex gap-2">
                                                            <Input
                                                                placeholder="Argument"
                                                                value={arg}
                                                                onChange={(e) =>
                                                                    updateArgument(serverIndex, argIndex, e.target.value)
                                                                }
                                                            />
                                                            <Button
                                                                type="button"
                                                                variant="ghost"
                                                                size="icon"
                                                                onClick={() => removeArgument(serverIndex, argIndex)}
                                                            >
                                                                <Trash2 className="h-4 w-4" />
                                                            </Button>
                                                        </div>
                                                    ))}
                                                </div>

                                                <div className="space-y-2">
                                                    <div className="flex items-center justify-between">
                                                        <label className="text-xs font-medium">Environment Variables</label>
                                                        <Button
                                                            type="button"
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() => addEnvVar(serverIndex)}
                                                        >
                                                            <Plus className="h-3 w-3 mr-1" />
                                                            Add
                                                        </Button>
                                                    </div>
                                                    {Object.entries(serverData.stdio?.environment_variables || {}).map(
                                                        ([key, value]) => (
                                                            <div key={key} className="flex gap-2">
                                                                <Input
                                                                    placeholder="Key"
                                                                    value={key}
                                                                    onChange={(e) =>
                                                                        updateEnvVar(
                                                                            serverIndex,
                                                                            key,
                                                                            e.target.value,
                                                                            value
                                                                        )
                                                                    }
                                                                    className="flex-1"
                                                                />
                                                                <Input
                                                                    placeholder="Value"
                                                                    value={value}
                                                                    onChange={(e) =>
                                                                        updateEnvVar(
                                                                            serverIndex,
                                                                            key,
                                                                            key,
                                                                            e.target.value
                                                                        )
                                                                    }
                                                                    className="flex-1"
                                                                />
                                                                <Button
                                                                    type="button"
                                                                    variant="ghost"
                                                                    size="icon"
                                                                    onClick={() => removeEnvVar(serverIndex, key)}
                                                                >
                                                                    <Trash2 className="h-4 w-4" />
                                                                </Button>
                                                            </div>
                                                        )
                                                    )}
                                                </div>
                                            </>
                                        ) : (
                                            <>
                                                <div className="space-y-2">
                                                    <label className="text-xs font-medium">URL</label>
                                                    <Input
                                                        placeholder="https://example.com/mcp"
                                                        value={serverData.http?.url || ""}
                                                        onChange={(e) =>
                                                            updateServer(serverIndex, { 
                                                                http: { ...serverData.http, url: e.target.value }
                                                            })
                                                        }
                                                    />
                                                </div>

                                                <div className="space-y-2">
                                                    <div className="flex items-center justify-between">
                                                        <label className="text-xs font-medium">Headers</label>
                                                        <Button
                                                            type="button"
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() => addHeader(serverIndex)}
                                                        >
                                                            <Plus className="h-3 w-3 mr-1" />
                                                            Add
                                                        </Button>
                                                    </div>
                                                    {Object.entries(serverData.http?.headers || {}).map(([key, value]) => (
                                                        <div key={key} className="flex gap-2">
                                                            <Input
                                                                placeholder="Header name"
                                                                value={key}
                                                                onChange={(e) =>
                                                                    updateHeader(
                                                                        serverIndex,
                                                                        key,
                                                                        e.target.value,
                                                                        value
                                                                    )
                                                                }
                                                                className="flex-1"
                                                            />
                                                            <Input
                                                                placeholder="Header value"
                                                                value={value}
                                                                onChange={(e) =>
                                                                    updateHeader(
                                                                        serverIndex,
                                                                        key,
                                                                        key,
                                                                        e.target.value
                                                                    )
                                                                }
                                                                className="flex-1"
                                                            />
                                                            <Button
                                                                type="button"
                                                                variant="ghost"
                                                                size="icon"
                                                                onClick={() => removeHeader(serverIndex, key)}
                                                            >
                                                                <Trash2 className="h-4 w-4" />
                                                            </Button>
                                                        </div>
                                                    ))}
                                                </div>

                                                <div className="space-y-2">
                                                    <label className="text-xs font-medium">Timeout (seconds)</label>
                                                    <Input
                                                        type="number"
                                                        placeholder="30"
                                                        value={serverData.http?.timeout_seconds || ""}
                                                        onChange={(e) =>
                                                            updateServer(serverIndex, {
                                                                http: { 
                                                                    ...serverData.http, 
                                                                    timeout_seconds: parseInt(e.target.value) || 0 
                                                                }
                                                            })
                                                        }
                                                    />
                                                </div>
                                            </>
                                        )}
                                    </div>
                                )}
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
}

