// Playwright fixtures that manage the app server and its SQLite database.
//
// Isolation model: every test gets its own database, server process, and port,
// so tests never share state and can run fully in parallel.

import { spawn, type ChildProcess } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { createServer } from "node:net";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";

import { test as base, expect } from "@playwright/test";

import { startTmdbMock } from "./helpers/tmdbMock";

// 1x1 transparent PNG used to stub TMDB poster images so the browser never
// reaches out to image.tmdb.org during tests.
const STUB_POSTER_PNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+M8AAAMBAQDJ/pLvAAAAAElFTkSuQmCC",
  "base64"
);

type AppServer = {
  baseURL: string;
  process: ChildProcess;
};

async function getFreePort(): Promise<number> {
  return new Promise((resolvePort, reject) => {
    const server = createServer();
    server.on("error", reject);
    server.listen(0, "127.0.0.1", () => {
      const address = server.address();
      if (address === null || typeof address === "string") {
        server.close(() => reject(new Error("failed to choose free port")));
        return;
      }
      const { port } = address;
      server.close((err) => {
        if (err) {
          reject(err);
          return;
        }
        resolvePort(port);
      });
    });
  });
}

async function waitForServer(
  baseURL: string,
  timeoutMs: number
): Promise<void> {
  const deadline = Date.now() + timeoutMs;
  let lastError: unknown;

  while (Date.now() < deadline) {
    try {
      // Any HTTP status proves the server is up, so don't follow redirects,
      // as a redirect to a broken location would misread readiness as failure.
      const response = await fetch(baseURL, { redirect: "manual" });
      if (response.status > 0) {
        return;
      }
    } catch (err) {
      lastError = err;
    }
    await new Promise((resolveDelay) => setTimeout(resolveDelay, 100));
  }

  throw new Error(`server did not start at ${baseURL}: ${String(lastError)}`);
}

function processOutput(chunks: Buffer[]): string {
  return Buffer.concat(chunks).toString("utf8");
}

async function stopServer(process: ChildProcess): Promise<void> {
  if (process.exitCode !== null) {
    return;
  }

  await new Promise<void>((resolveStop) => {
    process.once("exit", () => resolveStop());
    process.kill("SIGTERM");
    setTimeout(() => {
      if (process.exitCode === null) {
        process.kill("SIGKILL");
      }
    }, 5_000);
  });
}

// Spawns bin/screenjournal-dev against the given database and TMDB mock, then
// waits until it responds over HTTP. On startup failure, stops the process and
// throws with its captured stdout/stderr.
async function startServer(
  port: number,
  dbPath: string,
  tmdbBaseURL: string
): Promise<AppServer> {
  const baseURL = `http://127.0.0.1:${port}`;
  const binaryPath = resolve(process.cwd(), "bin/screenjournal-dev");
  const stdoutChunks: Buffer[] = [];
  const stderrChunks: Buffer[] = [];
  const serverProcess = spawn(binaryPath, ["--db", dbPath], {
    env: {
      ...process.env,
      PORT: String(port),
      SJ_TMDB_API: "dummy-api-key",
      SJ_TMDB_API_BASE_URL: tmdbBaseURL,
    },
    stdio: ["ignore", "pipe", "pipe"],
  });
  serverProcess.stdout?.on("data", (chunk: Buffer) => stdoutChunks.push(chunk));
  serverProcess.stderr?.on("data", (chunk: Buffer) => stderrChunks.push(chunk));
  try {
    await waitForServer(`${baseURL}/about`, 15_000);
  } catch (err) {
    await stopServer(serverProcess);
    throw new Error(
      `${String(err)}\nstdout:\n${processOutput(
        stdoutChunks
      )}\nstderr:\n${processOutput(stderrChunks)}`
    );
  }
  // Catch a server that answered the health check but died right after.
  if (serverProcess.exitCode !== null) {
    throw new Error(
      `screenjournal-dev exited before test startup (code=${
        serverProcess.exitCode
      })\nstdout:\n${processOutput(stdoutChunks)}\nstderr:\n${processOutput(
        stderrChunks
      )}`
    );
  }
  return { baseURL, process: serverProcess };
}

export const test = base.extend<{
  server: AppServer;
  stubPosters: void;
}>({
  server: [
    async ({}, use) => {
      const tempDir = await mkdtemp(join(tmpdir(), "screenjournal-e2e-"));
      const port = await getFreePort();
      const dbPath = join(tempDir, "e2e-test.sqlite3");
      const tmdbMock = await startTmdbMock();
      const server = await startServer(port, dbPath, tmdbMock.baseURL);

      // The test body runs during this call.
      await use(server);

      await stopServer(server.process);
      await tmdbMock.close();
      await rm(tempDir, { recursive: true, force: true });
    },
    // auto makes the fixture run for every test, even tests that never
    // reference it directly.
    { auto: true },
  ],
  // Point Playwright's built-in baseURL option at this test's server so that
  // page.goto("/") and friends hit the right port.
  baseURL: async ({ server }, use) => await use(server.baseURL),
  stubPosters: [
    async ({ page }, use) => {
      await page.route("**image.tmdb.org**", (route) =>
        route.fulfill({
          status: 200,
          contentType: "image/png",
          body: STUB_POSTER_PNG,
        })
      );
      await use();
    },
    { auto: true },
  ],
});

export { expect };
