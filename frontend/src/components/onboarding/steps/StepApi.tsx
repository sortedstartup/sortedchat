import { useState } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { $onboardingData, onboardingActions } from '@/store/setting';

export function StepApi() {
  const data = useStore($onboardingData);
  const [isValidating, setIsValidating] = useState(false);
  const [validationError, setValidationError] = useState<string>('');
  
  const handleNext = async () => {
    // Simple validation - require API key, optional API URL
    if (!data.OPENAI_API_KEY.trim()) {
      setValidationError('OpenAI API Key is required');
      return;
    }
    setValidationError('');
    setIsValidating(true);
    setTimeout(() => {
      onboardingActions.nextStep();
      setIsValidating(false);
    }, 300);
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div className="space-y-4">
          <div className="space-y-4">
            <div>
              <Label htmlFor="api-key">OpenAI API Key *</Label>
              <Input
                id="api-key"
                type="password"
                placeholder="sk-..."
                value={data.OPENAI_API_KEY}
                onChange={(e) => onboardingActions.setApiKey(e.target.value)}
                className={validationError ? 'border-red-500' : ''}
              />
              {validationError && (
                <p className="text-sm text-red-600 mt-1">{validationError}</p>
              )}
            </div>
            
            <div>
              <Label htmlFor="api-url">Custom API URL (optional)</Label>
              <Input
                id="api-url"
                type="url"
                placeholder="https://api.openai.com/v1"
                value={data.OPENAI_API_URL}
                onChange={(e) => onboardingActions.setApiUrl(e.target.value)}
              />
              <p className="text-sm text-gray-500 mt-1">
                Leave empty to use the default OpenAI API
              </p>
            </div>
          </div>
        </div>
      </div>
      
      <div className="flex justify-end">
        <Button
          onClick={handleNext}
          disabled={isValidating}
          className="min-w-[100px]"
        >
          {isValidating ? 'Validating...' : 'Next'}
        </Button>
      </div>
    </div>
  );
}
