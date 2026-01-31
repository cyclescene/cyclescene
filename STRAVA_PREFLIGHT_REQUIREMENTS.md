# Strava Import Feature - Preflight Requirements

## ⚠️ MANDATORY: Read This Before Starting ANY Implementation

This document outlines **non-negotiable requirements** that MUST be met at every step of the Strava import feature implementation. Failure to follow these requirements will result in rejected work and wasted time.

---

## 🎯 Core Principles

### 1. **The Codebase MUST Always Build**

At **every single commit**, the codebase must build successfully:

```bash
# Backend MUST build without errors
cd backend
go build ./cmd/api
go build ./cmd/image-optimizer
go build ./cmd/scraperv2

# Frontend MUST build without errors
cd frontends/form
npm run build

cd frontends/pwa
npm run build

cd frontends/dashboard
npm run build

cd frontends/directory
npm run build
```

**Zero tolerance for build failures.**

If your changes break the build:
- ❌ Do NOT commit
- ❌ Do NOT move to the next task
- ❌ Do NOT ask "should I fix this later?"
- ✅ Fix it immediately before proceeding

---

### 2. **Zero TypeScript Errors**

TypeScript errors are **NOT warnings** - they are **ERRORS**.

```bash
# Frontend TypeScript check MUST pass
cd frontends/form
npm run check
# Expected output: "No errors found"
```

**Every TypeScript error must be resolved before committing.**

Common mistakes to avoid:
- ❌ Using `any` types to bypass errors
- ❌ Using `@ts-ignore` without justification
- ❌ Leaving unused variables/imports
- ❌ Missing type definitions for new data structures

**Acceptable TypeScript:**
- ✅ Properly typed interfaces and types
- ✅ Generic types with constraints
- ✅ Type guards for runtime validation
- ✅ `@ts-ignore` ONLY with explanatory comment for legitimate edge cases

---

### 3. **Debug Mode is NOT Optional**

Every feature MUST be debuggable via the `STRAVA_DEBUG` environment variable.

```bash
# Backend debug mode
STRAVA_DEBUG=true go run ./cmd/api

# This MUST enable verbose logging for:
# - OAuth flow (state generation, token exchange)
# - API requests to Strava (URLs, headers, responses)
# - Admin checks (which clubs, admin status)
# - Event fetching (club IDs, event counts)
# - Event conversion (field mappings, timezone conversions)
# - WebSocket messages (connections, progress updates)
# - Error details (stack traces, API responses)
```

**Debug logging requirements:**

1. **Use structured logging:**
   ```go
   // ✅ GOOD - Structured logging
   slog.Info("Fetching admin clubs", "athlete_id", athleteID, "session_id", sessionID)
   slog.Error("Failed to fetch events", "error", err, "club_id", clubID)

   // ❌ BAD - Unstructured logging
   log.Println("Fetching admin clubs for", athleteID)
   fmt.Printf("Error: %v\n", err)
   ```

2. **Check debug flag:**
   ```go
   // ✅ GOOD - Conditional debug logging
   if os.Getenv("STRAVA_DEBUG") == "true" {
       slog.Debug("OAuth token response", "response", tokenResp)
   }

   // ❌ BAD - No debug flag check (logs in production)
   slog.Info("OAuth token response", "response", tokenResp)
   ```

3. **Frontend debug logging:**
   ```typescript
   // ✅ GOOD - Console logging for debugging
   if (import.meta.env.PUBLIC_STRAVA_DEBUG === 'true') {
       console.log('[Strava] Fetching admin clubs', { sessionId });
   }

   // ❌ BAD - No debug flag
   console.log('Fetching admin clubs');
   ```

**What to log in debug mode:**
- All API requests/responses (URLs, status codes, headers)
- State transitions (OAuth flow, WebSocket connections)
- Data transformations (Strava → CycleScene mapping)
- Error contexts (what operation failed, why, with what data)
- Performance metrics (request durations, event counts)

