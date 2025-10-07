import { atom, computed } from 'nanostores';

export interface OnboardingState {
  OPENAI_API_KEY: string;
  OPENAI_API_URL: string;
  OLLAMA_URL: string;
  step: 0 | 1;
  validationErrors: {
    apiKey?: string;
    apiUrl?: string;
    ollamaUrl?: string;
  };
}

const initialState: OnboardingState = {
  OPENAI_API_KEY: '',
  OPENAI_API_URL: '',
  OLLAMA_URL: 'http://localhost:11434',
  step: 0,
  validationErrors: {},
};

export const onboardingStore = atom<OnboardingState>(initialState);

// Computed values
export const currentStep = computed(onboardingStore, (state) => state.step);

// Actions
export const onboardingActions = {
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
    if (state.step < 1) {
      onboardingStore.set({
        ...state,
        step: (state.step + 1) as 0 | 1,
      });
    }
  },

  prevStep: () => {
    const state = onboardingStore.get();
    if (state.step > 0) {
      onboardingStore.set({
        ...state,
        step: (state.step - 1) as 0 | 1,
      });
    }
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
