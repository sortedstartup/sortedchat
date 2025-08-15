INSERT INTO model_metadata (id, name, url, provider, input_token_cost, output_token_cost)
   VALUES 
   ('gpt-4.1', 'GPT-4.1', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01),
   ('gpt-4o', 'GPT-4o', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01),
   ('o3-mini', 'o3-mini', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01),
   ('o4-mini', 'o4-mini', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01),
   ('o3', 'o3', 'https://api.openai.com/v1/responses', 'openai', 0.01, 0.01),
   ('gemini-2.5-flash', 'gemini-2.5-flash', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent', 'gemini', 0.01, 0.01),
   ('gemini-2.0-flash', 'gemini-2.0-flash', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent', 'gemini', 0.01, 0.01),
   ('gemini-2.5-pro', 'gemini-2.5-pro', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent', 'gemini', 0.01, 0.01),
   ('claude-3.5-haiku', 'claude-3.5-haiku', 'https://api.anthropic.com/v1/messages', 'claude', 0.01, 0.01),
   ('claude-3.7-sonnet', 'claude-3.7-haiku', 'https://api.anthropic.com/v1/messages', 'claude', 0.01, 0.01),
   ('claude-4-sonnet', 'claude-4-sonnet', 'https://api.anthropic.com/v1/messages', 'claude', 0.01, 0.01);


