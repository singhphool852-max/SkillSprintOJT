
export const API_URL = process.env.NEXT_PUBLIC_API_URL;

if (!API_URL && typeof window !== 'undefined') {
  console.warn("NEXT_PUBLIC_API_URL is not defined. API calls will likely fail.");
}

// Ensure WebSocket uses secure protocol (wss) if API is HTTPS
export const WS_BASE = API_URL?.startsWith("https") 
  ? API_URL.replace("https", "wss") 
  : API_URL?.replace("http", "ws") || "";
