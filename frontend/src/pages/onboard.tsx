import { useStore } from '@nanostores/react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { $onboardingStep } from '@/store/setting';
import { StepWelcome } from '@/components/onboarding/steps/StepWelcome';
import { StepLocal } from '@/components/onboarding/steps/StepLocal';
import { StepRemote } from '@/components/onboarding/steps/StepRemote';

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
        return <StepLocal />;
      case 2:
        return <StepRemote />;
      default:
        return <StepWelcome />;
    }
  };

  // First step (welcome) renders full screen
  if (step === 0) {
    return <StepWelcome />;
  }

  return (
    <div className="h-screen bg-gray-50 dark:bg-gray-900 flex flex-col p-4">
      <div className="text-center mb-4 flex-shrink-0">
        <h1 className="text-2xl font-bold mb-2">Setup SortedChat</h1>
        <p className="text-gray-600 dark:text-gray-400">Configure your AI providers</p>
      </div>

      <Card className="flex-1 flex flex-col overflow-hidden w-full max-w-[1600px] mx-auto">
        <CardHeader className="flex-shrink-0 border-b">
          <CardTitle className="text-center">
            {steps[step]} ({step + 1}/3)
          </CardTitle>
        </CardHeader>

        <CardContent className="flex-1 overflow-hidden p-6">
          {renderStep()}
        </CardContent>
      </Card>
    </div>
  );
}
