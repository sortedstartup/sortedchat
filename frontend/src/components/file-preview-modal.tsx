import { Button } from "@/components/ui/button";
import { FileText, XCircle, Download, Code2, FileType } from "lucide-react";
import { useEffect, useState } from "react";

export type FilePreviewType = 'html' | 'pdf' | 'text' | 'json' | 'markdown';

interface FilePreviewModalProps {
    isOpen: boolean;
    onClose: () => void;
    fileName: string;
    fileContent: string;
    fileType: FilePreviewType;
    enableScripts?: boolean;
    onScriptsToggle?: (enabled: boolean) => void;
}

export function FilePreviewModal({
    isOpen,
    onClose,
    fileName,
    fileContent,
    fileType,
    enableScripts = true,
    onScriptsToggle,
}: FilePreviewModalProps) {
    const [pdfUrl, setPdfUrl] = useState('');


    useEffect(() => {
        if (isOpen && fileType === 'pdf') {
            const pdfBlob = new Blob([fileContent], { type: 'application/pdf' });
            const url = URL.createObjectURL(pdfBlob);
            setPdfUrl(url);

            return () => {
                URL.revokeObjectURL(url);
                setPdfUrl('');
            };
        }
    }, [isOpen, fileType, fileContent]);

    const [localEnableScripts, setLocalEnableScripts] = useState(enableScripts);

    if (!isOpen) return null;

    const handleScriptsToggle = (checked: boolean) => {
        setLocalEnableScripts(checked);
        onScriptsToggle?.(checked);
    };

    const downloadFile = () => {
        const mimeTypes: Record<FilePreviewType, string> = {
            html: 'text/html',
            pdf: 'application/pdf',
            text: 'text/plain',
            json: 'application/json',
            markdown: 'text/markdown',
        };

        const blob = new Blob([fileContent], { type: mimeTypes[fileType] });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = fileName;
        a.click();
        URL.revokeObjectURL(url);
    };

    const getFileIcon = () => {
        switch (fileType) {
            case 'html':
                return <FileText className="w-5 h-5 text-blue-600 dark:text-blue-400" />;
            case 'pdf':
                return <FileType className="w-5 h-5 text-red-600 dark:text-red-400" />;
            case 'json':
                return <Code2 className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />;
            case 'markdown':
                return <FileText className="w-5 h-5 text-purple-600 dark:text-purple-400" />;
            default:
                return <FileText className="w-5 h-5 text-gray-600 dark:text-gray-400" />;
        }
    };

    const getFileTypeLabel = () => {
        return fileType.toUpperCase();
    };

    const renderContent = () => {
    switch (fileType) {
        case 'html': {
            const sandboxPermissions = localEnableScripts
                ? "allow-same-origin allow-scripts"
                : "allow-same-origin";

            return (
                <>
                    <div className="flex items-center justify-between px-4 py-2 bg-yellow-50 dark:bg-yellow-950/30 border-b border-yellow-200 dark:border-yellow-800">
                        <span className="text-xs text-yellow-800 dark:text-yellow-200">
                            Preview is sandboxed for security
                        </span>
                        {onScriptsToggle && (
                            <label className="flex items-center gap-2 text-xs cursor-pointer">
                                <input
                                    type="checkbox"
                                    checked={localEnableScripts}
                                    onChange={(e) => handleScriptsToggle(e.target.checked)}
                                    className="rounded"
                                />
                                <span className="text-yellow-800 dark:text-yellow-200">Enable JavaScript</span>
                            </label>
                        )}
                    </div>
                    <iframe
                        key={localEnableScripts ? 'with-scripts' : 'no-scripts'}
                        srcDoc={fileContent}
                        sandbox={sandboxPermissions}
                        className="w-full h-full bg-white"
                        title={`Preview: ${fileName}`}
                    />
                </>
            );
        }

        case 'pdf': {
            return (
                <iframe
                    src={pdfUrl}
                    className="w-full h-full bg-white"
                    title={`Preview: ${fileName}`}
                />
            );
        }

        case 'json': {
            try {
                const formatted = JSON.stringify(JSON.parse(fileContent), null, 2);
                return (
                    <pre className="w-full h-full overflow-auto p-4 bg-muted/20 text-xs font-mono">
                        <code>{formatted}</code>
                    </pre>
                );
            } catch {
                return (
                    <pre className="w-full h-full overflow-auto p-4 bg-muted/20 text-xs font-mono">
                        <code>{fileContent}</code>
                    </pre>
                );
            }
        }

        case 'markdown':
        case 'text':
        default: {
            return (
                <pre className="w-full h-full overflow-auto p-4 bg-muted/20 text-sm font-mono whitespace-pre-wrap break-words">
                    <code>{fileContent}</code>
                </pre>
            );
        }
    }
};

    return (
        <div
            className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4"
            onClick={onClose}
        >
            <div
                className="bg-card rounded-lg shadow-xl w-full h-full max-w-7xl max-h-[90vh] flex flex-col"
                onClick={(e) => e.stopPropagation()}
            >
                {/* Modal Header */}
                <div className="flex items-center justify-between p-4 border-b border-border flex-shrink-0">
                    <div className="flex items-center gap-2">
                        {getFileIcon()}
                        <span className="text-lg font-semibold truncate max-w-md" title={fileName}>
                            {fileName}
                        </span>
                        <span className="text-xs bg-blue-100 dark:bg-blue-900/40 text-blue-800 dark:text-blue-200 px-2 py-0.5 rounded">
                            {getFileTypeLabel()}
                        </span>
                    </div>
                    <div className="flex items-center gap-2">
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={downloadFile}
                            className="h-8 gap-1"
                            title="Download file"
                        >
                            <Download className="w-4 h-4" />
                            <span className="text-xs">Download</span>
                        </Button>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={onClose}
                            className="h-8"
                            title="Close"
                        >
                            <XCircle className="w-4 h-4" />
                        </Button>
                    </div>
                </div>

                {/* Modal Content */}
                <div className="flex-1 overflow-hidden">
                    {renderContent()}
                </div>
            </div>
        </div>
    );
}

