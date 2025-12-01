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
    // Allow proceeding without API key (user can skip)
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
          <div>
            <Label htmlFor="api-key">OpenAI API Key (optional)</Label>
            <Input
              id="api-key"
              type="password"
              placeholder="sk-..."
              value={data.OPENAI_API_KEY}
              onChange={(e) => onboardingActions.setOpenaiApiKey(e.target.value)}
              className={validationError ? 'border-red-500' : ''}
            />
            {validationError && (
              <p className="text-sm text-red-600 mt-1">{validationError}</p>
            )}
          </div>


          <div>
            <Label htmlFor="api-key">Gemini API Key (optional)</Label>
            <Input
              id="api-key"
              type="password"
              placeholder="sk-..."
              value={data.GEMINI_API_KEY}
              onChange={(e) => onboardingActions.setGeminiApiKey(e.target.value)}
              className={validationError ? 'border-red-500' : ''}
            />
            {validationError && (
              <p className="text-sm text-red-600 mt-1">{validationError}</p>
            )}
          </div>

          <div>
            <Label htmlFor="api-url">Claude API Key (optional)</Label>
            <div className="flex gap-2">
              <Input
                id="api-url"
                type="url"
                placeholder="sk-..."
                value={data.CLAUDE_API_KEY}
                onChange={(e) => onboardingActions.setClaudeApiKey(e.target.value)}
              />
            </div>

            <p className="text-sm text-gray-500 mt-1">
              Provide the endpoint for either LiteLLM or OpenAI.
            </p>
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