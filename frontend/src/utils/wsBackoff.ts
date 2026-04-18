/**
 * getBackoffDelay calculates the next reconnection delay using exponential backoff
 * with a base of 1s and a maximum of 30s, plus a small random jitter.
 */
export const getBackoffDelay = (attempt: number): number => {
  const base = 1000;
  const max = 30000;
  const delay = Math.min(max, base * Math.pow(2, attempt));
  
  // Add random jitter (0-1000ms) to prevent thundering herd
  return delay + Math.random() * 1000;
};
