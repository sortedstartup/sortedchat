# Update Models List
- Till now, models list was coming from `shared_models_metadata` table
- We were seeding the models list in the table 

## Problem 
- To update the model list, we have to add new seed migration and have to again run the server to update the local db 

## Solution 
- We hosted the models.json on github
```
{
  "metadata" : {
    "json_schema_version": "v1"
  },
  "models_list": [{},{}]
}
```
- we have a models_version.json file 
```
{
  "json_schema_version": "v1",
  "model_revision_version": "20260522.2"
}
```
- remote(model_revision_version) != local(model_revision_version) && remote(json_schema_version) == local(json_schema_version) -> fetch model_json and update local db
- remote(model_revision_version) == local(model_revision_version) && remote(json_schema_version) == local(json_schema_version) -> dont fetch(already up to date)
- remote(model_revision_version) != local(model_revision_version) && remote(json_schema_version) != local(json_schema_version) -> Update Sortedchat (cant update models list)

### Desktop App
- On service init, we do above steps
- so whenever user start the desktop app and if there is version mismatch, models list will be automatically updated

### Web-App
- For users, who have deployed sortedchat 
- Added new RPC to Update local models list
- They can navigate to Settings page, and can update local models list from there