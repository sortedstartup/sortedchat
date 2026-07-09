## Tailscale Serve
Tailscale Serve allows you to share a service within your tailnet, and Tailscale will set the header Tailscale-User-Login with the email address of the requester.

`tailscale serve` lets you share a local service securely within your tailnet.
 [https://tailscale.com/kb/1312/serve](https://tailscale.com/kb/1312/serve)

It sets these headers which can be used for auth - 
|Header|Purpose|
|-------|-------|
|Tailscale-User-Login|Requester's login name (for example, alice@example.com) |
|Tailscale-User-Name| Filled with the requester's display name (for example, Alice Architect) |
|Tailscale-User-Profile-Pic| If their identity provider provides one (for example, https://example.com/photo.jpg)|

## Cloudflare Tunnel
Cloudflare Tunnel can be used with Cloudflare Access to protect Open WebUI with SSO. This is barely documented by Cloudflare, but Cf-Access-Authenticated-User-Email is set with the email address of the authenticated user.


## OAuth Proxy
https://oauth2-proxy.github.io/oauth2-proxy/


https://goauthentik.io/

https://www.authelia.com/