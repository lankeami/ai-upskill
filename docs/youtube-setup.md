# YouTube Upload Setup

One-time setup to enable the daily pipeline to upload videos to the Toiletpaper Press YouTube channel.

## 1. Create a Google Cloud project

1. Go to [console.cloud.google.com](https://console.cloud.google.com) → **New Project**
2. Name it (e.g., `ai-upskill`) and create it
3. Go to **APIs & Services → Library** → search for **YouTube Data API v3** → **Enable**

## 2. Create OAuth2 credentials

1. Go to **APIs & Services → Credentials → Create Credentials → OAuth 2.0 Client ID**
2. If prompted, configure the OAuth consent screen first:
   - User type: **External**
   - Add `lankeami@gmail.com` as a test user
3. Back in Credentials → Create Credentials → OAuth 2.0 Client ID:
   - Application type: **Desktop app**
   - Name: anything (e.g., `ai-upskill-upload`)
4. Click **Download JSON** — open it and note `client_id` and `client_secret`

## 3. Add credentials to `.env`

Add the following to your `.env` file in the project root:

```
YOUTUBE_CLIENT_ID=your_client_id_here
YOUTUBE_CLIENT_SECRET=your_client_secret_here
```

## 4. Install dependencies

```bash
source .venv/bin/activate
pip install google-api-python-client google-auth-oauthlib
```

## 5. Run the one-time auth flow

```bash
source .venv/bin/activate
python scripts/upload-youtube.py --auth
```

This opens a browser window. Sign in as `lankeami@gmail.com` and grant the YouTube upload permission. The token is saved to `.youtube-token.json` in the project root. **This only needs to be done once.** The token auto-refreshes on subsequent runs.

## Troubleshooting

- **"Access blocked"**: The OAuth app is in test mode. Make sure `lankeami@gmail.com` is listed as a test user in the OAuth consent screen.
- **"Token file missing"**: Run `python scripts/upload-youtube.py --auth` again.
- **Quota errors**: The YouTube Data API allows ~6 uploads per day (10,000 quota units, ~1,600 per upload). If you hit the limit, wait 24 hours.
