import { useState } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, Info } from 'lucide-react';
import { onboardingStore, onboardingActions, canProceed } from '@/stores/onboardingStore';
import { validateCurrentStep } from '@/lib/validators/onboarding';

export function StepEmbeddings() {
  const state = useStore(onboardingStore);
  const canProceedToNext = useStore(canProceed);
  const [isValidating, setIsValidating] = useState(false);
  
  const handleNext = async () => {
    setIsValidating(true);
    onboardingActions.setValidating(true);
    onboardingActions.clearValidationErrors();
    
    try {
      const errors = await validateCurrentStep(
        1,
        state.provider,
        state.OPENAI_API_KEY,
        state.OPENAI_API_URL,
        state.OLLAMA_URL
      );
      
      if (Object.keys(errors).length > 0) {
        Object.entries(errors).forEach(([field, error]) => {
          onboardingActions.setValidationError(field as any, error);
        });
      } else {
        onboardingActions.nextStep();
      }
    } catch (error) {
      onboardingActions.setValidationError('ollamaUrl', 'Validation failed. Please try again.');
    } finally {
      setIsValidating(false);
      onboardingActions.setValidating(false);
    }
  };
  
  const handleBack = () => {
    onboardingActions.prevStep();
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div>
          <Label htmlFor="ollama-url">Ollama Server URL *</Label>
          <Input
            id="ollama-url"
            type="url"
            placeholder="http://localhost:11434"
            value={state.OLLAMA_URL}
            onChange={(e) => onboardingActions.setOllamaUrl(e.target.value)}
            className={state.validationErrors.ollamaUrl ? 'border-red-500' : ''}
          />
          {state.validationErrors.ollamaUrl && (
            <p className="text-sm text-red-600 mt-1">{state.validationErrors.ollamaUrl}</p>
          )}
          <p className="text-sm text-gray-500 mt-1">
            URL where your Ollama server is running
          </p>
        </div>
      </div>
      
      <Alert>
        <Info className="h-4 w-4" />
        <AlertDescription>
          <div className="space-y-2">
            <p>
              Ollama is used for generating embeddings to power document search and retrieval.
            </p>
            <p>
              <strong>Don't have Ollama installed?</strong>
            </p>
            <ol className="list-decimal list-inside space-y-1 text-sm">
              <li>Download from <a href="https://ollama.ai" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">ollama.ai</a></li>
              <li>Install and start Ollama</li>
              <li>Run: <code className="bg-gray-100 px-1 rounded">ollama pull nomic-embed-text</code></li>
            </ol>
          </div>
        </AlertDescription>
      </Alert>
      
      <div className="flex justify-between">
        <Button
          variant="outline"
          onClick={handleBack}
        >
          Back
        </Button>
        
        <Button
          onClick={handleNext}
          disabled={!canProceedToNext || isValidating}
          className="min-w-[100px]"
        >
          {isValidating ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Validating...
            </>
          ) : (
            'Next'
          )}
        </Button>
      </div>
    </div>
  );
}
