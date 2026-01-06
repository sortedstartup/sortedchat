import { useEffect } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { $onboardingData, onboardingActions } from '@/store/setting';
import { $modelsByProvider, $isLoadingModels, ListLLMModels } from '@/store/inference';
import { ModelCard } from '@/components/model-card';

export function StepEmbeddings() {
  const data = useStore($onboardingData);
  const modelsByProvider = useStore($modelsByProvider);
  const isLoadingModels = useStore($isLoadingModels);

  // Load models on mount
  useEffect(() => {
    ListLLMModels();
  }, []);

  const handleNext = () => {
    onboardingActions.nextStep();
  };

  const handleBack = () => {
    onboardingActions.prevStep();
  };

  // Get local models
  const localModels = modelsByProvider['local'] || [];

  return (
    <div className="space-y-6">
      {/* Ollama URL Setting */}
      <div className="space-y-4">
        <div>
          <Label htmlFor="ollama-url" className="text-base font-semibold">Ollama Server URL</Label>
          <Input
            id="ollama-url"
            type="url"
            placeholder="http://localhost:8081"
            value={data.OLLAMA_URL}
            onChange={(e) => onboardingActions.setOllamaUrl(e.target.value)}
            className="mt-2"
          />
          <p className="text-sm text-muted-foreground mt-1">
            Connect to your local Ollama server for private AI models
          </p>
        </div>
      </div>

      {/* Models Section */}
      <div className="border-t pt-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-lg font-semibold">Local Models (Llama.cpp)</h3>
            <p className="text-sm text-muted-foreground">
              {localModels.length} models available
            </p>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => ListLLMModels()}
            disabled={isLoadingModels}
          >
            {isLoadingModels ? 'Refreshing...' : 'Refresh'}
          </Button>
        </div>

        {/* Models Grid */}
        {isLoadingModels && localModels.length === 0 ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-muted-foreground">Loading models...</p>
            </div>
          </div>
        ) : localModels.length === 0 ? (
          <div className="text-center py-12 border rounded-lg bg-muted/30">
            <p className="text-muted-foreground">No local models found. You can skip this step and add models later.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 max-h-96 overflow-y-auto pr-2">
            {localModels.map((model) => (
              <ModelCard key={model.id} model={model} isLocal={true} />
            ))}
          </div>
        )}
      </div>

      <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
        <p className="text-sm text-blue-800 dark:text-blue-200">
          <strong>Note:</strong> Download models to run AI completely offline. Using <code className="bg-blue-100 dark:bg-blue-900 px-1 rounded">nomic-embed-text</code> for embeddings.
        </p>
      </div>

      <div className="flex justify-between pt-4">
        <Button
          variant="outline"
          onClick={handleBack}
        >
          Back
        </Button>

        <Button
          onClick={handleNext}
        >
          Next
        </Button>
      </div>
    </div>
  );
}	