---

## 🔍 Pre-Commit Checklist

Before EVERY commit, verify:

### Backend Changes
- [ ] `cd backend && go build ./cmd/api` - **MUST succeed**
- [ ] `cd backend && go test ./internal/strava/... -v` - **MUST pass (if tests exist)**
- [ ] `cd backend && go vet ./...` - **MUST have no warnings**
- [ ] Debug logging added for new functionality
- [ ] `STRAVA_DEBUG=true` enables verbose logging for new code
- [ ] No secrets/tokens logged (even in debug mode)

### Frontend Changes
- [ ] `cd frontends/form && npm run check` - **MUST have zero errors**
- [ ] `cd frontends/form && npm run build` - **MUST succeed**
- [ ] `cd frontends/form && npm run lint` - **MUST pass**
- [ ] TypeScript types defined for all new data structures
- [ ] No `any` types (unless absolutely necessary with justification)
- [ ] Debug logging added for new functionality
- [ ] All imports resolve correctly

### General
- [ ] Code follows existing patterns in the codebase
- [ ] Environment variables documented in `.env.example`
- [ ] No hardcoded URLs, credentials, or magic numbers
- [ ] Error handling in place (no silent failures)
- [ ] Comments explain "why", not "what"

---

## 🚨 Common Pitfalls to Avoid

### 1. **"It builds on my machine"**
- Always build the **entire** backend (`go build ./cmd/api`, not just `go run`)
- Run full TypeScript check (`npm run check`), not just hot reload
- Test with a clean `node_modules` (`rm -rf node_modules && npm install`)

### 2. **"I'll add debug logging later"**
- ❌ **NO.** Debug logging is part of the implementation, not an afterthought
- Add debug logs as you write code, not after
- Test that `STRAVA_DEBUG=true` actually shows your logs

### 3. **"TypeScript errors are just warnings"**
- ❌ **WRONG.** TypeScript errors will break production builds
- Fix them immediately, don't accumulate technical debt
- If you don't understand a type error, ask - don't bypass with `any`

### 4. **"I'll fix tests later"**
- ❌ **NO.** If you break existing tests, fix them in the same commit
- If you add new functionality, add tests for it
- Don't commit broken tests

### 5. **"Debug mode is too verbose"**
- ✅ **THAT'S THE POINT.** Debug mode should be extremely verbose
- Production should be quiet (INFO level only)
- Debug mode (`STRAVA_DEBUG=true`) should show everything

---

## 📋 Milestone Completion Criteria

Before marking a milestone as complete, ALL of the following must be true:

### Technical Requirements
- [ ] All backend services build successfully
- [ ] All frontend apps build successfully
- [ ] Zero TypeScript errors across all frontends
- [ ] All existing tests pass
- [ ] New tests added for new functionality
- [ ] `go vet` passes with no warnings

### Debug Requirements
- [ ] `STRAVA_DEBUG=true` enables verbose logging
- [ ] All new functions have debug log points
- [ ] Errors include context (what failed, why, with what data)
- [ ] WebSocket messages logged in debug mode
- [ ] API requests/responses logged in debug mode

### Code Quality
- [ ] No commented-out code
- [ ] No TODO comments without GitHub issues
- [ ] No hardcoded values (use env vars or constants)
- [ ] Error handling for all external calls (Strava API, database)
- [ ] Consistent code style with existing codebase

### Documentation
- [ ] Environment variables added to `.env.example`
- [ ] Milestone section in `STRAVA_IMPORT_CHECKLIST.md` updated
- [ ] Any new learnings added to `STRAVA_OAUTH_LEARNINGS.md`
- [ ] Commit message describes changes clearly

---

## 🎓 Working Incrementally

### The Right Way to Work

