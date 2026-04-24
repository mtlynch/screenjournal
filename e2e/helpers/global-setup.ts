import fetch from "isomorphic-fetch";

import { FullConfig } from "@playwright/test";

async function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function waitForServer(baseURL: string, maxRetries = 60): Promise<void> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(baseURL + "/");
      if (response.ok) {
        // Server is up, wait a bit more to ensure it's fully ready
        await sleep(1000);
        return;
      }
    } catch {
      // Server not ready yet
    }
    await sleep(500);
  }
  throw new Error(`Server at ${baseURL} did not become ready`);
}

async function enablePerSessionDatabases(
  baseURL: string,
  maxRetries = 5
): Promise<void> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(baseURL + "/api/debug/db/per-session", {
        method: "POST",
      });
      if (response.ok) {
        // Wait a bit to ensure the setting is applied
        await sleep(500);
        return;
      }
      console.error(
        `Failed to enable per-session databases (attempt ${i + 1}): ${response.status}`
      );
    } catch (error) {
      console.error(
        `Error enabling per-session databases (attempt ${i + 1}):`,
        error
      );
    }
    await sleep(500);
  }
  throw new Error(`Failed to enable per-session databases after ${maxRetries} attempts`);
}

async function globalSetup(config: FullConfig) {
  const { baseURL } = config.projects[0].use;

  // Wait for server to be ready
  await waitForServer(baseURL as string);

  // Enable per-session databases so that test results stay independent
  await enablePerSessionDatabases(baseURL as string);

  // Verify per-session mode is working by making a test request
  // that should return a new db-token cookie
  const verifyResponse = await fetch(baseURL + "/");
  const setCookie = verifyResponse.headers.get("set-cookie");
  if (!setCookie?.includes("db-token=")) {
    throw new Error(
      "Per-session mode verification failed: no db-token cookie set"
    );
  }

  // Additional wait to ensure everything is stable
  await sleep(500);
}

export default globalSetup;
