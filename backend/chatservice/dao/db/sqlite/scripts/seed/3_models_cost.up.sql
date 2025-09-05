INSERT INTO model_metadata (id,name, url, provider, input_token_cost, cached_token_cost, output_token_cost) VALUES
  ('gpt-4.1', 'GPT-4.1', 'https://api.openai.com/v1/responses', 'openai',  3.00, 0.75, 12.00),
  ('gpt-4o', 'GPT-4o', 'https://api.openai.com/v1/responses', 'openai',  5.00, 1.25, 15.00),
  ('o3-mini', 'o3-mini', 'https://api.openai.com/v1/responses', 'openai', 0.40, 0.10, 1.60),
  ('o4-mini', 'o4-mini', 'https://api.openai.com/v1/responses', 'openai', 4.00, 1.00, 16.00),
  ('o3', 'o3', 'https://api.openai.com/v1/responses', 'openai', 0.40, 0.10, 1.60),
  ('gemini-2.5-flash', 'gemini-2.5-flash', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent', 'gemini', 1.00, 0.25, 4.00),
  ('gemini-2.0-flash', 'gemini-2.0-flash', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent', 'gemini', 1.50, 0.35, 6.00),
  ('gemini-2.5-pro', 'gemini-2.5-pro', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent',   'gemini', 1.00, 0.25, 4.00),
  ('claude-3.5-haiku', 'claude-3.5-haiku', 'https://api.anthropic.com/v1/messages', 'claude',   0.80, 0.20, 4.00),
  ('claude-3.7-sonnet', 'claude-3.7-sonnet', 'https://api.anthropic.com/v1/messages', 'claude', 2.00, 0.50, 8.00),
  ('claude-4-sonnet', 'claude-4-sonnet', 'https://api.anthropic.com/v1/messages', 'claude', 3.00, 0.75, 12.00),
  ('gpt-5', 'GPT-5', 'https://api.openai.com/v1/responses', 'openai', 1.25, 0.125, 10.00),
  ('gpt-5-mini', 'GPT-5-mini', 'https://api.openai.com/v1/responses', 'openai', 0.25, 0.025, 2.00),
  ('gpt-5-nano', 'GPT-5-nano', 'https://api.openai.com/v1/responses', 'openai', 0.05, 0.005, 0.40)
ON CONFLICT(id) DO UPDATE SET
  input_token_cost=excluded.input_token_cost,
  cached_token_cost=excluded.cached_token_cost,
  output_token_cost=excluded.output_token_cost;
