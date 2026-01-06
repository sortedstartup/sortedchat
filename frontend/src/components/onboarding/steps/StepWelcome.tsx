import { Button } from '@/components/ui/button';
import { onboardingActions } from '@/store/setting';

export function StepWelcome() {
  const handleNext = () => {
    onboardingActions.nextStep();
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 flex items-center justify-center p-8">
      <div className="w-full max-w-7xl">
        {/* Title Section */}
        <div className="text-center mb-12">
          <h1 className="text-5xl font-bold text-white mb-4">Welcome to SortedChat</h1>
          <p className="text-xl text-gray-300">Your privacy-first AI chat platform</p>
        </div>

        {/* Bento Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-12">
          {/* The Model Hub - Large Card */}
          <div 
            className="md:row-span-2 rounded-2xl p-8 bg-cover bg-center relative overflow-hidden group hover:scale-[1.02] transition-transform duration-300"
            style={{
              backgroundImage: "linear-gradient(rgba(0,0,0,0.7), rgba(0,0,0,0.7)), url('data:image/svg+xml,%3Csvg width=\"100\" height=\"100\" xmlns=\"http://www.w3.org/2000/svg\"%3E%3Cpath d=\"M0 0h100v100H0z\" fill=\"%23374151\"/%3E%3C/svg%3E')"
            }}
          >
            <div className="relative z-10">
              <div className="inline-block px-3 py-1 bg-purple-500/20 text-purple-300 text-sm font-semibold rounded-full mb-4">
                MULTI-MODEL
              </div>
              <h2 className="text-3xl font-bold text-white mb-4">The Model Hub</h2>
              <p className="text-gray-300 text-lg">
                Switch between OpenAI, Gemini, Anthropic, or your own local models with a single click.
              </p>
            </div>
          </div>

          {/* Local First */}
          <div className="rounded-2xl p-8 bg-gradient-to-br from-emerald-600 to-teal-700 hover:scale-[1.02] transition-transform duration-300">
            <div className="flex items-start justify-between mb-4">
              <div>
                <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                </div>
                <h3 className="text-2xl font-bold text-white mb-2">Local First</h3>
                <p className="text-emerald-100">
                  Download models from the library to ensure your data never leaves your machine.
                </p>
              </div>
            </div>
          </div>

          {/* Document RAG */}
          <div className="rounded-2xl p-8 bg-gradient-to-br from-orange-500 to-red-600 hover:scale-[1.02] transition-transform duration-300">
            <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4">
              <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <h3 className="text-2xl font-bold text-white mb-2">Document RAG</h3>
            <p className="text-orange-100">
              Create projects, upload docs, and get cited answers instantly.
            </p>
          </div>
        </div>

        {/* Get Started Button */}
        <div className="text-center">
          <Button
            onClick={handleNext}
            size="lg"
            className="px-12 py-6 text-lg bg-purple-600 hover:bg-purple-700 text-white rounded-xl shadow-lg hover:shadow-xl transition-all"
          >
            Get Started
          </Button>
        </div>
      </div>
    </div>
  );
}

