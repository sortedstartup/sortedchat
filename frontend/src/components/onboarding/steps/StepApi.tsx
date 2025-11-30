import { useState } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { $onboardingData, onboardingActions } from '@/store/setting';
import { CheckCircle, XCircle, Loader2 } from 'lucide-react';
import { ConnectionType } from '../../../../proto/chatservice';

export function StepApi() {
  const data = useStore($onboardingData);
  const [isValidating, setIsValidating] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [validationError, setValidationError] = useState<string>('');
  const [testResult, setTestResult] = useState<{success: boolean, message: string} | null>(null);
  
  const handleTest = async () => {
    if (!data.OPENAI_API_URL.trim()) {
      setValidationError('Please enter an API URL to test');
      return;
    }
    
    setIsTesting(true);
    setTestResult(null);
    setValidationError('');
    
    try {
      const result = await onboardingActions.testConnection(data.OPENAI_API_URL, ConnectionType.OPENAI);
      setTestResult({
        success: result.success,
        message: result.message
      });
    } catch (error) {
      setTestResult({
        success: false,
        message: 'Connection test failed'
      });
    } finally {
      setIsTesting(false);
    }
  };
  
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
            <Label htmlFor="api-url">Custom API URL (optional)</Label>
            <div className="flex gap-2">
              <Input
                id="api-url"
                type="url"
                placeholder="https://api.openai.com/v1/chat/completions"
                value={data.OPENAI_API_URL}
                onChange={(e) => onboardingActions.setApiUrl(e.target.value)}
              />
              <Button
                variant="outline"
                onClick={handleTest}
                disabled={isTesting || !data.OPENAI_API_URL.trim()}
                className="min-w-[80px]"
              >
                {isTesting ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Test'}
              </Button>
            </div>
            
            {testResult && (
              <div className={`flex items-center gap-2 mt-2 text-sm ${testResult.success ? 'text-green-600' : 'text-red-600'}`}>
                {testResult.success ? <CheckCircle className="h-4 w-4" /> : <XCircle className="h-4 w-4" />}
                {testResult.message}
              </div>
            )}
            
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