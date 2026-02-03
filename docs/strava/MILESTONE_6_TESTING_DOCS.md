# Milestone 6: Testing & Documentation

**Goal:** Comprehensive testing and documentation

---

## Files Summary (Milestone 6)

**New Test Files to Create:**
- `backend/internal/api/strava/handler_test.go` - Handler tests
- `backend/internal/api/strava/websocket_test.go` - WebSocket tests
- `frontends/form/src/lib/stores/strava.test.ts` - Store tests
- `frontends/form/src/lib/components/StravaImport.test.ts` - Component tests

**Documentation Files to Update:**
- `README.md` - Add Strava setup instructions
- `STRAVA_OAUTH_LEARNINGS.md` - Add any new discoveries
- `backend/README.md` - Document Strava API endpoints
- `frontends/form/README.md` - Document Strava import feature

**Test Files Already Exist (from M1-M2):**
- `backend/internal/strava/client_test.go`
- `backend/internal/strava/service_test.go`
- `backend/internal/strava/converter_test.go`

**No new feature files:**
- All feature code complete in M1-M5

---

## Tasks

### 6.1 - Backend Integration Tests
- [ ] Create `backend/internal/api/strava/handler_test.go`
- [ ] Test OAuth flow with mocked Strava API
- [ ] Test WebSocket import with multiple events
- [ ] Test error scenarios

### 6.2 - Frontend Component Tests
- [ ] Test Strava store (auth, logout, session management)
- [ ] Test StravaImport component (event selection)
- [ ] Test ImportProgress component (WebSocket messages)

### 6.3 - End-to-End Testing
- [ ] Test full OAuth flow
- [ ] Test importing single event
- [ ] Test importing multiple events (3+)
- [ ] Test error recovery (session expiry, rate limit)
- [ ] Test on different browsers (Chrome, Firefox, Safari)

**Manual Test Checklist:**
- [ ] OAuth flow works
- [ ] Only admin clubs are shown
- [ ] Events display correctly
- [ ] Multi-event import shows progress
- [ ] Edit links work after import
- [ ] Magic link emails are sent with event titles

### 6.4 - Update Documentation
- [ ] Add Strava setup instructions to README
- [ ] Document environment variables
- [ ] Add screenshots to `STRAVA_OAUTH_LEARNINGS.md`
- [ ] Create user guide for importing events

**Files to Modify:**
- `README.md`
- `STRAVA_OAUTH_LEARNINGS.md`
- `backend/README.md`
- `frontends/form/README.md`

### 6.5 - Performance Testing
- [ ] Test with user in 10+ clubs
- [ ] Test importing 10+ events simultaneously
- [ ] Monitor memory usage (session cleanup)
- [ ] Test WebSocket connection limits

**Validation:**
```bash
# Backend tests
cd backend
go test ./... -v -cover

# Frontend tests
cd frontends/form
npm run test
npm run check
npm run build

# E2E validation
# Manual testing with real Strava account
```
