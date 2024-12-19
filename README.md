## SOTORKUBE
Sotorkube is an alerting system for failing pod in kubernetes cluster. Sotorkube name means-

- Sotorko - Alert in bangla
- Kube - Kubernetes

You can attach your own webhook and configure alerting.

### How to use

Install dependencies:
```
go mod tidy
```

Sotorkube requires the following environment variables:

- **WEBHOOK_BODY**: JSON body template for webhook message (Put `<TITLE>` and `<MESSAGE>` in the body, they will be replaced with actual value)

- **WEBHOOK_TITLE**: Title for error alerts

- **WEBHOOK_URL**: URL of the webhook endpoint

- **INTERVAL**: Time interval between checks in seconds

Run the application:
```
go run main.go
```
