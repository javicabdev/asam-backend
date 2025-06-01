# ASAM Backend - Deployment Guide

## Prerequisites

- Google Cloud SDK installed and configured
- Docker installed (for local builds)
- Access to the Google Cloud project
- PostgreSQL database already configured (Aiven)

## Environment Variables

The application requires the following environment variables in production:

### Required Variables

- `ENVIRONMENT`: Set to "production"
- `ADMIN_USER`: Administrator username for accessing protected endpoints
- `ADMIN_PASSWORD`: Administrator password
- `JWT_ACCESS_SECRET`: Secret key for JWT access tokens (auto-generated)
- `JWT_REFRESH_SECRET`: Secret key for JWT refresh tokens (auto-generated)

**Important**: Do NOT set the `PORT` variable. Cloud Run automatically assigns and manages this variable.

### Optional Variables

- `SMTP_USER`: SMTP username for email notifications (optional)
- `SMTP_PASSWORD`: SMTP password for email notifications (optional)

If SMTP credentials are not provided, the notification service will be disabled but the application will continue to function.

## Deployment Steps

### 1. Fix Lint Issues (if any)

The project uses golangci-lint v2.1.6. If you encounter lint issues, they are configured in `.golangci.yml`.

### 2. Deploy to Cloud Run

The deployment is automated through GitHub Actions, but you can also deploy manually:

```bash
# Build and push to Google Container Registry
gcloud builds submit --tag gcr.io/YOUR-PROJECT-ID/asam-backend

# Deploy to Cloud Run
gcloud run deploy asam-backend \
  --image gcr.io/YOUR-PROJECT-ID/asam-backend \
  --region europe-west1 \
  --platform managed \
  --allow-unauthenticated
```

### 3. Configure Environment Variables

After deployment, you need to set the required environment variables.

#### On Windows (PowerShell):

```powershell
# Run the PowerShell script
.\scripts\Set-CloudRunEnv.ps1
```

#### On Linux/Mac:

```bash
# Run the bash script
./scripts/set-cloudrun-env.sh
```

The script will:
- Generate secure random JWT secrets
- Prompt you for admin credentials
- Optionally configure SMTP settings
- Update the Cloud Run service with these variables

### 4. Verify Deployment

Check the service health:

```bash
# Get the service URL
gcloud run services describe asam-backend --region europe-west1 --format 'value(status.url)'

# Test the health endpoint
curl https://YOUR-SERVICE-URL/health
```

## Troubleshooting

### Database Connection Issues

If you see "admin credentials must be set for production" error:
- Ensure you've run the environment configuration script
- Verify the ADMIN_USER and ADMIN_PASSWORD are set

### SMTP Configuration Issues

If you see "SMTP credentials not configured" warning:
- This is not an error - the app will run without email notifications
- To enable emails, provide SMTP_USER and SMTP_PASSWORD when running the configuration script

### Build Failures

If the Docker build fails at the GraphQL generation step:
- The Dockerfile now uses `gqlgen` directly instead of the custom generator
- Ensure all GraphQL schema files are present in `internal/adapters/gql/schema/`

## Security Notes

1. **JWT Secrets**: The configuration scripts generate cryptographically secure random secrets. Store these safely.

2. **Admin Credentials**: Choose strong credentials for the admin user. These are used to access:
   - `/debug/*` endpoints in production
   - Other protected administrative endpoints

3. **Database**: The database password is set through Cloud Run build triggers and should not be committed to the repository.

## Monitoring

Once deployed, you can monitor the application through:

- **Logs**: `gcloud run services logs read asam-backend --region europe-west1`
- **Metrics**: Available at `/metrics` endpoint (Prometheus format)
- **Health**: `/health` endpoint provides detailed health status
- **Cloud Run Console**: View metrics, logs, and revisions in the GCP Console

## Rolling Back

If you need to rollback to a previous version:

```bash
# List revisions
gcloud run revisions list --service asam-backend --region europe-west1

# Route traffic to a previous revision
gcloud run services update-traffic asam-backend \
  --region europe-west1 \
  --to-revisions REVISION-NAME=100
```
