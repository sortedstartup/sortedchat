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
    <div className="h-screen flex flex-col bg-background">
      {/* Scrollable Content Area */}
      <div className="flex-1 overflow-auto">
        <div className="p-3 sm:p-4 lg:p-6 max-w-7xl mx-auto">
          <div className="mb-3 sm:mb-4 text-center max-w-3xl mx-auto">
            <h2 className="text-lg sm:text-xl lg:text-2xl font-extrabold mb-1 sm:mb-2 bg-clip-text text-transparent bg-gradient-to-r from-blue-600 to-cyan-500">
              Download Local Models
            </h2>
            <p className="text-gray-600 dark:text-gray-400 text-xs sm:text-sm lg:text-base leading-relaxed">
              Power your privacy-first AI by downloading models in <span className="font-semibold text-blue-600 dark:text-blue-400">GGUF</span> format.
              We utilize <code className="px-1 py-0.5 rounded bg-gray-100 dark:bg-gray-800 font-mono text-xs">llama-server</code> to enable seamless local interactions.
            </p>
          </div>

          <div className="border rounded-2xl bg-background/50 backdrop-blur-sm shadow-inner mb-3 sm:mb-4 overflow-auto max-h-[35vh]">
            <LocalProvider />
          </div>

          <div className="bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-100 dark:border-indigo-900/50 rounded-xl p-2 sm:p-3 shadow-sm">
            <div className="flex items-center gap-2">
              <div className="flex-shrink-0 w-6 h-6 sm:w-8 sm:h-8 bg-indigo-100 dark:bg-indigo-900 rounded-full flex items-center justify-center">
                <svg className="w-4 h-4 sm:w-5 sm:h-5 text-indigo-600 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <p className="text-xs text-indigo-900 dark:text-indigo-100">
                <span className="font-bold">Pro Tip:</span> Download <code className="px-1 py-0.5 rounded bg-indigo-100 dark:bg-indigo-900 font-mono font-bold text-xs">nomic-embed-text</code> to enable <strong>RAG</strong> capabilities and chat with your documents.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Sticky Button Footer */}
      <div className="sticky bottom-0 bg-background/95 backdrop-blur-sm border-t border-border p-3 sm:p-4 lg:p-6">
        <div className="max-w-7xl mx-auto flex justify-between gap-3">
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
    </div>
  );
}
