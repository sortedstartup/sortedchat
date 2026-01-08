import React, { useState, useEffect, useRef } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChevronLeft, Loader2, Trash2, Download, File, ChevronRight, ChevronDown, FileText, Code2, Save, AlertCircle } from "lucide-react";
import { FileUploader } from "@/components/FileUploader";
import { getAgentClient, updateAgent } from "@/store/agents";
import { GetAgentsRequest } from "../../proto/chatservice";
import { toast } from "sonner";
import { getJWTToken } from "@/lib/auth";
import { getUIConfig } from "@/lib/config";
import { ModelSelector } from "@/components/ModelSelector";
import CodeMirror from '@uiw/react-codemirror';
import { githubLight, githubDark } from '@uiw/codemirror-theme-github';
import { javascript } from '@codemirror/lang-javascript';
import { python } from '@codemirror/lang-python';
import { html } from '@codemirror/lang-html';
import { markdown as codeMirrorMarkdown } from '@codemirror/lang-markdown';
import { EditorState } from 'prosemirror-state';
import { EditorView } from 'prosemirror-view';
import { defaultMarkdownParser, defaultMarkdownSerializer, schema as markdownSchema } from 'prosemirror-markdown';
import { keymap } from 'prosemirror-keymap';
import { history, undo, redo } from 'prosemirror-history';
import { baseKeymap } from 'prosemirror-commands';
import { inputRules, wrappingInputRule, textblockTypeInputRule, InputRule } from 'prosemirror-inputrules';
import { splitListItem, liftListItem, sinkListItem } from 'prosemirror-schema-list';

type TabType = "agent" | "files" | "editor";

