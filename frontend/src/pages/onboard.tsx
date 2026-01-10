import { useStore } from '@nanostores/react';
import { Card, CardContent } from '@/components/ui/card';
import { $onboardingStep } from '@/store/setting';
import { StepWelcome } from '@/components/onboarding/steps/StepWelcome';
import { StepLocal } from '@/components/onboarding/steps/StepLocal';
import { StepRemote } from '@/components/onboarding/steps/StepRemote';

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
    <div className="h-screen bg-gray-50 dark:bg-gray-900 flex flex-col p-2 sm:p-3">
      <div className="text-center mb-2 sm:mb-3 flex-shrink-0">
        <h1 className="text-xl sm:text-2xl font-bold mb-1">Setup SortedChat</h1>
        <p className="text-sm text-gray-600 dark:text-gray-400">Configure your AI providers</p>
      </div>

      <Card className="flex-1 flex flex-col overflow-hidden w-full max-w-[1600px] mx-auto">
        <CardContent className="flex-1 overflow-hidden p-0">
          {renderStep()}
        </CardContent>
      </Card>
    </div>
  );
}
