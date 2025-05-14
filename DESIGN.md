# First milestone: create a chatgpt clone, but allow model switching
## Features
- Allow chat and maintain chat context in a long conversation
- Auto Save chats and be visible on the left sidebar
   - Chats should be saved in a database (sqlite) using dao layer, the dao layer will allow us to switch databases e.g. postgres
- On click on any chat on the side bar show that chat (like chatgpt)

# APIs
  - Chat API - takes in a text request and streaming reponse
     - later worry about pdf, images and audio any video (probably going to come through http api)
  - SaveChat ( develop further )
     - may be chat api should automatically do it
  - GetChatHistory (paginated)
  - GetChat(id)
  
---
  Later
  - Search across chats

# Dao
  - Simple support only sqlite



