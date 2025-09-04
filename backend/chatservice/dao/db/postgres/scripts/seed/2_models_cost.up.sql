BEGIN;

INSERT INTO model_metadata (id, input_token_cost, cached_token_cost, output_token_cost)
VALUES
  ('gpt-4.1',        3.00, 0.75, 12.00),
  ('gpt-4o',         5.00, 1.25, 15.00),
  ('o3-mini',        0.40, 0.10, 1.60),
  ('o4-mini',        4.00, 1.00, 16.00),
  ('o3',             0.40, 0.10, 1.60),
  ('gemini-2.5-flash', 1.00, 0.25, 4.00),
  ('gemini-2.0-flash', 1.50, 0.35, 6.00),
  ('gemini-2.5-pro', 1.00, 0.25, 4.00),
  ('claude-3.5-haiku', 0.80, 0.20, 4.00),
  ('claude-3.7-sonnet', 2.00, 0.50, 8.00),
  ('claude-4-sonnet', 3.00, 0.75, 12.00),
  ('gpt-5',          1.25, 0.125, 10.00),
  ('gpt-5-mini',     0.25, 0.025, 2.00),
  ('gpt-5-nano',     0.05, 0.005, 0.40)
ON CONFLICT (id) DO UPDATE SET
  input_token_cost  = EXCLUDED.input_token_cost,
  cached_token_cost = EXCLUDED.cached_token_cost,
  output_token_cost = EXCLUDED.output_token_cost;
COMMIT;