// Create markdown input rules for WYSIWYG editing
function buildMarkdownInputRules() {
    const rules: InputRule[] = [];
    
    // Heading rules: # to ######
    for (let level = 1; level <= 6; level++) {
        rules.push(
            textblockTypeInputRule(
                new RegExp(`^(#{${level}})\\s$`),
                markdownSchema.nodes.heading,
                { level }
            )
        );
    }
    
    // Bullet list rule: * or -
    if (markdownSchema.nodes.bullet_list) {
        rules.push(
            wrappingInputRule(
                /^\s*([-*])\s$/,
                markdownSchema.nodes.bullet_list
            )
        );
    }
    
    // Ordered list rule: 1.
    if (markdownSchema.nodes.ordered_list) {
        rules.push(
            wrappingInputRule(
                /^(\d+)\.\s$/,
                markdownSchema.nodes.ordered_list,
                (match) => ({ order: +match[1] }),
                (match, node) => node.childCount + node.attrs.order === +match[1]
            )
        );
    }
    
    // Code block rule: ```
    rules.push(
        textblockTypeInputRule(
            /^```$/,
            markdownSchema.nodes.code_block
        )
    );
    
    // Blockquote rule: >
    rules.push(
        wrappingInputRule(
            /^\s*>\s$/,
            markdownSchema.nodes.blockquote
        )
    );
    
    return inputRules({ rules });
}

function MarkdownEditor({ content, isDark, onChange }: { content: string; isDark: boolean; onChange: (text: string) => void }) {
    const mountRef = useRef<HTMLDivElement>(null);
    const [editorView, setEditorView] = useState<EditorView | null>(null);
    
    useEffect(() => {
        if (!mountRef.current) return;

        const doc = defaultMarkdownParser.parse(content) || markdownSchema.node('doc', null, [markdownSchema.node('paragraph')]);
        
        const state = EditorState.create({
            doc,
            schema: markdownSchema,
            plugins: [
                buildMarkdownInputRules(),
                history(),
                keymap({
                    'Enter': splitListItem(markdownSchema.nodes.list_item),
                    'Mod-[': liftListItem(markdownSchema.nodes.list_item),
                    'Mod-]': sinkListItem(markdownSchema.nodes.list_item),
                    'Mod-z': undo,
                    'Mod-y': redo,
                    'Mod-Shift-z': redo,
                }),
                keymap(baseKeymap),
            ],
        });

        const view = new EditorView(mountRef.current, {
            state,
            dispatchTransaction(transaction) {
                const newState = view.state.apply(transaction);
                view.updateState(newState);
                
                // Update text content for potential saving later
                const markdown = defaultMarkdownSerializer.serialize(newState.doc);
                onChange(markdown);
            },
        });

        setEditorView(view);

        return () => {
            view.destroy();
        };
    }, []);

    // Update content when file changes
    useEffect(() => {
        if (editorView && content !== defaultMarkdownSerializer.serialize(editorView.state.doc)) {
            const doc = defaultMarkdownParser.parse(content) || markdownSchema.node('doc', null, [markdownSchema.node('paragraph')]);
            const state = EditorState.create({
                doc,
                schema: markdownSchema,
                plugins: editorView.state.plugins,
            });
            editorView.updateState(state);
        }
    }, [content, editorView]);

    return (
        <div className={`h-full overflow-y-auto p-4 prose ${isDark ? 'prose-invert' : ''} max-w-none`}>
            <div ref={mountRef} className="ProseMirror-wrapper" />
        </div>
    );
}

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
    
    // Editor tab state
    const [selectedFile, setSelectedFile] = useState<AgentFile | null>(null);
    const [fileContent, setFileContent] = useState<string>("");
    const [originalContent, setOriginalContent] = useState<string>("");
    const [loadingContent, setLoadingContent] = useState(false);
    const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set());
    const [isDarkMode, setIsDarkMode] = useState(false);
    const [isRawMode, setIsRawMode] = useState(false);
    const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
    const [isSavingFile, setIsSavingFile] = useState(false);
    const [showUnsavedModal, setShowUnsavedModal] = useState(false);
    const [pendingFile, setPendingFile] = useState<AgentFile | null>(null);

    // Detect dark mode
    useEffect(() => {
        const checkDarkMode = () => {
            const isDark = document.documentElement.classList.contains('dark');
            setIsDarkMode(isDark);
        };
        
        checkDarkMode();
        
        // Watch for theme changes
        const observer = new MutationObserver(checkDarkMode);
        observer.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class']
        });
        
        return () => observer.disconnect();
    }, []);
    
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
        if (!agentId) return;

        try {
            await updateAgent(
                agentId,
                formData.name,
                formData.description,
                formData.systemPrompt,
                formData.model,
                formData.provider
            );
            // navigate("/");
        } catch (error) {
            console.error("Failed to update agent", error);
        } finally {
            setIsSaving(false);
        }
    };

    const handleFileUploadComplete = () => {
        loadAgentFiles();
    };

    const isMarkdownFile = (fileName: string) => {
        const ext = fileName.split('.').pop()?.toLowerCase();
        return ext === 'md' || ext === 'markdown';
    };

    const loadFileContent = async (file: AgentFile) => {
        try {
            setLoadingContent(true);
            setIsRawMode(false); // Reset to preview mode when loading a new file
            const config = getUIConfig();
            if (!config) {
                toast.error("Configuration not loaded");
                return;
            }

            const token = getJWTToken();
            const url = `${config.API_UPLOAD_URL}/agents/files/${file.docs_id}`;
            console.log("Loading file:", { url, docs_id: file.docs_id, file_name: file.file_name });
            
            const response = await fetch(url, {
                method: "GET",
                headers: {
                    Authorization: `Bearer ${token}`,
                },
                cache: 'no-store', // Don't use cached version
            });

            console.log("Load response:", response.status);

            if (response.ok) {
                const text = await response.text();
                console.log("Loaded content length:", text.length, "first 100 chars:", text.substring(0, 100));
                setFileContent(text);
                setOriginalContent(text);
                setSelectedFile(file);
                setHasUnsavedChanges(false);
            } else {
                toast.error("Failed to load file content");
            }
        } catch (error) {
            console.error("Failed to load file content", error);
            toast.error("Failed to load file content");
        } finally {
            setLoadingContent(false);
        }
    };

    const handleFileContentChange = (newContent: string) => {
        setFileContent(newContent);
        setHasUnsavedChanges(newContent !== originalContent);
    };

    const saveFileContent = async (): Promise<boolean> => {
        if (!selectedFile) return false;

        try {
            setIsSavingFile(true);
            const config = getUIConfig();
            if (!config) {
                toast.error("Configuration not loaded");
                return false;
            }

            const token = getJWTToken();
            const url = `${config.API_UPLOAD_URL}/agents/files/update`;
            
            console.log("Saving file:", {
                url,
                agent_id: agentId,
                docs_id: selectedFile.docs_id,
                contentLength: fileContent.length
            });
            
            const response = await fetch(url, {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify({
                    agent_id: agentId,
                    docs_id: selectedFile.docs_id,
                    content: fileContent,
                }),
            });

            console.log("Save response:", response.status, response.statusText);

            if (response.ok) {
                const result = await response.json();
                console.log("Save result:", result);
                setOriginalContent(fileContent);
                setHasUnsavedChanges(false);
                toast.success("File saved successfully");
                return true;
            } else {
                const errorText = await response.text();
                console.error("Failed to save file:", response.status, errorText);
                toast.error(`Failed to save file: ${response.status}`);
                return false;
            }
        } catch (error) {
            console.error("Failed to save file", error);
            toast.error("Failed to save file: " + (error as Error).message);
            return false;
        } finally {
            setIsSavingFile(false);
        }
    };

    const handleFileSelect = (file: AgentFile) => {
        if (hasUnsavedChanges && selectedFile && selectedFile.id !== file.id) {
            setPendingFile(file);
            setShowUnsavedModal(true);
        } else {
            loadFileContent(file);
        }
    };

    const handleDiscardChanges = () => {
        setShowUnsavedModal(false);
        if (pendingFile) {
            loadFileContent(pendingFile);
            setPendingFile(null);
        }
    };

    const handleSaveAndContinue = async () => {
        const saved = await saveFileContent();
        if (saved) {
            setShowUnsavedModal(false);
            if (pendingFile) {
                loadFileContent(pendingFile);
                setPendingFile(null);
            }
        } else {
            // Keep the modal open if save failed
            toast.error("Could not switch files due to save failure");
        }
    };

    const handleCancelSwitch = () => {
        setShowUnsavedModal(false);
        setPendingFile(null);
    };

    const toggleFolder = (folderPath: string) => {
        setExpandedFolders(prev => {
            const next = new Set(prev);
            if (next.has(folderPath)) {
                next.delete(folderPath);
            } else {
                next.add(folderPath);
            }
            return next;
        });
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

    const renderCompactFileTree = (node: any, path = "", depth = 0) => {
        return Object.entries(node).map(([key, value]: [string, any]) => {
            const currentPath = path + key;
            
            if (value.docs_id) {
                // It's a file
                const isSelected = selectedFile?.id === value.id;
                const fileHasChanges = isSelected && hasUnsavedChanges;
                return (
                    <div
                        key={value.id}
                        className={`flex items-center gap-2 px-2 py-1 cursor-pointer hover:bg-muted rounded text-sm ${
                            isSelected ? "bg-primary/10 text-primary" : ""
                        }`}
                        style={{ paddingLeft: `${depth * 12 + 8}px` }}
                        onClick={() => handleFileSelect(value)}
                    >
                        <FileText className="h-3 w-3 shrink-0" />
                        <span className="truncate">{key}</span>
                        {fileHasChanges && (
                            <span className="h-2 w-2 rounded-full bg-red-500 shrink-0" title="Unsaved changes" />
                        )}
                    </div>
                );
            } else {
                // It's a folder
                const isExpanded = expandedFolders.has(currentPath);
                return (
                    <div key={currentPath}>
                        <div
                            className="flex items-center gap-1 px-2 py-1 cursor-pointer hover:bg-muted rounded text-sm"
                            style={{ paddingLeft: `${depth * 12 + 8}px` }}
                            onClick={() => toggleFolder(currentPath)}
                        >
                            {isExpanded ? (
                                <ChevronDown className="h-3 w-3 shrink-0" />
                            ) : (
                                <ChevronRight className="h-3 w-3 shrink-0" />
                            )}
                            <span className="font-medium">{key}</span>
                        </div>
                        {isExpanded && renderCompactFileTree(value, currentPath + "/", depth + 1)}
                    </div>
                );
            }
        });
    };

    const getLanguageExtension = (fileName: string) => {
        const ext = fileName.split('.').pop()?.toLowerCase();
        switch (ext) {
            case 'js':
            case 'jsx':
                return [javascript({ jsx: true })];
            case 'ts':
            case 'tsx':
                return [javascript({ jsx: true, typescript: true })];
            case 'py':
                return [python()];
            case 'html':
            case 'htm':
                return [html()];
            case 'md':
            case 'markdown':
                return [codeMirrorMarkdown()];
            default:
                return [];
        }
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
                <div className="max-w-full mx-auto w-full">
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
                        <button
                            className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                                activeTab === "editor"
                                    ? "border-primary text-primary"
                                    : "border-transparent text-muted-foreground hover:text-foreground"
                            }`}
                            onClick={() => setActiveTab("editor")}
                        >
                            Editor
                        </button>
                    </div>
                </div>
            </div>

            {/* Scrollable Content */}
            <div className="flex-1 overflow-y-auto min-h-0 bg-background">
                <div className={`${activeTab === "editor" ? "max-w-full h-full" : "max-w-4xl"} mx-auto w-full ${activeTab === "editor" ? "" : "p-6 pb-12"}`}>
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

                    {activeTab === "editor" && (
                        <div className="flex h-full p-4 gap-4">
                            {/* Left sidebar - File tree */}
                            <div className="w-64 border border-border bg-card overflow-y-auto rounded-lg">
                                <div className="p-3 border-b border-border">
                                    <h3 className="text-sm font-semibold">Files</h3>
                                </div>
                                <div className="py-2">
                                    {loadingFiles ? (
                                        <div className="flex items-center justify-center p-4">
                                            <Loader2 className="h-4 w-4 animate-spin" />
                                        </div>
                                    ) : agentFiles.length > 0 ? (
                                        renderCompactFileTree(buildFileTree(agentFiles))
                                    ) : (
                                        <div className="text-center text-muted-foreground text-xs p-4">
                                            No files
                                        </div>
                                    )}
                                </div>
                            </div>

                            {/* Right panel - Editor */}
                            <div className="flex-1 flex flex-col border border-border rounded-lg overflow-hidden bg-card">
                                {selectedFile ? (
                                    <>
                                        <div className="px-4 py-2 border-b border-border bg-card">
                                            <div className="flex items-center justify-between">
                                                <div className="flex items-center gap-3">
                                                    <FileText className="h-4 w-4" />
                                                    <div className="flex items-center gap-2">
                                                        <span className="text-sm font-medium">{selectedFile.file_name}</span>
                                                        {hasUnsavedChanges && (
                                                            <span className="h-2 w-2 rounded-full bg-red-500" title="Unsaved changes" />
                                                        )}
                                                    </div>
                                                    {isMarkdownFile(selectedFile.file_name) && (
                                                        <Button
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() => setIsRawMode(!isRawMode)}
                                                            className="h-7 gap-1"
                                                        >
                                                            <Code2 className="h-3 w-3" />
                                                            <span className="text-xs">{isRawMode ? 'Preview' : 'Raw'}</span>
                                                        </Button>
                                                    )}
                                                </div>
                                                <Button
                                                    variant="default"
                                                    size="sm"
                                                    onClick={saveFileContent}
                                                    disabled={!hasUnsavedChanges || isSavingFile}
                                                    className="h-7 gap-1"
                                                >
                                                    {isSavingFile ? (
                                                        <Loader2 className="h-3 w-3 animate-spin" />
                                                    ) : (
                                                        <Save className="h-3 w-3" />
                                                    )}
                                                    <span className="text-xs">Save</span>
                                                </Button>
                                            </div>
                                        </div>
                                        <div className="flex-1 overflow-hidden">
                                            {loadingContent ? (
                                                <div className="flex items-center justify-center h-full">
                                                    <Loader2 className="h-6 w-6 animate-spin" />
                                                </div>
                                            ) : isMarkdownFile(selectedFile.file_name) && !isRawMode ? (
                                                <MarkdownEditor
                                                    key={selectedFile.docs_id}
                                                    content={fileContent}
                                                    isDark={isDarkMode}
                                                    onChange={handleFileContentChange}
                                                />
                                            ) : (
                                                <CodeMirror
                                                    key={selectedFile.docs_id}
                                                    value={fileContent}
                                                    height="100%"
                                                    theme={isDarkMode ? githubDark : githubLight}
                                                    extensions={getLanguageExtension(selectedFile.file_name)}
                                                    onChange={(value) => handleFileContentChange(value)}
                                                    className="h-full"
                                                />
                                            )}
                                        </div>
                                    </>
                                ) : (
                                    <div className="flex items-center justify-center h-full text-muted-foreground">
                                        <div className="text-center">
                                            <FileText className="h-12 w-12 mx-auto mb-2 opacity-50" />
                                            <p className="text-sm">Select a file to edit</p>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {/* Unsaved Changes Modal */}
            {showUnsavedModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-card border border-border rounded-lg p-6 max-w-md w-full mx-4 shadow-lg">
                        <div className="flex items-start gap-4">
                            <div className="flex-shrink-0">
                                <AlertCircle className="h-6 w-6 text-yellow-500" />
                            </div>
                            <div className="flex-1">
                                <h3 className="text-lg font-semibold mb-2">Unsaved Changes</h3>
                                <p className="text-sm text-muted-foreground mb-4">
                                    You have unsaved changes in "{selectedFile?.file_name}". 
                                    Do you want to save them before switching files?
                                </p>
                                <div className="flex gap-2 justify-end">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={handleCancelSwitch}
                                    >
                                        Cancel
                                    </Button>
                                    <Button
                                        variant="destructive"
                                        size="sm"
                                        onClick={handleDiscardChanges}
                                    >
                                        Discard
                                    </Button>
                                    <Button
                                        variant="default"
                                        size="sm"
                                        onClick={handleSaveAndContinue}
                                    >
                                        Save & Continue
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

