# Collected Auth

## Installation

### 1. Create a **.envrc** file:

```
export PROJECT=YOUR_GOOGLE_CLOUD_PROJECT_ID
```

### 2. Create a **.env** file:

```
GITHUB_CLIENT_ID = …
GITHUB_CLIENT_SECRET = …
GITHUB_REDIRECT_URL = "http://localhost:8080/signin/github/callback"
```

## Deploying

### 1. Copy **app.yaml** to a **app.prod.yaml** file, and add:

```yaml
env_variables:
  GITHUB_CLIENT_ID: "…"
  GITHUB_CLIENT_SECRET: "…"
  GITHUB_REDIRECT_URL: "…/signin/github/callback"
```
