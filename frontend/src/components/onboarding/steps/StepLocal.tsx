import { Button } from '@/components/ui/button';
import { onboardingActions } from '@/store/setting';
import { LocalProvider } from '@/components/providers/local-provider';

export function StepLocal() {
  const handleNext = () => {
    onboardingActions.nextStep();
  };

  const handleBack = () => {
    onboardingActions.prevStep();
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 overflow-hidden border rounded-lg bg-background mb-4">
        <LocalProvider />
      </div>

      <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-4">
        <p className="text-sm text-blue-800 dark:text-blue-200">
          <strong>Note:</strong> Download models to run AI completely offline. Using <code className="bg-blue-100 dark:bg-blue-900 px-1 rounded">nomic-embed-text</code> for embeddings.
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
        >
          Next
        </Button>
      </div>
    </div>
  );
}
