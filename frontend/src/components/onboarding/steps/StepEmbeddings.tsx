import { useState } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { $onboardingData, onboardingActions } from '@/store/setting';

export function StepEmbeddings() {
  const data = useStore($onboardingData);
  const [isValidating, setIsValidating] = useState(false);
  const [validationError, setValidationError] = useState<string>('');
  
  const handleNext = async () => {
    // Simple validation - just check if URL is filled
    if (!data.OLLAMA_URL.trim()) {
      setValidationError('Ollama URL is required');
      return;
    }
    
    setValidationError('');
    setIsValidating(true);
    
    // Complete onboarding using the settings store
    try {
      await onboardingActions.completeOnboarding();
      
    } catch (error) {
      console.error('Failed to complete onboarding:', error);
      setValidationError('Failed to save settings. Please try again.');
    } finally {
      setIsValidating(false);
    }
  };
  
  const handleBack = () => {
    onboardingActions.prevStep();
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div>
          <Label htmlFor="ollama-url">Ollama Server URL</Label>
          <Input
            id="ollama-url"
            type="url"
            placeholder="http://localhost:11434"
            value={data.OLLAMA_URL}
            onChange={(e) => onboardingActions.setOllamaUrl(e.target.value)}
            className={validationError ? 'border-red-500' : ''}
          />
          {validationError && (
            <p className="text-sm text-red-600 mt-1">{validationError}</p>
          )}
          <p className="text-sm text-gray-500 mt-1">
            URL where your Ollama server is running
          </p>
        </div>
      </div>
      
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <p className="text-sm text-blue-800 mb-2">
          <strong>Need Ollama?</strong> Download from ollama.ai and run: <code className="bg-blue-100 px-1 rounded">ollama pull nomic-embed-text</code>
        </p>
      </div>
      
      <div className="flex justify-between">
        <Button
          variant="outline"
          onClick={handleBack}
        >
          Back
        </Button>
        
        <Button
          onClick={handleNext}
          disabled={isValidating}
        >
          {isValidating ? 'Saving...' : 'Complete Setup'}
        </Button>
      </div>
    </div>
  );
}
