import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { createServer } from "node:net";
import { spawn, type ChildProcess } from "node:child_process";

import { test as base, expect } from "@playwright/test";

import { startTmdbMock } from "./helpers/tmdbMock";

// 1x1 transparent PNG used to stub TMDB poster images so the browser never
// reaches out to image.tmdb.org during tests.
const STUB_POSTER_PNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+M8AAAMBAQDJ/pLvAAAAAElFTkSuQmCC",
  "base64"
);

type WorkerServer = {
  baseURL: string;
  restart: () => Promise<void>;
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

async function stopProcess(process: ChildProcess | null): Promise<void> {
  if (process === null || process.exitCode !== null) {
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

export const test = base.extend<
  {
    resetServer: void;
    stubPosters: void;
  },
  {
    workerServer: WorkerServer;
  }
>({
  workerServer: [
    async ({}, use, workerInfo) => {
      const tempDir = await mkdtemp(join(tmpdir(), "screenjournal-e2e-"));
      const port = await getFreePort();
      const baseURL = `http://127.0.0.1:${port}`;
      const tmdbMock = await startTmdbMock();
      let serverProcess: ChildProcess | null = null;
      let nextDatabaseID = 0;

      const restart = async (): Promise<void> => {
        await stopProcess(serverProcess);

        nextDatabaseID += 1;
        const dbPath = join(
          tempDir,
          `worker-${workerInfo.workerIndex}-test-${nextDatabaseID}.sqlite3`
        );
        const binaryPath = resolve(process.cwd(), "bin/screenjournal-dev");
        const stdoutChunks: Buffer[] = [];
        const stderrChunks: Buffer[] = [];

        serverProcess = spawn(binaryPath, ["--db", dbPath], {
          env: {
            ...process.env,
            PORT: String(port),
            SJ_TMDB_API: "dummy-api-key",
            SJ_TMDB_API_BASE_URL: tmdbMock.baseURL,
          },
          stdio: ["ignore", "pipe", "pipe"],
        });

        serverProcess.stdout?.on("data", (chunk: Buffer) => {
          stdoutChunks.push(chunk);
        });
        serverProcess.stderr?.on("data", (chunk: Buffer) => {
          stderrChunks.push(chunk);
        });

        try {
          await waitForServer(baseURL, 15_000);
        } catch (err) {
          const stdout = processOutput(stdoutChunks);
          const stderr = processOutput(stderrChunks);
          await stopProcess(serverProcess);
          serverProcess = null;
          throw new Error(
            `${String(err)}\nstdout:\n${stdout}\nstderr:\n${stderr}`
          );
        }

        if (serverProcess.exitCode !== null) {
          const stdout = processOutput(stdoutChunks);
          const stderr = processOutput(stderrChunks);
          throw new Error(
            `screenjournal-dev exited before test startup (code=${serverProcess.exitCode})\nstdout:\n${stdout}\nstderr:\n${stderr}`
          );
        }
      };

      await restart();
      await use({
        baseURL,
        restart,
      });
      await stopProcess(serverProcess);
      await tmdbMock.close();
      await rm(tempDir, { recursive: true, force: true });
    },
    { scope: "worker" },
  ],
  baseURL: async ({ workerServer }, use) => {
    await use(workerServer.baseURL);
  },
  resetServer: [
    async ({ workerServer }, use) => {
      await workerServer.restart();
      await use();
    },
    { auto: true },
  ],
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
