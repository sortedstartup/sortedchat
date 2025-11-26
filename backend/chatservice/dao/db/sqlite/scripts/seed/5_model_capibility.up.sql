INSERT INTO model_metadata (id, name, url, provider, input_token_cost, cached_token_cost, output_token_cost, capabilities) VALUES
  ('gpt-4.1', 'GPT-4.1', 'https://api.openai.com/v1/responses', 'openai', 2.00, 0.50, 8.00, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('gpt-4o', 'GPT-4o', 'https://api.openai.com/v1/responses', 'openai', 2.50, 1.25, 10.00, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('o3-mini', 'o3-mini', 'https://api.openai.com/v1/responses', 'openai', 1.10, 0.55, 4.40, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":false,"output":false},"realtime":false}'),
  ('o4-mini', 'o4-mini', 'https://api.openai.com/v1/responses', 'openai', 1.10, 0.275, 4.40, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('o3', 'o3', 'https://api.openai.com/v1/responses', 'openai', 2.00, 0.50, 8.00, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('gemini-2.5-flash', 'gemini-2.5-flash', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent', 'gemini', 0.30, 0.075, 2.50, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":false},"video":{"input":true,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('gemini-2.0-flash', 'gemini-2.0-flash', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent', 'gemini', 0.10, 0.025, 0.40, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":false},"video":{"input":true,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('gemini-2.5-pro', 'gemini-2.5-pro', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent', 'gemini', 1.25, 0.31, 10.00, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":false},"video":{"input":true,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('claude-3.5-haiku', 'claude-3.5-haiku', 'https://api.anthropic.com/v1/messages', 'claude', 0.80, 0.08, 4.00, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('claude-3.7-sonnet', 'claude-3.7-sonnet', 'https://api.anthropic.com/v1/messages', 'claude', 3.00, 0.30, 15.00, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('claude-4-sonnet', 'claude-4-sonnet', 'https://api.anthropic.com/v1/messages', 'claude', 3.00, 0.30, 15.00, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('gpt-5', 'GPT-5', 'https://api.openai.com/v1/responses', 'openai', 1.25, 0.125, 10.00, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('gpt-5-mini', 'GPT-5-mini', 'https://api.openai.com/v1/responses', 'openai', 0.25, 0.025, 2.00, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('gpt-5-nano', 'GPT-5-nano', 'https://api.openai.com/v1/responses', 'openai', 0.05, 0.005, 0.40, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}'),
  ('gpt-realtime', 'GPT-realtime', 'https://api.openai.com/v1/realtime', 'openai', 32.00, 0.50, 64.00, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":true},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":true}'),
  ('gpt-realtime-mini', 'GPT-realtime-mini', 'https://api.openai.com/v1/realtime', 'openai', 10.00, 0.30, 20.00, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":true},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":true}'),
  ('gpt-4o-realtime', 'GPT-4o-realtime', 'https://api.openai.com/v1/realtime', 'openai', 40.00, 2.50, 80.00, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":true},"video":{"input":false,"output":false},"image":{"input":false,"output":false},"realtime":true}'),
  ('gpt-4o-mini-realtime', 'GPT-4o-mini-realtime', 'https://api.openai.com/v1/realtime', 'openai', 10.00, 0.30, 20.00, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":true},"video":{"input":false,"output":false},"image":{"input":false,"output":false},"realtime":true}'),
  ('gemini-live-2.5-flash-preview', 'gemini-live-2.5-flash', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent', 'gemini', 3.00, 0.00, 12.00, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":true},"video":{"input":true,"output":false},"image":{"input":true,"output":false},"realtime":true}')


ON CONFLICT(id) DO UPDATE SET
  name=excluded.name,
  url=excluded.url,
  provider=excluded.provider,
  input_token_cost=excluded.input_token_cost,
  cached_token_cost=excluded.cached_token_cost,
  output_token_cost=excluded.output_token_cost,
  capabilities=excluded.capabilities;
