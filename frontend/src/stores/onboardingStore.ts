import { atom, computed } from 'nanostores';

export type OnboardingProvider = 'openai' | 'litellm';

export interface OnboardingState {
  provider: OnboardingProvider;
  OPENAI_API_KEY: string;
  OPENAI_API_URL: string;
  OLLAMA_URL: string;
  step: 0 | 1 | 2;
  isValidating: boolean;
  validationErrors: {
    apiKey?: string;
    apiUrl?: string;
    ollamaUrl?: string;
  };
}

const initialState: OnboardingState = {
  provider: 'openai',
  OPENAI_API_KEY: '',
  OPENAI_API_URL: '',
  OLLAMA_URL: 'http://localhost:11434',
  step: 0,
  isValidating: false,
  validationErrors: {},
};

export const onboardingStore = atom<OnboardingState>(initialState);

// Computed values
export const currentStep = computed(onboardingStore, (state) => state.step);
export const isFirstStep = computed(onboardingStore, (state) => state.step === 0);
export const isLastStep = computed(onboardingStore, (state) => state.step === 2);
export const canProceed = computed(onboardingStore, (state) => {
  const { provider, OPENAI_API_KEY, OPENAI_API_URL, OLLAMA_URL, step, validationErrors } = state;
  
  if (Object.keys(validationErrors).length > 0) return false;
  
  switch (step) {
    case 0: // API provider step
      if (provider === 'openai') {
        return OPENAI_API_KEY.trim() !== '';
      } else {
        return OPENAI_API_URL.trim() !== '';
      }
    case 1: // Ollama step
      return OLLAMA_URL.trim() !== '';
    case 2: // Finish step
      return true;
    default:
      return false;
  }
});

// Actions
export const onboardingActions = {
  setProvider: (provider: OnboardingProvider) => {
    onboardingStore.set({
      ...onboardingStore.get(),
      provider,
      validationErrors: {},
    });
  },

  setApiKey: (key: string) => {
    const state = onboardingStore.get();
    onboardingStore.set({
      ...state,
      OPENAI_API_KEY: key,
      validationErrors: { ...state.validationErrors, apiKey: undefined },
    });
  },

  setApiUrl: (url: string) => {
    const state = onboardingStore.get();
    onboardingStore.set({
      ...state,
      OPENAI_API_URL: url,
      validationErrors: { ...state.validationErrors, apiUrl: undefined },
    });
  },

  setOllamaUrl: (url: string) => {
    const state = onboardingStore.get();
    onboardingStore.set({
      ...state,
      OLLAMA_URL: url,
      validationErrors: { ...state.validationErrors, ollamaUrl: undefined },
    });
  },

  nextStep: () => {
    const state = onboardingStore.get();
    if (state.step < 2) {
      onboardingStore.set({
        ...state,
        step: (state.step + 1) as 0 | 1 | 2,
      });
    }
  },

  prevStep: () => {
    const state = onboardingStore.get();
    if (state.step > 0) {
      onboardingStore.set({
        ...state,
        step: (state.step - 1) as 0 | 1 | 2,
      });
    }
  },

  setValidating: (isValidating: boolean) => {
    onboardingStore.set({
      ...onboardingStore.get(),
      isValidating,
    });
  },

  setValidationError: (field: keyof OnboardingState['validationErrors'], error?: string) => {
    const state = onboardingStore.get();
    onboardingStore.set({
      ...state,
      validationErrors: {
        ...state.validationErrors,
        [field]: error,
      },
    });
  },

  clearValidationErrors: () => {
    onboardingStore.set({
      ...onboardingStore.get(),
      validationErrors: {},
    });
  },

  reset: () => {
    onboardingStore.set(initialState);
  },
};
