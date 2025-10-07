import { useStore } from '@nanostores/react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { $onboardingStep } from '@/store/setting';
import { StepApi } from './steps/StepApi';
import { StepEmbeddings } from './steps/StepEmbeddings';

const steps = [
  'OpenAI API Setup',
  'Ollama Setup'
];

export function OnboardingWizard() {
  const step = useStore($onboardingStep);
  
  const renderStep = () => {
    switch (step) {
      case 0:
        return <StepApi />;
      case 1:
        return <StepEmbeddings />;
      default:
        return <StepApi />;
    }
  };
  
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="w-full max-w-lg">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold mb-2">Welcome to SortedChat</h1>
          <p className="text-gray-600">Let's set up your AI providers</p>
        </div>
        
        <Card>
          <CardHeader>
            <CardTitle className="text-center">
              {steps[step]} ({step + 1}/2)
            </CardTitle>
          </CardHeader>
          
          <CardContent>
            {renderStep()}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
