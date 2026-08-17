import type { CreateClientConfig } from "./api/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  // Same-origin: embedded UI and Next rewrite both serve /api on this origin.
  baseUrl: "",
});
