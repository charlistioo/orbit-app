# Deploying ORBIT for external testing

This deploys `backend/` (Go + PostgreSQL) to Railway and `web/`
(Next.js) to Vercel, using their free tiers and default subdomains -
good enough for letting outside testers try the app on a real,
always-on URL instead of your laptop.

## 0. Prerequisites

- A GitHub account, with this repo pushed to it (see step 1).
- A Railway account (railway.app) and a Vercel account (vercel.com) -
  both let you sign in with GitHub directly, no separate password
  needed.

## 1. Push this repo to GitHub

This folder isn't a git repository yet. From `C:\ai-agents-work\vibe-coded-web\orbit`:

```
git init
git add .
git commit -m "Initial commit"
```

Then create a new (private is fine) repository on github.com, and push:

```
git remote add origin https://github.com/<your-username>/<repo-name>.git
git branch -M main
git push -u origin main
```

## 2. Backend on Railway

1. On railway.app, **New Project → Deploy from GitHub repo** → pick
   this repo.
2. Railway will ask for a root directory - set it to `backend`.
3. It should detect `backend/Dockerfile` automatically and build with
   it (this repo's Dockerfile builds both the API server and the
   migration runner, and runs migrations automatically on every
   deploy before starting the server - see `backend/entrypoint.sh`).
4. **Add a PostgreSQL database**: in the same Railway project, "+ New"
   → Database → PostgreSQL. Railway automatically injects
   `DATABASE_URL` into your backend service - no need to copy/paste it.
5. In the backend service's **Variables** tab, add:
   - `GOOGLE_CLIENT_ID` - the same client ID already used locally
     (`595683522917-al9clne9te05q8gb55a1t8v2t0kvf39u.apps.googleusercontent.com`),
     or a new one if you'd rather keep production separate.
   - `SESSION_SECRET` - a **new** random value, not the local
     `dev-test-secret`. Generate one with `openssl rand -base64 48` (or
     ask me to generate one for you).
6. Deploy. Railway gives you a public URL like
   `https://orbit-backend-production.up.railway.app`. Verify with:
   `curl https://<that-url>/api/v1/health` → `{"status":"healthy"}`.

## 3. Frontend on Vercel

1. On vercel.com, **Add New → Project** → import the same GitHub repo.
2. Set **Root Directory** to `web`.
3. Vercel auto-detects Next.js - no build command changes needed.
4. In **Environment Variables**, add:
   - `NEXT_PUBLIC_API_BASE` = `https://<your-railway-url>/api/v1`
   - `NEXT_PUBLIC_GOOGLE_CLIENT_ID` = the same client ID as the backend
5. Deploy. Vercel gives you a URL like `https://orbit-xyz.vercel.app`.

## 4. Register the new domain with Google

Google Identity Services enforces Authorized JavaScript origins per
OAuth client (the same requirement we hit setting this up locally):

1. Go to [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials).
2. Open the OAuth 2.0 Client ID used above.
3. Under **Authorized JavaScript origins**, add your Vercel URL exactly
   (e.g. `https://orbit-xyz.vercel.app`, no trailing slash).
4. Save, wait ~1-2 minutes for propagation.

## 5. Try it

Open your Vercel URL in a browser you're not already signed into for
this project, sign in with Google, and walk through onboarding. Share
the link with testers - each person gets their own account
automatically on first sign-in.

## Notes

- This production database is separate from the local
  `orbit-test-db3` Docker container - external testers' data never
  mixes with your local test data.
- Every `git push` to `main` auto-redeploys both Railway and Vercel -
  no manual redeploy steps needed after this initial setup.
- If you want a real custom domain later, both Railway and Vercel
  support adding one in their dashboard (and it needs adding to Google
  Cloud Console's authorized origins too, same as step 4).
