import { useState } from 'react';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
    DialogDescription,
} from '@/components/ui/dialog';
import { SetProviderSetting } from '@/store/setting';
import { ProviderSettings } from '../../proto/chatservice';
import { toast } from 'sonner';

interface AddProviderDialogProps {
    isOpen: boolean;
    onOpenChange: (open: boolean) => void;
}

export function AddProviderDialog({
    isOpen,
    onOpenChange,
}: AddProviderDialogProps) {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [formData, setFormData] = useState({
        provider_name: '',
        api_url: '',
        api_key: '',
        is_enabled: true,
    });

    const handleChange = (
        e: React.ChangeEvent<HTMLInputElement>
    ) => {
        const { name, value, type, checked } = e.target;
        setFormData((prev) => ({
            ...prev,
            [name]: type === 'checkbox' ? checked : value,
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!formData.provider_name.trim()) {
            toast.error('Provider Name is required');
            return;
        }

        setIsSubmitting(true);
        try {
            const settings = new ProviderSettings({
                api_url: formData.api_url,
                api_key: formData.api_key,
                is_enabled: formData.is_enabled,
            });

            // Automatically converts provider name to lower case for consistency
            const providerName = formData.provider_name.trim().toLowerCase();
            await SetProviderSetting(providerName, settings);

            toast.success('Provider added successfully');
            onOpenChange(false);
            // Reset form
            setFormData({
                provider_name: '',
                api_url: '',
                api_key: '',
                is_enabled: true,
            });
        } catch (error) {
            console.error(error);
            toast.error('Failed to add provider');
        } finally {
            setIsSubmitting(false);
        }
    };

    const inputClass = "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50";
    const labelClass = "text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 text-foreground";

    return (
        <Dialog open={isOpen} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>Add New Provider</DialogTitle>
                    <DialogDescription>
                        Add an OpenAI-compatible provider.
                    </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4 py-4">
                    <div className="space-y-2">
                        <label htmlFor="provider_name" className={labelClass}>
                            Provider Name <span className="text-destructive">*</span>
                        </label>
                        <input
                            id="provider_name"
                            name="provider_name"
                            value={formData.provider_name}
                            onChange={handleChange}
                            placeholder="e.g., groq"
                            required
                            className={inputClass}
                        />
                    </div>

                    <div className="space-y-2">
                        <label htmlFor="api_url" className={labelClass}>
                            API URL
                        </label>
                        <input
                            id="api_url"
                            name="api_url"
                            value={formData.api_url}
                            onChange={handleChange}
                            placeholder="e.g., https://api.groq.com/openai/v1/chat/completions"
                            className={inputClass}
                        />
                        <p className="text-[10px] text-muted-foreground mt-1">
                            Note: Use OpenAI chat completions API compatible endpoint only
                        </p>
                    </div>

                    <div className="space-y-2">
                        <label htmlFor="api_key" className={labelClass}>
                            API Key
                        </label>
                        <input
                            id="api_key"
                            name="api_key"
                            type="password"
                            value={formData.api_key}
                            onChange={handleChange}
                            placeholder="Enter API Key"
                            className={inputClass}
                        />
                    </div>

                    <div className="space-y-2 flex items-center gap-2 mt-4">
                        <input
                            type="checkbox"
                            id="is_enabled"
                            name="is_enabled"
                            checked={formData.is_enabled}
                            onChange={handleChange}
                            className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
                        />
                        <label htmlFor="is_enabled" className={labelClass}>
                            Enable this provider
                        </label>
                    </div>

                    <DialogFooter className="pt-4">
                        <button
                            type="button"
                            className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground h-9 px-4 py-2"
                            onClick={() => onOpenChange(false)}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={isSubmitting || !formData.provider_name.trim()}
                            className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground shadow hover:bg-primary/90 h-9 px-4 py-2"
                        >
                            {isSubmitting ? 'Adding...' : 'Add Provider'}
                        </button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
