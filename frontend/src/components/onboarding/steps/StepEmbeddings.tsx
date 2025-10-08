import { useState } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { $onboardingData, onboardingActions } from '@/store/setting';
import { CheckCircle, XCircle, Loader2 } from 'lucide-react';
import { ConnectionType } from '../../../../proto/chatservice';

export function StepEmbeddings() {
  const data = useStore($onboardingData);
  const [isValidating, setIsValidating] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [validationError, setValidationError] = useState<string>('');
  const [testResult, setTestResult] = useState<{success: boolean, message: string} | null>(null);
  
  const handleTest = async () => {
    if (!data.OLLAMA_URL.trim()) {
      setValidationError('Please enter an Ollama URL to test');
      return;
    }
    
    setIsTesting(true);
    setTestResult(null);
    setValidationError('');
    
    try { 
      const result = await onboardingActions.testConnection(data.OLLAMA_URL, ConnectionType.OLLAMA);
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
    setValidationError('');
    setIsValidating(true);
    
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
  
  const handleSkip = async () => {
    // Skip this step and complete onboarding
    setIsValidating(true);
    try {
      await onboardingActions.completeOnboarding();
    } catch (error) {
      console.error('Failed to complete onboarding:', error);
      setValidationError('Failed to save settings. Please try again.');
    } finally {
      setIsValidating(false);
    }
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div>
          <Label htmlFor="ollama-url">Ollama Server URL (optional)</Label>
          <div className="flex gap-2">
            <Input
              id="ollama-url"
              type="url"
              placeholder="http://localhost:11434"
              value={data.OLLAMA_URL}
              onChange={(e) => onboardingActions.setOllamaUrl(e.target.value)}
              className={validationError ? 'border-red-500' : ''}
            />
            <Button
              variant="outline"
              onClick={handleTest}
              disabled={isTesting || !data.OLLAMA_URL.trim()}
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
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleBack}
          >
            Back
          </Button>
          <Button
            variant="ghost"
            onClick={handleSkip}
            disabled={isValidating}
          >
            Skip & Finish
          </Button>
        </div>
        
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