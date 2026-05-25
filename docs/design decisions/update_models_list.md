# Update Models List

Until now, the models list has been coming from the `shared_models_metadata` table. We were seeding the models list into this table.

## Problem

To update the models list, we have to add a new seed migration and run the server again to update the local database.

## Solution

We host `models.json` on GitHub.

``` 
{
  "metadata": {
    "json_schema_version": "v1"
  },
  "models_list": [{}, {}]
}
```

We also have a `models_version.json` file.

```
{
  "json_schema_version": "v1",
  "model_revision_version": "20260522.2"
}
```

The update flow is:

- `remote(model_revision_version) != local(model_revision_version)` and `remote(json_schema_version) == local(json_schema_version)`: fetch `models.json` and update the local database.
- `remote(model_revision_version) == local(model_revision_version)` and `remote(json_schema_version) == local(json_schema_version)`: do not fetch, because the local models list is already up to date.
- `remote(model_revision_version) != local(model_revision_version)` and `remote(json_schema_version) != local(json_schema_version)`: update SortedChat, because of json_schema_version mismatch

### Desktop App

On service layer initialization, we perform the steps above. So whenever a user starts the desktop app, if there is a version mismatch, the models list is automatically updated.

### Web-App

For users who have deployed SortedChat:

- We added a new RPC to update the local models list.
- They can navigate to the Settings page and update the local models list from there.
