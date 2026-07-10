// Helper to determine the backend API base URL
// In development (running on port 3000), we hit the Go backend directly on port 8085,
// matching the exact hostname (localhost or 127.0.0.1) to avoid browser CORS / loopback security blocks.
// In production (served by Go on port 8085), we use relative paths.
export const getAPIBase = (): string => {
  if (typeof window !== "undefined") {
    if (window.location.port === "3000") {
      const hostname = window.location.hostname;
      const protocol = window.location.protocol;
      return `${protocol}//${hostname}:8085`;
    }
  }
  return "";
};
