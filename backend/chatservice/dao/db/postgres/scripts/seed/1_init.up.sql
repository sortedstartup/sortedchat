INSERT INTO shared_models_metadata (id, name, url, provider, input_token_cost, output_token_cost,progress,is_downloaded,is_downloadable,status,filestore_id,is_embedding_model,capabilities,cached_token_cost,is_enabled)
   VALUES
  ('gpt-4.1', 'GPT-4.1', 'https://api.openai.com/v1/responses', 'openai', 2.00, 8.00, '',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.50,TRUE),
  ('gpt-4o', 'GPT-4o', 'https://api.openai.com/v1/responses', 'openai', 2.50, 10.00, '',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}',1.25,TRUE),
  ('o3-mini', 'o3-mini', 'https://api.openai.com/v1/responses', 'openai', 1.10, 4.40, '',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":false,"output":false},"realtime":false}',0.55,TRUE),
  ('o4-mini', 'o4-mini', 'https://api.openai.com/v1/responses', 'openai', 1.10, 4.40, '',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.275,TRUE),
  ('o3', 'o3', 'https://api.openai.com/v1/responses', 'openai', 2.00, 8.00, '',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.50,TRUE),
  ('gemini-2.5-flash', 'gemini-2.5-flash', 'https://generativelanguage.googleapis.com/v1beta/chat/completions', 'gemini', 0.30, 2.50,'',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.075,TRUE),
  ('gemini-3-flash-preview', 'gemini-3-flash-preview', 'https://generativelanguage.googleapis.com/v1beta/chat/completions', 'gemini', 0.50, 5.00,'',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":true,"output":false},"video":{"input":true,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.125,TRUE),
  ('gemini-2.0-flash', 'gemini-2.0-flash', 'https://generativelanguage.googleapis.com/v1beta/chat/completions', 'gemini', 0.10, 0.40,'',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":true,"output":false},"video":{"input":true,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.025,TRUE),
  ('gemini-2.5-pro', 'gemini-2.5-pro', 'https://generativelanguage.googleapis.com/v1beta/chat/completions', 'gemini', 1.25, 10.00,'',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":true,"output":false},"video":{"input":true,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.31,TRUE),
  ('gemini-3-pro-preview', 'gemini-3-pro-preview', 'https://generativelanguage.googleapis.com/v1beta/chat/completions', 'gemini', 1.50, 10.00,'',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":true,"output":false},"video":{"input":true,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.375,TRUE),
  ('claude-3.5-haiku','claude-3.5-haiku','https://api.anthropic.com/v1/messages', 'claude', 0.80, 4.00,'',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.08,TRUE),
  ('claude-3-7-sonnet','claude-3.7-sonnet','https://api.anthropic.com/v1/messages', 'claude', 3.00, 15.00,'',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.30,TRUE),
  ('claude-4-sonnet', 'claude-4-sonnet','https://api.anthropic.com/v1/messages', 'claude', 3.00, 15.00,'',FALSE,FALSE,0,NULL,FALSE,'{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}',0.30,TRUE),
  ('gpt-5', 'GPT-5', 'https://api.openai.com/v1/responses', 'openai', 1.25, 10.00, '', FALSE, FALSE, 0, NULL, FALSE, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}', 0.125, TRUE),
  ('gpt-5-mini', 'GPT-5-mini', 'https://api.openai.com/v1/responses', 'openai', 0.25, 2.00, '', FALSE, FALSE, 0, NULL, FALSE, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}', 0.025, TRUE),
  ('gpt-5-nano', 'GPT-5-nano', 'https://api.openai.com/v1/responses', 'openai', 0.05, 0.40, '', FALSE, FALSE, 0, NULL, FALSE, '{"text":{"input":true,"output":true},"audio":{"input":false,"output":false},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":false}', 0.005, TRUE),
  ('gpt-realtime', 'GPT-realtime', 'https://api.openai.com/v1/realtime', 'openai', 32.00, 64.00, '', FALSE, FALSE, 0, NULL, FALSE, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":true},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":true}', 0.40, TRUE),
  ('gpt-realtime-mini', 'GPT-realtime-mini', 'https://api.openai.com/v1/realtime', 'openai', 10.00, 20.00, '', FALSE, FALSE, 0, NULL, FALSE, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":true},"video":{"input":false,"output":false},"image":{"input":true,"output":false},"realtime":true}', 0.30, TRUE),
  ('gpt-4o-realtime', 'GPT-4o-realtime', 'https://api.openai.com/v1/realtime', 'openai', 40.00, 80.00, '', FALSE, FALSE, 0, NULL, FALSE, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":true},"video":{"input":false,"output":false},"image":{"input":false,"output":false},"realtime":true}', 2.50, TRUE),
  ('gpt-4o-mini-realtime', 'GPT-4o-mini-realtime', 'https://api.openai.com/v1/realtime', 'openai', 10.00, 20.00, '', FALSE, FALSE, 0, NULL, FALSE, '{"text":{"input":true,"output":true},"audio":{"input":true,"output":true},"video":{"input":false,"output":false},"image":{"input":false,"output":false},"realtime":true}', 0.30, TRUE),
  ('gemma-3-270m-it-GGUF', 'gemma-3-270m-it-GGUF','https://huggingface.co/unsloth/gemma-3-270m-it-GGUF/resolve/main/gemma-3-270m-it-Q8_0.gguf?download=true', 'local', 0.00, 0.00,'',FALSE,TRUE,0,NULL,FALSE,'',0.00,TRUE),
  ('embeddinggemma-300M-Q8_0', 'embeddinggemma-300M-Q8_0','https://huggingface.co/ggml-org/embeddinggemma-300M-GGUF/resolve/main/embeddinggemma-300M-Q8_0.gguf', 'local', 0.00, 0.00,'',FALSE,TRUE,0,NULL,TRUE,'',0.00,TRUE),
  ('nomic-embed-text-v1.5.Q8_0', 'nomic-embed-text-v1.5.Q8_0','https://huggingface.co/nomic-ai/nomic-embed-text-v1.5-GGUF/resolve/main/nomic-embed-text-v1.5.Q8_0.gguf', 'local', 0.00, 0.00,'',FALSE,TRUE,0,NULL,TRUE,'',0.00,TRUE),
  ('tinyLLama-1.1B-Chat-v1.0-GGUF','tinyLLama-1.1B-Chat-v1.0-GGUF', 'https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q8_0.gguf?download=true', 'local', 0.00, 0.00,'',FALSE,TRUE,0,NULL,FALSE,'',0.00,TRUE),

  ('functiongemma-270m-it-q8_0','functiongemma-270m-it-q8_0', 'https://huggingface.co/ggml-org/functiongemma-270m-it-GGUF/resolve/main/functiongemma-270m-it-q8_0.gguf', 'local', 0, 0,'',FALSE,TRUE,0,NULL,FALSE,'{}',0.00,TRUE),
  ('granite-4.0-h-small-Q4_K_M','granite-4.0-h-small-Q4_K_M', 'https://huggingface.co/ibm-granite/granite-4.0-h-small-GGUF/resolve/main/granite-4.0-h-small-Q4_K_M.gguf', 'local', 0, 0,'',FALSE,TRUE,0,NULL,FALSE,'{}',0.00,TRUE),
  ('granite-4.0-h-small-Q8_0','granite-4.0-h-small-Q8_0', 'https://huggingface.co/ibm-granite/granite-4.0-h-small-GGUF/resolve/main/granite-4.0-h-small-Q8_0.gguf', 'local', 0, 0,'',FALSE,TRUE,0,NULL,FALSE,'{}',0.00,TRUE),
  ('granite-4.0-1b-Q8_0','granite-4.0-1b-Q8_0', 'https://huggingface.co/ibm-granite/granite-4.0-1b-GGUF/resolve/main/granite-4.0-1b-Q8_0.gguf', 'local', 0, 0,'',FALSE,TRUE,0,NULL,FALSE,'{}',0,TRUE),
  ('granite-3.3-8b-instruct-Q8_0', 'granite-3.3-8b-instruct-Q8_0', 'https://huggingface.co/ibm-granite/granite-3.3-8b-instruct-GGUF/resolve/main/granite-3.3-8b-instruct-Q8_0.gguf', 'local', 0.00, 0.00, '', FALSE, TRUE, 0, NULL, FALSE, '{}', 0.00, TRUE),
  ('qwen2.5-coder-7b-instruct-q8_0', 'qwen2.5-coder-7b-instruct-q8_0', 'https://huggingface.co/Qwen/Qwen2.5-Coder-7B-Instruct-GGUF/resolve/main/qwen2.5-coder-7b-instruct-q8_0.gguf', 'local', 0.00, 0.00, '', FALSE, TRUE, 0, NULL, FALSE, '{}', 0.00, TRUE),
  ('Devstral-Small-2-24B-Instruct-2512-Q4_0', 'Devstral-Small-2-24B-Instruct-2512-Q4_0', 'https://huggingface.co/unsloth/Devstral-Small-2-24B-Instruct-2512-GGUF/resolve/main/Devstral-Small-2-24B-Instruct-2512-Q4_0.gguf', 'local', 0.00, 0.00, '', FALSE, TRUE, 0, NULL, FALSE, '{}', 0.00, TRUE),
  ('Devstral-Small-2-24B-Instruct-2512-Q8_0', 'Devstral-Small-2-24B-Instruct-2512-Q8_0', 'https://huggingface.co/unsloth/Devstral-Small-2-24B-Instruct-2512-GGUF/resolve/main/Devstral-Small-2-24B-Instruct-2512-Q8_0.gguf', 'local', 0.00, 0.00, '', FALSE, TRUE, 0, NULL, FALSE, '{}', 0.00, TRUE),
  ('granite-4.0-h-tiny-Q8_0', 'granite-4.0-h-tiny-Q8_0', 'https://huggingface.co/unsloth/granite-4.0-h-tiny-GGUF/resolve/main/granite-4.0-h-tiny-Q8_0.gguf', 'local', 0.00, 0.00, '', FALSE, TRUE, 0, NULL, FALSE, '{}', 0.00, TRUE)

ON CONFLICT(id) DO UPDATE SET
  name=excluded.name,
  url=excluded.url,
  provider=excluded.provider,
  input_token_cost=excluded.input_token_cost,
  output_token_cost=excluded.output_token_cost,
  progress=excluded.progress,
  is_downloaded=excluded.is_downloaded,
  is_downloadable=excluded.is_downloadable,
  status=excluded.status,
  filestore_id=excluded.filestore_id,
  is_embedding_model=excluded.is_embedding_model,
  capabilities=excluded.capabilities,
  cached_token_cost=excluded.cached_token_cost,
  is_enabled=excluded.is_enabled;

INSERT INTO agents (id, name, description, system_prompt, provider, model, local_tools)
VALUES ('ui-widget-designer', 'UI Widget Designer', 'Expert in designing beautiful and functional UI widgets using Tailwind CSS', 'You are a expert UI Widget Generator Agent.

Your main job is to take UI widget requirement from the user as plain text
and create 5 beautiful UX design variations of the widget.

To achieve this follow these steps, start with the template below

0. This will be a standalone html file so always CDN version of libraries and fonts
1. Create $projectname-$timestamp.html to create 5 UX design variations of the widget.
2. Make sure you have all the variants in the same file.
3. For each UI widget explain the ux thinking behind that variant.
6. You can add charts using charts.js
7. Add interactivity using js + jquery(already in index.html) to give it a more real feel
8. You can Use svg.js to draw vector graphics
9. You can use motion.dev for animations
10. Depending on UX requirements decide to use charts, animations
11. Add simple business logic using jquery
11. DO NOT Add any new library on your own, ask the user before adding any new CDN library

if you need logos/ icon use this online service from google in a image tag
<img src="https://www.google.com/s2/favicons?domain=github.com&sz=64">

for general images use this service: https://picsum.photos/400/300, where 400 and 300 is width and height

------
html starter template
------
<!doctype html>
<html>
    <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
        <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js"></script>
        <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.7.1/jquery.min.js"></script>

        <script src="https://cdnjs.cloudflare.com/ajax/libs/svg.js/3.2.5/svg.min.js" crossorigin="anonymous" referrerpolicy="no-referrer" ></script>
        <script src="https://cdn.jsdelivr.net/npm/motion@latest/dist/motion.js"></script>
        <script>
          const { animate, scroll } = Motion
        </script>
    </head>

    <body>
        <h1 class="text-3xl font-bold underline">Hello world!</h1>
    </body>
</html>
', 'openai', 'gemini-3-flash-preview', '[]')
ON CONFLICT(id) DO UPDATE SET
  name=excluded.name,
  description=excluded.description,
  system_prompt=excluded.system_prompt,
  provider=excluded.provider,
  model=excluded.model,
  local_tools=excluded.local_tools;
