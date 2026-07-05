// TODO: no deployment target is fixed yet (backend/README.md only documents local dev via
// compose.yaml). Replace with a real build-time value once that's decided.
export const API_BASE_URL = 'http://localhost:8080'

// VAPID public key used to authorize this client's push subscriptions (LOGIC-004, FEAT-03/FEAT-06
// in 6-development/FEATURES_IMPLEMENTATION_PLAN.md). Hardcoded like API_BASE_URL above — this
// project doesn't use client-side .env — and must match the private key the backend signs with.
// Not a secret: VAPID public keys are meant to be exposed to the client by design.
export const VAPID_PUBLIC_KEY = 'BI_GG7M7ixlJflnMuhL-dM2xgmH7k_SCB2TwEb3NhnVcaL37mmL1ibRWp6F4sxU4cYw-563a8fsG-qEooA31dSg'
