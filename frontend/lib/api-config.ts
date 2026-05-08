
// Production backend URL — hardcoded to prevent Amplify build-time env var issues.
// .env.local is gitignored and does NOT exist on AWS Amplify,
// so process.env.NEXT_PUBLIC_API_URL would be undefined at build time.
// Next.js inlines env vars at build, making runtime fallbacks unreliable.
export const API_URL = "https://skillsprintojt.onrender.com";

// Ensure WebSocket uses secure protocol (wss) if API is HTTPS
export const WS_BASE = API_URL.startsWith("https") 
  ? API_URL.replace("https", "wss") 
  : API_URL.replace("http", "ws");
