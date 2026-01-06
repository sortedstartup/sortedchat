import { useStore } from '@nanostores/react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { $onboardingStep } from '@/store/setting';
import { StepWelcome } from '@/components/onboarding/steps/StepWelcome';
import { StepEmbeddings } from '@/components/onboarding/steps/StepEmbeddings';
import { StepApi } from '@/components/onboarding/steps/StepApi';

const steps = [
  'Welcome',
  'Local Models Setup',
  'Remote Providers Setup'
];

export function OnboardingWizard() {
  const step = useStore($onboardingStep);
  
  const renderStep = () => {
    switch (step) {
      case 0:
        return <StepWelcome />;
      case 1:
        return <StepEmbeddings />;
      case 2:
        return <StepApi />;
      default:
        return <StepWelcome />;
    }
  };
  
  // First step (welcome) renders full screen
  if (step === 0) {
    return <StepWelcome />;
  }
  
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center p-4">
      <div className="w-full max-w-4xl">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold mb-2">Setup SortedChat</h1>
          <p className="text-gray-600 dark:text-gray-400">Configure your AI providers</p>
        </div>
        
        <Card>
          <CardHeader>
            <CardTitle className="text-center">
              {steps[step]} ({step + 1}/3)
            </CardTitle>
          </CardHeader>
          
          <CardContent className="max-h-[70vh] overflow-y-auto">
            {renderStep()}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
