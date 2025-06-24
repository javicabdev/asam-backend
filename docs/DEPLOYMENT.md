# ASAM Backend - Deployment Guide

## Prerequisites

- Google Cloud SDK installed and configured
- Docker installed (for local builds)
- Access to the Google Cloud project
- PostgreSQL database already configured (Aiven)

## Quick Start - Pre-deployment Check

Before deploying, run the pre-deployment check script to ensure everything is configured correctly:

```powershell
# Windows
.\scripts\gcp\pre-deploy-check.ps1 -ProjectId YOUR-PROJECT-ID

# Linux/Mac
./scripts/gcp/pre-deploy-check.sh YOUR-PROJECT-ID
```

This script will verify:
- Google Cloud SDK installation and authentication
- Required APIs are enabled
- Service account exists with correct permissions
- All required secrets are configured
- Docker images availability

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

## Database Secrets Configuration

Before deploying, ensure all database secrets are configured in Google Secret Manager:

```powershell
# Windows - Verify and create secrets
.\scripts\gcp\verify-db-secrets.ps1 -ProjectId YOUR-PROJECT-ID -CreateSecrets

# Linux/Mac - Verify and create secrets
./scripts/gcp/verify-db-secrets.sh YOUR-PROJECT-ID --create-secrets
```

Required secrets:
- `db-host`: PostgreSQL host (e.g., pg-xxx.aivencloud.com)
- `db-port`: PostgreSQL port (e.g., 14276)
- `db-user`: Database user (e.g., avnadmin)
- `db-password`: Database password
- `db-name`: Database name (e.g., defaultdb)

## Deployment Steps

### 1. Fix Lint Issues (if any)

The project uses golangci-lint v2.1.6. If you encounter lint issues, they are configured in `.golangci.yml`.

### 2. Deploy to Cloud Run

The deployment is automated through GitHub Actions. To deploy:

1. Go to your repository on GitHub
2. Navigate to Actions tab
3. Select "Deploy to Google Cloud Run" workflow
4. Click "Run workflow"
5. Select options:
   - Environment: production or staging
   - Image tag: latest or specific version
   - Run migrations: check if you need to run database migrations

You can also deploy manually:

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

For detailed troubleshooting guide, see [TROUBLESHOOTING.md](./TROUBLESHOOTING.md).

### Common Issues

#### Database Migration Errors

If migrations fail with "connection parameters not found":
1. Verify secrets are configured: `.\scripts\gcp\verify-db-secrets.ps1 -ProjectId YOUR-PROJECT-ID`
2. The workflow exports variables with both `DB_` and `POSTGRES_` prefixes for compatibility
3. Check the migration logs in GitHub Actions for specific errors

#### Service Account Permission Issues

If you see permission errors:
```bash
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member='serviceAccount:github-actions-deploy@YOUR-PROJECT-ID.iam.gserviceaccount.com' \
  --role='roles/secretmanager.secretAccessor'
```

#### Build Failures

If the Docker build fails:
- Ensure all GraphQL schema files are present
- Check that the image tag format is correct
- Verify Container Registry API is enabled

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

### Useful Commands

```bash
# View service details
gcloud run services describe asam-backend --region europe-west1

# Stream logs in real-time
gcloud run services logs tail asam-backend --region europe-west1

# List all revisions
gcloud run revisions list --service asam-backend --region europe-west1

# Check current traffic allocation
gcloud run services describe asam-backend --region europe-west1 --format="value(spec.traffic[].percent)"
```

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
