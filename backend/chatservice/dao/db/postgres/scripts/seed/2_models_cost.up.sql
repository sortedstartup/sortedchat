UPDATE model_metadata SET
   input_token_cost = 3.00,    
   cached_token_cost = 0.75,   
   output_token_cost = 12.00   
   WHERE id = 'gpt-4.1';

UPDATE model_metadata SET
   input_token_cost = 5.00,    
   cached_token_cost = 1.25,   
   output_token_cost = 15.00   
   WHERE id = 'gpt-4o';

UPDATE model_metadata SET
   input_token_cost = 0.40,    
   cached_token_cost = 0.10,   
   output_token_cost = 1.60    
   WHERE id = 'o3-mini';

UPDATE model_metadata SET
   input_token_cost = 4.00,    
   cached_token_cost = 1.00,   
   output_token_cost = 16.00   
   WHERE id = 'o4-mini';

UPDATE model_metadata SET
   input_token_cost = 0.40,    
   cached_token_cost = 0.10,   
   output_token_cost = 1.60    
   WHERE id = 'o3';

UPDATE model_metadata SET
   input_token_cost = 1.00,    
   cached_token_cost = 0.25,   
   output_token_cost = 4.00    
   WHERE id = 'gemini-2.5-flash';

UPDATE model_metadata SET
   input_token_cost = 1.50,    
   cached_token_cost = 0.35,   
   output_token_cost = 6.00    
   WHERE id = 'gemini-2.0-flash';

UPDATE model_metadata SET
   input_token_cost = 1.00,    
   cached_token_cost = 0.25,   
   output_token_cost = 4.00    
   WHERE id = 'gemini-2.5-pro';

UPDATE model_metadata SET
   input_token_cost = 0.80,    
   cached_token_cost = 0.20,   
   output_token_cost = 4.00    
   WHERE id = 'claude-3.5-haiku';

UPDATE model_metadata SET
   input_token_cost = 2.00,    
   cached_token_cost = 0.50,   
   output_token_cost = 8.00    
   WHERE id = 'claude-3.7-sonnet';

UPDATE model_metadata SET
   input_token_cost = 3.00,    
   cached_token_cost = 0.75,   
   output_token_cost = 12.00   
   WHERE id = 'claude-4-sonnet';

UPDATE model_metadata SET
   input_token_cost = 1.25,    
   cached_token_cost = 0.125,
   output_token_cost = 10.00   
   WHERE id = 'gpt-5';

UPDATE model_metadata SET
   input_token_cost = 0.25,    
   cached_token_cost = 0.025,  
   output_token_cost = 2.00   
   WHERE id = 'gpt-5-mini';

UPDATE model_metadata SET
   input_token_cost = 0.05,    
   cached_token_cost = 0.005,   
   output_token_cost = 0.40   
   WHERE id = 'gpt-5-nano';