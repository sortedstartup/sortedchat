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
      <div className="mb-8 text-center max-w-3xl mx-auto">
        <h2 className="text-3xl font-extrabold mb-3 bg-clip-text text-transparent bg-gradient-to-r from-blue-600 to-cyan-500">
          Download Local Models
        </h2>
        <p className="text-gray-600 dark:text-gray-400 text-lg leading-relaxed">
          Power your privacy-first AI by downloading models in <span className="font-semibold text-blue-600 dark:text-blue-400">GGUF</span> format. 
          We utilize <code className="px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800 font-mono text-sm">llama-server</code> to enable seamless local interactions.
        </p>
      </div>

      <div className="flex-1 overflow-hidden border rounded-2xl bg-background/50 backdrop-blur-sm shadow-inner mb-6">
        <LocalProvider />
      </div>

      <div className="bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-100 dark:border-indigo-900/50 rounded-xl p-5 mb-6 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex-shrink-0 w-10 h-10 bg-indigo-100 dark:bg-indigo-900 rounded-full flex items-center justify-center">
            <svg className="w-6 h-6 text-indigo-600 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <p className="text-sm text-indigo-900 dark:text-indigo-100">
            <span className="font-bold">Pro Tip:</span> Download <code className="px-1.5 py-0.5 rounded bg-indigo-100 dark:bg-indigo-900 font-mono font-bold">nomic-embed-text</code> to enable <strong>RAG</strong> capabilities and chat with your documents.
          </p>
        </div>
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
