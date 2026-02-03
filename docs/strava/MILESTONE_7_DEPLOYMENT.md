# Milestone 7: Deployment Preparation

**Goal:** Prepare for production deployment

---

## Files Summary (Milestone 7)

**Infrastructure Files to Modify:**
- `backend/cmd/api/infra/main.tf` - Add Strava secrets to Cloud Run
- `backend/cmd/api/infra/variables.tf` - Add Strava variable definitions
- `backend/cmd/api/infra/terraform.tfvars.example` - Document Strava vars

**Environment Files to Update:**
- Production `.env` files with real Strava credentials
- Verify all `.env.example` files are complete

**Configuration Reviews:**
- CORS settings in `backend/cmd/api/main.go`
- Rate limiting configurations
- WebSocket connection limits
- Session cleanup intervals

**No new feature code:**
- All feature development complete in M1-M6

**Focus Areas:**
- Security review (no token logging)
- Monitoring and metrics
- Deployment and rollback procedures

---

## Tasks

### 7.1 - Environment Configuration
- [ ] Add production Strava OAuth app
- [ ] Configure production callback URL
- [ ] Set `STRAVA_DEBUG=false` in production
- [ ] Add Strava credentials to Cloud Run secrets

**Files to Modify:**
- `backend/cmd/api/infra/main.tf` (add secrets)
- Production environment configuration

### 7.2 - Security Review
- [ ] Ensure tokens are never logged in production
- [ ] Validate CSRF state tokens properly
- [ ] Check WebSocket authentication
- [ ] Review CORS configuration

### 7.3 - Monitoring & Logging
- [ ] Add metrics for OAuth success/failure rates
- [ ] Add metrics for import success/failure rates
- [ ] Monitor rate limit usage
- [ ] Set up alerts for errors

### 7.4 - Rollout Plan
- [ ] Deploy backend with feature flag disabled
- [ ] Test in production with internal users
- [ ] Enable for beta testers (specific cities)
- [ ] Full public rollout

### 7.5 - Rollback Plan
- [ ] Document how to disable feature (environment variable)
- [ ] Ensure existing functionality unaffected
- [ ] Keep test tool available for debugging

**Validation:**
```bash
# Terraform validation
cd backend/cmd/api/infra
terraform plan

# Deploy to staging
# Test OAuth flow in staging
# Monitor logs and metrics
```
