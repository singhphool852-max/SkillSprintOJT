
let configuredUrl = process.env.NEXT_PUBLIC_API_URL;

// If the environment variable is missing, empty, or just a slash, enforce the production backend URL
if (!configuredUrl || configuredUrl.trim() === "" || configuredUrl === "/") {
  configuredUrl = "https://skillsprintojt.onrender.com";
  if (typeof window !== 'undefined') {
    console.warn("NEXT_PUBLIC_API_URL was invalid or missing. Enforcing hardcoded production backend URL.");
  }
}

// Remove trailing slash if present to prevent double slashes in routes
export const API_URL = configuredUrl.replace(/\/$/, "");

// Ensure WebSocket uses secure protocol (wss) if API is HTTPS
export const WS_BASE = API_URL?.startsWith("https") 
  ? API_URL.replace("https", "wss") 
  : API_URL?.replace("http", "ws") || "";
