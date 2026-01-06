import { useState, useMemo } from 'react';
import { Check, ChevronsUpDown, Search, Settings, Wrench, Eye } from "lucide-react";
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

interface ModelSelectorProps {
    models: ModelListInfo[];
    selectedModelId: string;
    onSelectModel: (modelId: string, provider: string) => void;
    className?: string;
}

export function ModelSelector({
    models,
    selectedModelId,
    onSelectModel,
    className,
}: ModelSelectorProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [isOpen, setIsOpen] = useState(false);

    const selectedModel = models.find((m) => m.id === selectedModelId);

    const filteredModels = useMemo(() => {
        return models.filter((model) =>
            model.label.toLowerCase().includes(searchQuery.toLowerCase())
        );
    }, [models, searchQuery]);

    const groupedModels = useMemo(() => {
        const groups: Record<string, ModelListInfo[]> = {};
        filteredModels.forEach((model) => {
            const provider = model.provider || "Other";
            if (!groups[provider]) {
                groups[provider] = [];
            }
            groups[provider].push(model);
        });
        return groups;
    }, [filteredModels]);

    return (
        <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
            <DropdownMenuTrigger asChild>
                <Button
                    variant="outline"
                    role="combobox"
                    aria-expanded={isOpen}
                    className={cn("w-[200px] justify-between text-xs", className)}
                    size="sm"
                >
                    <span className="truncate">
                        {selectedModel ? selectedModel.label : "Select model..."}
                    </span>
                    <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-[300px] p-0" align="start">
                <div className="flex items-center border-b px-3">
                    <Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
                    <Input
                        placeholder="Search models..."
                        className="flex h-9 w-full rounded-md bg-transparent py-1 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50 border-none focus-visible:ring-0 shadow-none"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        onKeyDown={(e) => e.stopPropagation()}
                    />
                </div>
                <div className="max-h-[300px] overflow-y-auto">
                    {Object.entries(groupedModels).length === 0 && (
                        <div className="py-6 text-center text-sm text-muted-foreground">
                            No models found.
                        </div>
                    )}
                    {Object.entries(groupedModels).map(([provider, providerModels]) => (
                        <div key={provider}>
                            <DropdownMenuLabel className="flex items-center justify-between text-xs font-semibold text-muted-foreground bg-popover px-2 py-1.5 sticky top-0 z-10 border-b">
                                <div className="flex items-center gap-2">
                                    {/* You might want to add provider icons here if available */}
                                    {provider}
                                </div>
                                <Settings className="h-3 w-3 cursor-pointer hover:text-foreground" />
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
                                    <span className="truncate">{model.label}</span>
                                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <Button variant="ghost" size="icon" className="h-6 w-6">
                                            <Wrench className="h-3 w-3" />
                                        </Button>
                                        <Button variant="ghost" size="icon" className="h-6 w-6">
                                            <Eye className="h-3 w-3" />
                                        </Button>
                                    </div>
                                    {selectedModelId === model.id && (
                                        <Check className="h-4 w-4 ml-2" />
                                    )}
                                </DropdownMenuItem>
                            ))}
                        </div>
                    ))}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
