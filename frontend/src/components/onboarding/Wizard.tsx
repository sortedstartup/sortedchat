import { useStore } from '@nanostores/react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { currentStep } from '@/stores/onboardingStore';
import { StepApi } from './steps/StepApi';
import { StepEmbeddings } from './steps/StepEmbeddings';
import { StepFinish } from './steps/StepFinish';

const steps = [
  {
    title: 'Connect your LLM provider',
    description: 'Use an OpenAI key or a LiteLLM proxy URL.',
  },
  {
    title: 'Enable embeddings (Ollama)',
    description: 'Provide your local Ollama URL to power retrieval and embeddings.',
  },
  {
    title: 'All set!',
    description: 'Your configuration is saved. You can change it anytime in Settings.',
  },
];

export function OnboardingWizard() {
  const step = useStore(currentStep);
  
  const progress = ((step + 1) / steps.length) * 100;
  
  const renderStep = () => {
    switch (step) {
      case 0:
        return <StepApi />;
      case 1:
        return <StepEmbeddings />;
      case 2:
        return <StepFinish />;
      default:
        return <StepApi />;
    }
  };
  
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800 flex items-center justify-center p-4">
      <div className="w-full max-w-2xl">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
            Welcome to SortedChat
          </h1>
          <p className="text-gray-600 dark:text-gray-300">
            Let's get you set up with your AI providers
          </p>
        </div>
        
        <Card className="shadow-xl">
          <CardHeader className="pb-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
                  Step {step + 1} of {steps.length}
                </span>
              </div>
              <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
                {Math.round(progress)}%
              </span>
            </div>
            <Progress value={progress} className="mb-6" />
            
            <CardTitle className="text-xl">
              {steps[step].title}
            </CardTitle>
            <CardDescription className="text-base">
              {steps[step].description}
            </CardDescription>
          </CardHeader>
          
          <CardContent className="pt-0">
            {renderStep()}
          </CardContent>
        </Card>
        
        <div className="text-center mt-6">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Need help? Check our{' '}
            <a href="#" className="text-blue-600 dark:text-blue-400 hover:underline">
              documentation
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}