1. **Read the milestone tasks** in `STRAVA_IMPORT_CHECKLIST.md`
2. **Implement ONE task** at a time (not the whole milestone)
3. **Add debug logging** as you write the code
4. **Run build and type checks** immediately after finishing the task
5. **Fix any errors** before moving on
6. **Test the functionality** with `STRAVA_DEBUG=true`
7. **Commit** with a clear message
8. **Move to next task**

### Example: Milestone 1, Task 1.1

```bash
# 1. Implement GetAthleteClubs() method
# 2. Add debug logging:
if os.Getenv("STRAVA_DEBUG") == "true" {
    slog.Debug("Fetching athlete clubs", "athlete_id", athleteID)
}

# 3. Build immediately
cd backend
go build ./cmd/api
# ✅ Success? Continue. ❌ Errors? Fix them NOW.

# 4. Test manually
STRAVA_DEBUG=true go run ./cmd/api
# Trigger the function, check logs appear

# 5. Commit
git add backend/internal/strava/client.go
git commit -m "feat(strava): implement GetAthleteClubs method with debug logging"

# 6. Move to next method (GetClubDetails)
```

---

## 🛠️ Debug Mode Testing

Before committing ANY code, test debug mode works:

### Backend Debug Test
```bash
cd backend

# 1. Build
go build ./cmd/api

# 2. Run with debug enabled
STRAVA_DEBUG=true go run ./cmd/api

# 3. Trigger your new functionality (OAuth, import, etc.)

# 4. Verify logs appear in console:
# Expected output should include:
# - [DEBUG] messages for your new code
# - Structured fields (athlete_id, club_id, etc.)
# - Error contexts if failures occur

# 5. Run WITHOUT debug flag
go run ./cmd/api

# 6. Verify debug logs do NOT appear (only INFO/WARN/ERROR)
```

### Frontend Debug Test
```bash
cd frontends/form

# 1. Add to .env
PUBLIC_STRAVA_DEBUG=true

# 2. Run dev server
npm run dev

# 3. Open browser console, trigger functionality

# 4. Verify logs appear with [Strava] prefix

# 5. Remove debug flag, verify logs disappear
```

---

## ❌ Automatic Rejection Criteria

Your work will be **immediately rejected** if:

1. The backend doesn't build (`go build ./cmd/api` fails)
2. The frontend doesn't build (`npm run build` fails)
3. TypeScript errors exist (`npm run check` shows errors)
4. No debug logging added for new functionality
5. `STRAVA_DEBUG=true` doesn't show verbose logs
6. Existing tests broken without fixes
7. Secrets/tokens logged (even in debug mode)
8. Code committed with `TODO` or `FIXME` comments
9. Environment variables not documented in `.env.example`

---

## ✅ Success Indicators

You're on the right track if:

1. ✅ Every commit builds successfully
2. ✅ Every commit passes type checks
3. ✅ Every function has debug log points
4. ✅ `STRAVA_DEBUG=true` shows detailed operation logs
5. ✅ Error messages include context
6. ✅ No TypeScript `any` types (or justified with comments)
7. ✅ Tests exist and pass
8. ✅ Code follows existing patterns
9. ✅ Commits are small and focused
10. ✅ Documentation kept up to date

---

## 🎯 Remember

> **"If it doesn't build, it doesn't exist."**
>
> **"If it can't be debugged, it can't be fixed."**
>
> **"If TypeScript complains, listen."**

These requirements exist to:
- Prevent wasted time debugging broken code
- Ensure the feature is maintainable
- Make troubleshooting in production possible
- Keep the codebase healthy

**No shortcuts. No exceptions.**

---

## 📞 When in Doubt

If you're unsure about:
- How to fix a build error → Stop and ask
- Whether a TypeScript error matters → It matters, fix it
- How much to log in debug mode → Log more, not less
- Whether to commit broken code → Never. Fix it first.

**Ask questions early, not after committing broken code.**

---

**Last Updated:** 2026-01-31
**Status:** MANDATORY READING before starting implementation
**Branch:** `feature/stravaimport`
