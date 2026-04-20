import { useState } from 'react';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
    DialogDescription,
} from '@/components/ui/dialog';
import { addModel } from '@/store/chat';
import { AddModelRequest } from '../../proto/chatservice';

interface AddModelDialogProps {
    providerName: string;
    isOpen: boolean;
    onOpenChange: (open: boolean) => void;
}

interface AddModelFormState {
    model_id: string;
    model_name: string;
    input_token_cost: number;
    output_token_cost: number;
    cached_token_cost: number;
    is_embedding_model: boolean;
}

export function AddModelDialog({
    providerName,
    isOpen,
    onOpenChange,
}: AddModelDialogProps) {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [formData, setFormData] = useState<AddModelFormState>({
        model_id: '',
        model_name: '',
        input_token_cost: 0,
        output_token_cost: 0,
        cached_token_cost: 0,
        is_embedding_model: false,
    });

    const handleChange = (
        e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
    ) => {
        const { name, value, type } = e.target;

        if (type === 'checkbox') {
            const checked = (e.target as HTMLInputElement).checked;
            setFormData((prev) => ({ ...prev, [name]: checked }));
        } else if (type === 'number') {
            setFormData((prev) => ({ ...prev, [name]: parseFloat(value) || 0 }));
        } else {
            setFormData((prev) => ({ ...prev, [name]: value }));
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!formData.model_id) return;

        setIsSubmitting(true);
        try {
            await addModel(new AddModelRequest({
                provider_name: providerName,
                ...formData
            }));
            onOpenChange(false);
            // Reset form
            setFormData({
                model_id: '',
                model_name: '',
                input_token_cost: 0,
                output_token_cost: 0,
                cached_token_cost: 0,
                is_embedding_model: false,
            });
        } catch (error) {
            console.error(error);
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
                    <DialogTitle>Add New Model</DialogTitle>
                    <DialogDescription>
                        Add a new model for {providerName.charAt(0).toUpperCase() + providerName.slice(1)}.
                    </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4 py-4">
                    <div className="space-y-2">
                        <label htmlFor="model_id" className={labelClass}>
                            Model ID <span className="text-destructive">*</span>
                        </label>
                        <input
                            id="model_id"
                            name="model_id"
                            value={formData.model_id}
                            onChange={handleChange}
                            placeholder="e.g., gpt-4o"
                            required
                            className={inputClass}
                        />
                    </div>

                    <div className="space-y-2">
                        <label htmlFor="model_name" className={labelClass}>
                            Model Name
                        </label>
                        <input
                            id="model_name"
                            name="model_name"
                            value={formData.model_name}
                            onChange={handleChange}
                            placeholder="e.g., GPT-4o"
                            className={inputClass}
                        />
                    </div>

                    <div className="space-y-2 flex items-center gap-2 mt-4">
                        <input
                            type="checkbox"
                            id="is_embedding_model"
                            name="is_embedding_model"
                            checked={formData.is_embedding_model}
                            onChange={handleChange}
                            className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
                        />
                        <label htmlFor="is_embedding_model" className={labelClass}>
                            Is this an embedding model?
                        </label>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <label htmlFor="input_token_cost" className={labelClass}>
                                Input Token Cost
                            </label>
                            <input
                                id="input_token_cost"
                                name="input_token_cost"
                                type="number"
                                step="0.000001"
                                min="0"
                                value={formData.input_token_cost}
                                onChange={handleChange}
                                className={inputClass}
                            />
                        </div>
                        <div className="space-y-2">
                            <label htmlFor="output_token_cost" className={labelClass}>
                                Output Token Cost
                            </label>
                            <input
                                id="output_token_cost"
                                name="output_token_cost"
                                type="number"
                                step="0.000001"
                                min="0"
                                value={formData.output_token_cost}
                                onChange={handleChange}
                                className={inputClass}
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label htmlFor="cached_token_cost" className={labelClass}>
                            Cached Token Cost
                        </label>
                        <input
                            id="cached_token_cost"
                            name="cached_token_cost"
                            type="number"
                            step="0.000001"
                            min="0"
                            value={formData.cached_token_cost}
                            onChange={handleChange}
                            className={inputClass}
                        />
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
                            disabled={isSubmitting || !formData.model_id}
                            className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground shadow hover:bg-primary/90 h-9 px-4 py-2"
                        >
                            {isSubmitting ? 'Adding...' : 'Add Model'}
                        </button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
