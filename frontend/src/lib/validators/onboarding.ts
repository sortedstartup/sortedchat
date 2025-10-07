import { z } from 'zod';

// URL validation schema
const urlSchema = z.string().url('Please enter a valid URL');

// OpenAI API Key validation
export const validateOpenAIKey = (key: string): string | undefined => {
  if (!key.trim()) {
    return 'OpenAI API key is required';
  }
  
  if (!key.startsWith('sk-')) {
    return 'OpenAI API key should start with "sk-"';
  }
  
  if (key.length < 20) {
    return 'OpenAI API key appears to be too short';
  }
  
  return undefined;
};

// API URL validation
export const validateApiUrl = (url: string): string | undefined => {
  if (!url.trim()) {
    return 'API URL is required';
  }
  
  try {
    urlSchema.parse(url);
  } catch (error) {
    return 'Please enter a valid URL (e.g., https://api.openai.com/v1)';
  }
  
  if (!url.startsWith('http://') && !url.startsWith('https://')) {
    return 'URL must start with http:// or https://';
  }
  
  return undefined;
};

// Ollama URL validation
export const validateOllamaUrl = (url: string): string | undefined => {
  if (!url.trim()) {
    return 'Ollama URL is required';
  }
  
  try {
    urlSchema.parse(url);
  } catch (error) {
    return 'Please enter a valid URL (e.g., http://localhost:11434)';
  }
  
  if (!url.startsWith('http://') && !url.startsWith('https://')) {
    return 'URL must start with http:// or https://';
  }
  
  return undefined;
};

// Connectivity validation functions
export const validateOpenAIConnection = async (apiKey: string, apiUrl?: string): Promise<string | undefined> => {
  try {
    const baseUrl = apiUrl || 'https://api.openai.com/v1';
    const response = await fetch(`${baseUrl}/models`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json',
      },
    });
    
    if (!response.ok) {
      if (response.status === 401) {
        return 'Invalid API key or unauthorized access';
      } else if (response.status === 403) {
        return 'API key does not have sufficient permissions';
      } else {
        return `Connection failed: ${response.status} ${response.statusText}`;
      }
    }
    
    return undefined;
  } catch (error) {
    return `Connection failed: ${error instanceof Error ? error.message : 'Unknown error'}`;
  }
};

export const validateLiteLLMConnection = async (apiUrl: string): Promise<string | undefined> => {
  try {
    // Try to reach the LiteLLM health endpoint or models endpoint
    const healthUrl = `${apiUrl.replace(/\/$/, '')}/health`;
    const response = await fetch(healthUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    if (!response.ok) {
      // Try models endpoint as fallback
      const modelsUrl = `${apiUrl.replace(/\/$/, '')}/models`;
      const modelsResponse = await fetch(modelsUrl, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      if (!modelsResponse.ok) {
        return `Cannot connect to LiteLLM server at ${apiUrl}`;
      }
    }
    
    return undefined;
  } catch (error) {
    return `Connection failed: ${error instanceof Error ? error.message : 'Unknown error'}`;
  }
};

export const validateOllamaConnection = async (ollamaUrl: string): Promise<string | undefined> => {
  try {
    const tagsUrl = `${ollamaUrl.replace(/\/$/, '')}/api/tags`;
    const response = await fetch(tagsUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    if (!response.ok) {
      return `Cannot connect to Ollama server at ${ollamaUrl}`;
    }
    
    const data = await response.json();
    
    // Check if there are any models available
    if (!data.models || data.models.length === 0) {
      return 'Ollama is running but no models are installed. Please install an embedding model first.';
    }
    
    // Check for embedding models (common ones)
    const embeddingModels = data.models.filter((model: any) => 
      model.name.includes('embed') || 
      model.name.includes('nomic') ||
      model.name.includes('all-minilm')
    );
    
    if (embeddingModels.length === 0) {
      return 'No embedding models found. Please install an embedding model like "nomic-embed-text" or "all-minilm".';
    }
    
    return undefined;
  } catch (error) {
    if (error instanceof TypeError && error.message.includes('fetch')) {
      return `Cannot reach Ollama server at ${ollamaUrl}. Make sure Ollama is running.`;
    }
    return `Connection failed: ${error instanceof Error ? error.message : 'Unknown error'}`;
  }
};

// Combined validation for the current step
export const validateCurrentStep = async (
  step: number,
  provider: 'openai' | 'litellm',
  apiKey: string,
  apiUrl: string,
  ollamaUrl: string
): Promise<{ [key: string]: string }> => {
  const errors: { [key: string]: string } = {};
  
  if (step === 0) {
    // Validate API provider settings
    if (provider === 'openai') {
      const keyError = validateOpenAIKey(apiKey);
      if (keyError) errors.apiKey = keyError;
      
      if (apiUrl) {
        const urlError = validateApiUrl(apiUrl);
        if (urlError) errors.apiUrl = urlError;
      }
      
      // Test connection if basic validation passes
      if (!keyError && !errors.apiUrl) {
        const connectionError = await validateOpenAIConnection(apiKey, apiUrl);
        if (connectionError) errors.apiKey = connectionError;
      }
    } else {
      // LiteLLM
      const urlError = validateApiUrl(apiUrl);
      if (urlError) {
        errors.apiUrl = urlError;
      } else {
        const connectionError = await validateLiteLLMConnection(apiUrl);
        if (connectionError) errors.apiUrl = connectionError;
      }
    }
  } else if (step === 1) {
    // Validate Ollama settings
    const urlError = validateOllamaUrl(ollamaUrl);
    if (urlError) {
      errors.ollamaUrl = urlError;
    } else {
      const connectionError = await validateOllamaConnection(ollamaUrl);
      if (connectionError) errors.ollamaUrl = connectionError;
    }
  }
  
  return errors;
};
