# Types of Config

## Pending To think
 - If the deployment of the app is to be used only by a single team / family
   then settings like OPENAI_API_KEY are common to the entire team / family

 - If the deployment of the app is to be used on a public cloud / entire company
    then these values are also namespaced in a tenant / company

  - So we need to think which are really global values and which are namespaces (k,v)

  - I think the fundamental issue is, is this app multi tenanted or not
    If this is multi tenanted then settings/configurations like OPENAI_API_KEY etc. have to be namespaced
  
  - Global / Tenant / Teams - ?

# Global Deploment Level Configuration
DB_

# Tenant Namespaced Config
OPENAI_API_KEY
OPENI_API_URL

# Teams Level Config
?


