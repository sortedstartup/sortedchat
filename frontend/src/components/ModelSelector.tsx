
import { useState, useMemo, useEffect, useRef } from 'react';
import { Check, ChevronsUpDown, Search, Settings, Pin, PinOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
    DropdownMenuLabel,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { ModelListInfo } from "../../proto/chatservice";
import { cn } from "@/lib/utils";
import { useStore } from '@nanostores/react';
import { $pinnedModels, togglePinnedModel } from '../store/inference';
import { $providerSettings } from '../store/setting';
import { useNavigate } from 'react-router-dom';
import { $availableModels } from '@/store/chat';

interface ModelSelectorProps {
    selectedModelId: string;
    onSelectModel: (modelId: string, provider: string) => void;
    className?: string;
}

export function ModelSelector({
    selectedModelId,
    onSelectModel,
    className,
}: ModelSelectorProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [isOpen, setIsOpen] = useState(false);
    const pinnedModels = useStore($pinnedModels);
    const providerSettings = useStore($providerSettings);
    const rawAvailableModels = useStore($availableModels);
    const navigate = useNavigate();
    const scrollRef = useRef<HTMLDivElement>(null);

    const models = useMemo(() => {
        return rawAvailableModels.filter(
            (model) =>
                !model.is_embedding_model &&
                (!model.is_downloadable || model.is_downloaded)
        );
    }, [rawAvailableModels]);

    const selectedModel = models.find((m) => m.id === selectedModelId);

    const filteredModels = useMemo(() => {
        return models.filter((model) =>
            model.label.toLowerCase().includes(searchQuery.toLowerCase())
        );
    }, [models, searchQuery]);

    const { unpinnedGroups, pinnedList } = useMemo(() => {
        const groups: Record<string, ModelListInfo[]> = {};
        const pinned: ModelListInfo[] = [];

        filteredModels.forEach((model) => {
            const provider = model.provider || "Other";

            // Check if provider is enabled
            const settings = providerSettings.get(provider);
            if (settings && !settings.is_enabled) return;

            if (pinnedModels.includes(model.id)) {
                pinned.push(model);
            } else {
                if (!groups[provider]) {
                    groups[provider] = [];
                }
                groups[provider].push(model);
            }
        });

        return { unpinnedGroups: groups, pinnedList: pinned };
    }, [filteredModels, pinnedModels, providerSettings]);

    useEffect(() => {
        if (isOpen) {
            // Use a longer timeout to ensure the DOM is fully rendered and animations have started
            const timer = setTimeout(() => {
                if (scrollRef.current) {
                    scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
                }
            }, 100);
            return () => clearTimeout(timer);
        }
    }, [isOpen]);

    return (
        <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
            <DropdownMenuTrigger asChild>
                <Button
                    variant="outline"
                    role="combobox"
                    aria-expanded={isOpen}
                    className={cn("w-fit justify-between text-xs", className)}
                    size="sm"
                >
                    <span className="truncate">
                        {selectedModel ? selectedModel.label : "Select model..."}
                    </span>
                    <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-[300px] p-0" align="start">
                <div
                    ref={scrollRef}
                    className="max-h-[300px] overflow-y-auto [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-muted [&::-webkit-scrollbar-thumb]:rounded-full"
                >
                    {Object.entries(unpinnedGroups).length === 0 && pinnedList.length === 0 && (
                        <div className="py-6 text-center text-sm text-muted-foreground">
                            No models found.
                        </div>
                    )}

                    {/* Unpinned Models Grouped by Provider */}
                    {Object.entries(unpinnedGroups).map(([provider, providerModels]) => (
                        <div key={provider}>
                            <DropdownMenuLabel className="flex items-center justify-between text-xs font-semibold text-muted-foreground bg-popover px-2 py-1.5 sticky top-0 z-10 border-b">
                                <div className="flex items-center gap-2">
                                    {provider}
                                </div>
                                <Settings
                                    className="h-3 w-3 cursor-pointer hover:text-foreground"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        navigate('/models');
                                        setIsOpen(false);
                                    }}
                                />
                            </DropdownMenuLabel>
                            {providerModels.map((model) => (
                                <DropdownMenuItem
                                    key={model.id}
                                    onSelect={() => {
                                        onSelectModel(model.id, model.provider);
                                        setIsOpen(false);
                                    }}
                                    className="flex items-center justify-between group"
                                >
                                    <span className="truncate flex-1">{model.label}</span>
                                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-6 w-6"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                togglePinnedModel(model.id);
                                            }}
                                        >
                                            <Pin className="h-3 w-3" />
                                        </Button>
                                    </div>
                                    {selectedModelId === model.id && (
                                        <Check className="h-4 w-4 ml-2" />
                                    )}
                                </DropdownMenuItem>
                            ))}
                        </div>
                    ))}

                    {/* Pinned Models Section */}
                    {pinnedList.length > 0 && (
                        <div>
                            <DropdownMenuLabel className="flex items-center justify-between text-xs font-semibold text-muted-foreground bg-popover px-2 py-1.5 sticky top-0 z-10 border-b border-t mt-2">
                                <div className="flex items-center gap-2">
                                    Pinned
                                </div>
                            </DropdownMenuLabel>
                            {pinnedList.map((model) => (
                                <DropdownMenuItem
                                    key={model.id}
                                    onSelect={() => {
                                        onSelectModel(model.id, model.provider);
                                        setIsOpen(false);
                                    }}
                                    className="flex items-center justify-between group"
                                >
                                    <span className="truncate flex-1">{model.label}</span>
                                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-6 w-6"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                togglePinnedModel(model.id);
                                            }}
                                        >
                                            <PinOff className="h-3 w-3" />
                                        </Button>
                                    </div>
                                    {selectedModelId === model.id && (
                                        <Check className="h-4 w-4 ml-2" />
                                    )}
                                </DropdownMenuItem>
                            ))}
                        </div>
                    )}
                </div>
                <div className="flex items-center border-t px-3 bg-popover">
                    <Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
                    <Input
                        placeholder="Search models..."
                        className="flex h-9 w-full rounded-md bg-transparent py-1 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50 border-none focus-visible:ring-0 shadow-none"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        onKeyDown={(e) => e.stopPropagation()}
                        autoFocus
                    />
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
