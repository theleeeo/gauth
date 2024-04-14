# Thor
Since there is a trend of naming applications with the names of ancient gods...
Introducing Thor, named after Thor, the god.

Thor is an IAM Server providing JWT tokens for authentication and authorization (yes there is a difference).

## OAuth

### Login
Navigating to the page at /oauth/login/\<provider>/\<name> will initiate the login flow. The provider is the type of the oauth-provider (e.g. google, github, etc.) and the name is the name of the application.

The provider should be configured to redirect to \<base-url>/oauth/callback/\<provider>/\<name> where \<base-url> is the url att which the application is reachable.

### Providers
The following providers are supported:
- Google
- Github
