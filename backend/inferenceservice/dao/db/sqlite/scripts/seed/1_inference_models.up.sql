INSERT INTO inference_model_metadata (id, name, url, provider, input_token_cost, output_token_cost,progress,is_downloaded,is_downloadable,status)
   VALUES 
   ('gpt-4.1', 'GPT-4.1', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01,'',FALSE,FALSE,0),
   ('gpt-4o', 'GPT-4o', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01,'',FALSE,FALSE,0),
   ('o3-mini', 'o3-mini', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01,'',FALSE,FALSE,0),
   ('o4-mini', 'o4-mini', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01,'',FALSE,FALSE,0),
   ('o3', 'o3', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01,'',FALSE,FALSE,0),
   ('gemini-2.5-flash', 'gemini-2.5-flash', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent', 'gemini', 0.01, 0.01,'',FALSE,FALSE,0),
   ('gemini-2.0-flash', 'gemini-2.0-flash', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent', 'gemini', 0.01, 0.01,'',FALSE,FALSE,0),
   ('gemini-2.5-pro', 'gemini-2.5-pro', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent', 'gemini', 0.01, 0.01,'',FALSE,FALSE,0),
   ('claude-3.5-haiku', 'claude-3.5-haiku', 'https://api.anthropic.com/v1/messages', 'claude', 0.01, 0.01,'',FALSE,FALSE,0),
   ('claude-3.7-sonnet', 'claude-3.7-haiku', 'https://api.anthropic.com/v1/messages', 'claude', 0.01, 0.01,'',FALSE,FALSE,0),
   ('claude-4-sonnet', 'claude-4-sonnet', 'https://api.anthropic.com/v1/messages', 'claude', 0.01, 0.01,'',FALSE,FALSE,0),
   ('tinyLLama-1.1B-Chat-v1.0-GGUF', 'tinyLLama-1.1B-Chat-v1.0-GGUF','https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q8_0.gguf?download=true', 'TheBloke', 0.01, 0.01,'',FALSE,TRUE,0);