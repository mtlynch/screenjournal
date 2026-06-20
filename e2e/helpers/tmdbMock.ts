import { createServer, type Server } from "node:http";

// In-process mock of the TMDB API used by the e2e suite. The screenjournal
// backend calls TMDB server-side, so the test harness starts this server and
// points the backend at it via SJ_TMDB_API_BASE_URL. This keeps the e2e tests
// hermetic (no Internet) without baking fake data into the production binary.

type MockMovie = {
  id: number;
  title: string;
  releaseDate: string;
  imdbID: string;
  posterPath: string;
};

type MockTvShow = {
  id: number;
  name: string;
  firstAirDate: string;
  imdbID: string;
  posterPath: string;
  seasonCount: number;
};

const MOVIES: MockMovie[] = [
  {
    id: 333287,
    title: "Slow Learners",
    releaseDate: "2015-08-19",
    imdbID: "tt2597718",
    posterPath: "/slow-learners.jpg",
  },
  {
    id: 745,
    title: "The Sixth Sense",
    releaseDate: "1999-08-06",
    imdbID: "tt0167404",
    posterPath: "/sixth-sense.jpg",
  },
  {
    id: 928344,
    title: "Weird: The Al Yankovic Story",
    releaseDate: "2022-09-08",
    imdbID: "tt17076046",
    posterPath: "/qcj2z13G0KjaIgc01ifiUKu7W07.jpg",
  },
  {
    id: 38,
    title: "Eternal Sunshine of the Spotless Mind",
    releaseDate: "2004-03-19",
    imdbID: "tt0338013",
    posterPath: "/eternal-sunshine.jpg",
  },
  {
    id: 544,
    title: "There's Something About Mary",
    releaseDate: "1998-07-15",
    imdbID: "tt0129387",
    posterPath: "/something-about-mary.jpg",
  },
  {
    id: 409,
    title: "The English Patient",
    releaseDate: "1996-11-14",
    imdbID: "tt0116209",
    posterPath: "/english-patient.jpg",
  },
  {
    id: 238,
    title: "The Godfather",
    releaseDate: "1972-03-14",
    imdbID: "tt0068646",
    posterPath: "/godfather.jpg",
  },
  {
    id: 10663,
    title: "The Waterboy",
    releaseDate: "1998-11-06",
    imdbID: "tt0120484",
    posterPath: "/miT42qWYC4D0n2mXNzJ9VfhheWW.jpg",
  },
  {
    id: 11017,
    title: "Billy Madison",
    releaseDate: "1995-02-10",
    imdbID: "tt0112508",
    posterPath: "/iwk9pWR6MwTInEQc8Vw5vGHjeQ0.jpg",
  },
];

const TV_SHOWS: MockTvShow[] = [
  {
    id: 4608,
    name: "30 Rock",
    firstAirDate: "2006-10-11",
    imdbID: "tt0496424",
    posterPath: "/30-rock.jpg",
    seasonCount: 7,
  },
  {
    id: 1400,
    name: "Seinfeld",
    firstAirDate: "1989-07-05",
    imdbID: "tt0098904",
    posterPath: "/aCw8ONfyz3AhngVQa1E2Ss4KSUQ.jpg",
    seasonCount: 9,
  },
];

function matchesQuery(title: string, query: string): boolean {
  return title.toLowerCase().includes(query.toLowerCase());
}

// The backend recomputes a show's season count from this list, dropping seasons
// whose air date is not in the past, so every season needs a definitively-past
// air date and a non-zero episode count.
function buildSeasons(count: number) {
  const seasons = [];
  for (let seasonNumber = 1; seasonNumber <= count; seasonNumber++) {
    seasons.push({
      season_number: seasonNumber,
      name: `Season ${seasonNumber}`,
      episode_count: 10,
      air_date: "2000-01-01",
    });
  }
  return seasons;
}

export type TmdbMock = {
  baseURL: string;
  close: () => Promise<void>;
};

export async function startTmdbMock(): Promise<TmdbMock> {
  const server: Server = createServer((req, res) => {
    const requestURL = new URL(req.url ?? "", "http://127.0.0.1");
    const path = requestURL.pathname;
    const query = requestURL.searchParams.get("query") ?? "";

    const sendJSON = (body: unknown) => {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(body));
    };
    const sendNotFound = () => {
      res.writeHead(404, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ status_code: 34, status_message: "Not found" }));
    };

    if (path === "/search/movie") {
      sendJSON({
        results: MOVIES.filter((movie) => matchesQuery(movie.title, query)).map(
          (movie) => ({
            id: movie.id,
            title: movie.title,
            release_date: movie.releaseDate,
            poster_path: movie.posterPath,
          })
        ),
      });
      return;
    }

    if (path === "/search/tv") {
      sendJSON({
        results: TV_SHOWS.filter((show) => matchesQuery(show.name, query)).map(
          (show) => ({
            id: show.id,
            name: show.name,
            first_air_date: show.firstAirDate,
            poster_path: show.posterPath,
          })
        ),
      });
      return;
    }

    const externalIDsMatch = path.match(/^\/tv\/(\d+)\/external_ids$/);
    if (externalIDsMatch) {
      const show = TV_SHOWS.find((s) => s.id === Number(externalIDsMatch[1]));
      if (show === undefined) {
        sendNotFound();
        return;
      }
      sendJSON({ imdb_id: show.imdbID });
      return;
    }

    const tvMatch = path.match(/^\/tv\/(\d+)$/);
    if (tvMatch) {
      const show = TV_SHOWS.find((s) => s.id === Number(tvMatch[1]));
      if (show === undefined) {
        sendNotFound();
        return;
      }
      sendJSON({
        name: show.name,
        first_air_date: show.firstAirDate,
        poster_path: show.posterPath,
        seasons: buildSeasons(show.seasonCount),
      });
      return;
    }

    const movieMatch = path.match(/^\/movie\/(\d+)$/);
    if (movieMatch) {
      const movie = MOVIES.find((m) => m.id === Number(movieMatch[1]));
      if (movie === undefined) {
        sendNotFound();
        return;
      }
      sendJSON({
        title: movie.title,
        imdb_id: movie.imdbID,
        release_date: movie.releaseDate,
        poster_path: movie.posterPath,
      });
      return;
    }

    sendNotFound();
  });

  await new Promise<void>((resolveListen, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", () => resolveListen());
  });

  const address = server.address();
  if (address === null || typeof address === "string") {
    throw new Error("failed to start TMDB mock server");
  }
  const baseURL = `http://127.0.0.1:${address.port}`;

  return {
    baseURL,
    close: () =>
      new Promise<void>((resolveClose, reject) =>
        server.close((err) => (err ? reject(err) : resolveClose()))
      ),
  };
}
