import { Button } from '@/components/ui/button';
import { onboardingActions } from '@/store/setting';

export function StepWelcome() {
  const handleNext = () => {
    onboardingActions.nextStep();
  };

  return (
    <div className="h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 overflow-auto">
      <div className="flex items-center justify-center p-4 sm:p-6 lg:p-8">
        <div className="w-full max-w-7xl py-4 sm:py-6">
          {/* Title Section */}
          <div className="text-center mb-4 sm:mb-6 lg:mb-8">
            <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-white mb-2">Welcome to SortedChat</h1>
            <p className="text-sm sm:text-base lg:text-lg text-gray-300">Your privacy-first AI chat platform</p>
          </div>

          {/* Bento Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-3 sm:gap-4 lg:gap-6 mb-4 sm:mb-6 lg:mb-8">
            {/* Agents Section */}
            <div className="rounded-2xl sm:rounded-3xl p-4 sm:p-5 lg:p-6 bg-gradient-to-br from-indigo-600 to-violet-800 hover:scale-[1.02] transition-transform duration-300 shadow-2xl border border-white/10 flex flex-col justify-between">
              <div>
                <div className="w-8 h-8 sm:w-10 sm:h-10 lg:w-12 lg:h-12 bg-white/20 rounded-xl flex items-center justify-center mb-2 sm:mb-3 lg:mb-4">
                  <svg className="w-5 h-5 sm:w-6 sm:h-6 lg:w-7 lg:h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
                  </svg>
                </div>
                <h2 className="text-xl sm:text-2xl lg:text-3xl font-black text-white mb-1 tracking-tight">Agents</h2>
                <h3 className="text-sm sm:text-base lg:text-lg font-bold text-indigo-200 mb-2 sm:mb-3 lg:mb-4">Local / Remote Models</h3>
                <p className="text-indigo-50 text-xs sm:text-sm lg:text-base leading-relaxed mb-3 sm:mb-4 lg:mb-6">
                  Build autonomous agents that execute tasks using local privacy-first models or powerful cloud APIs.
                </p>
              </div>
              <div className="bg-white/10 backdrop-blur-sm rounded-lg sm:rounded-xl p-2 sm:p-3 border border-white/5">
                <p className="text-indigo-100 text-xs font-semibold flex items-center">
                  <svg className="w-3 h-3 sm:w-4 sm:h-4 mr-2 text-green-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                  </svg>
                  Supreme Privacy: Keep your agent data local.
                </p>
              </div>
            </div>

            {/* Chat Section */}
            <div className="rounded-2xl sm:rounded-3xl p-4 sm:p-5 lg:p-6 bg-gradient-to-br from-blue-600 to-cyan-800 hover:scale-[1.02] transition-transform duration-300 shadow-2xl border border-white/10 flex flex-col justify-between">
              <div>
                <div className="w-8 h-8 sm:w-10 sm:h-10 lg:w-12 lg:h-12 bg-white/20 rounded-xl flex items-center justify-center mb-2 sm:mb-3 lg:mb-4">
                  <svg className="w-5 h-5 sm:w-6 sm:h-6 lg:w-7 lg:h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                  </svg>
                </div>
                <h2 className="text-xl sm:text-2xl lg:text-3xl font-black text-white mb-1 tracking-tight">Chat</h2>
                <h3 className="text-sm sm:text-base lg:text-lg font-bold text-cyan-200 mb-2 sm:mb-3 lg:mb-4">Local / Remote Models</h3>
                <p className="text-cyan-50 text-xs sm:text-sm lg:text-base leading-relaxed mb-3 sm:mb-4 lg:mb-6">
                  Instant access to LLMs. Chat with local Llama, Mistral, or connect OpenAI, Gemini, and Claude.
                </p>
              </div>
              <div className="bg-white/10 backdrop-blur-sm rounded-lg sm:rounded-xl p-2 sm:p-3 border border-white/5">
                <p className="text-cyan-100 text-xs font-semibold flex items-center">
                  <svg className="w-3 h-3 sm:w-4 sm:h-4 mr-2 text-cyan-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  Blazing Fast: Optimized for local inference.
                </p>
              </div>
            </div>

            {/* RAG Section */}
            <div className="rounded-2xl sm:rounded-3xl p-4 sm:p-5 lg:p-6 bg-gradient-to-br from-emerald-600 to-teal-800 hover:scale-[1.02] transition-transform duration-300 shadow-2xl border border-white/10 flex flex-col justify-between">
              <div>
                <div className="w-8 h-8 sm:w-10 sm:h-10 lg:w-12 lg:h-12 bg-white/20 rounded-xl flex items-center justify-center mb-2 sm:mb-3 lg:mb-4">
                  <svg className="w-5 h-5 sm:w-6 sm:h-6 lg:w-7 lg:h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <h2 className="text-xl sm:text-2xl lg:text-3xl font-black text-white mb-1 tracking-tight">RAG</h2>
                <h3 className="text-sm sm:text-base lg:text-lg font-bold text-emerald-200 mb-2 sm:mb-3 lg:mb-4">Projects • Local / Remote</h3>
                <p className="text-emerald-50 text-xs sm:text-sm lg:text-base leading-relaxed mb-3 sm:mb-4 lg:mb-6">
                  Generate and store embeddings locally. Search relevant document snippets to provide context to any model.
                </p>
              </div>
              <div className="bg-white/10 backdrop-blur-sm rounded-lg sm:rounded-xl p-2 sm:p-3 border border-white/5">
                <p className="text-emerald-100 text-xs font-semibold flex items-center">
                  <svg className="w-3 h-3 sm:w-4 sm:h-4 mr-2 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                  Secure Storage: All embeddings stay on your machine.
                </p>
              </div>
            </div>
          </div>

          {/* Get Started Button */}
          <div className="text-center">
            <Button
              onClick={handleNext}
              size="lg"
              className="px-6 sm:px-8 lg:px-10 py-3 sm:py-4 lg:py-5 text-sm sm:text-base bg-purple-600 hover:bg-purple-700 text-white rounded-xl shadow-lg hover:shadow-xl transition-all"
            >
              Get Started
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

