# GitHub Actions Secrets Configuration

This document lists all the required secrets that need to be configured in GitHub for the CI/CD pipeline to work properly.

## Required Secrets

### Google Cloud Platform Secrets

1. **GCP_PROJECT_ID**
   - Your Google Cloud Project ID
   - Example: `my-project-12345`

2. **GCP_SA_KEY**
   - Service Account JSON key with permissions to deploy to Cloud Run
   - This should be the entire JSON content of the service account key file

### Database Secrets (Aiven PostgreSQL)

3. **AIVEN_DB_HOST**
   - Aiven PostgreSQL host
   - Example: `pg-asam-asam-backend-db.l.aivencloud.com`

4. **AIVEN_DB_PORT**
   - Aiven PostgreSQL port
   - Example: `14276`

5. **AIVEN_DB_USER**
   - Database user
   - Example: `avnadmin`

6. **AIVEN_DB_PASSWORD**
   - Database password

7. **AIVEN_DB_NAME**
   - Database name
   - Example: `defaultdb`

### Application Secrets

8. **JWT_ACCESS_SECRET**
   - Secret key for JWT access tokens
   - Should be a strong, random string (32+ characters)

9. **JWT_REFRESH_SECRET**
   - Secret key for JWT refresh tokens
   - Should be a strong, random string (32+ characters)

10. **ADMIN_USER**
    - Administrator username for protected endpoints
    - Example: `admin`

11. **ADMIN_PASSWORD**
    - Administrator password
    - Should be a strong password

### Optional Secrets

12. **SMTP_USER** (Optional)
    - SMTP username for email notifications
    - If not set, defaults to `noreply@asam.org`

13. **SMTP_PASSWORD** (Optional)
    - SMTP password for email notifications
    - If not set, defaults to `temp-smtp-pass`

## How to Configure Secrets

1. Go to your GitHub repository
2. Click on "Settings" → "Secrets and variables" → "Actions"
3. Click "New repository secret"
4. Add each secret with its name and value

## Generating Secure Secrets

For JWT secrets, you can generate secure random strings using:

```bash
# Linux/Mac
openssl rand -base64 32

# Or using Node.js
node -e "console.log(require('crypto').randomBytes(32).toString('base64'))"
```

## Service Account Permissions

The Google Cloud Service Account needs the following permissions:
- Cloud Run Admin
- Service Account User
- Artifact Registry Writer
- Cloud Build Service Account

You can create a service account with:

```bash
# Create service account
gcloud iam service-accounts create github-actions \
    --display-name="GitHub Actions"

# Grant permissions
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
    --member="serviceAccount:github-actions@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
    --role="roles/run.admin"

gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
    --member="serviceAccount:github-actions@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
    --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
    --member="serviceAccount:github-actions@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
    --role="roles/artifactregistry.writer"

# Create and download key
gcloud iam service-accounts keys create key.json \
    --iam-account=github-actions@YOUR-PROJECT-ID.iam.gserviceaccount.com
```

Copy the contents of `key.json` to the `GCP_SA_KEY` secret.
