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
          <div className='p-2'>
            <Label htmlFor="api-key" className='p-2'>OpenAI API Key (optional)</Label>
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


          <div className='p-2'>
            <Label htmlFor="api-key" className='p-2'>Gemini API Key (optional)</Label>
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

          <div className='p-2'>
            <Label htmlFor="api-key" className='p-2'>Claude API Key (optional)</Label>
            {/* <div className="flex gap-2"> */}
            <Input
              id="api-key"
              type="password"
              placeholder="sk-..."
              value={data.CLAUDE_API_KEY}
              onChange={(e) => onboardingActions.setClaudeApiKey(e.target.value)}
            />
          </div>

          <div className='p-2'>
            <Label htmlFor="claude-api-url" className='p-2'>Claude API URL (optional)</Label>
            <Input
              id="claude-api-url"
              type="url"
              value={data.CLAUDE_API_URL}
              onChange={(e) => onboardingActions.setClaudeApiUrl(e.target.value)}
              className={validationError ? 'border-red-500' : ''}
            />
          </div>

          <div className='p-2'>
            <Label htmlFor="openai-api-url" className='p-2'>Openai API URL (optional)</Label>
            <Input
              id="openai-api-url"
              type="url"
              value={data.OPENAI_API_URL}
              onChange={(e) => onboardingActions.setOpenaiApiUrl(e.target.value)}
              className={validationError ? 'border-red-500' : ''}
            />
          </div>

          <div className='p-2'>
            <Label htmlFor="gemini-api-url" className='p-2'>Gemini API URL (optional)</Label>
            <Input
              id="gemini-api-url"
              type="url"
              value={data.GEMINI_API_URL}
              onChange={(e) => onboardingActions.setGeminiApiUrl(e.target.value)}
              className={validationError ? 'border-red-500' : ''}
            />
          </div>
          {/* </div> */}
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