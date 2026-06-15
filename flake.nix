{
  description = "Dev environment for ScreenJournal";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";

    # Use https://www.nixhub.io/ to find the right nixpkgs commit for the
    # specific package version we want.

    # 1.26.1 release
    go-nixpkgs.url = "github:NixOS/nixpkgs/e607cb5360ff1234862ac9f8839522becb853bb9";

    # 3.51.1 release
    sqlite-nixpkgs.url = "github:NixOS/nixpkgs/13868c071cc73a5e9f610c47d7bb08e5da64fdd5";

    # 24.12.0 release
    nodejs-nixpkgs.url = "github:NixOS/nixpkgs/af84f9d270d404c17699522fab95bbf928a2d92f";

    # 0.11.0 release
    shellcheck-nixpkgs.url = "github:NixOS/nixpkgs/1d4c88323ac36805d09657d13a5273aea1b34f0c";

    # 3.5.0 release
    sqlfluff-nixpkgs.url = "github:NixOS/nixpkgs/8b6600824693a9c706ef09bd86711ca393703466";

    # 1.57.0
    playwright-nixpkgs.url = "github:NixOS/nixpkgs/5f02c91314c8ba4afe83b256b023756412218535";

    # 0.3.209 release
    flyctl-nixpkgs.url = "github:NixOS/nixpkgs/1d4c88323ac36805d09657d13a5273aea1b34f0c";

    # 0.3.13 release
    litestream-nixpkgs.url = "github:NixOS/nixpkgs/a343533bccc62400e8a9560423486a3b6c11a23b";
  };

  outputs = {
    self,
    flake-utils,
    go-nixpkgs,
    sqlite-nixpkgs,
    nodejs-nixpkgs,
    shellcheck-nixpkgs,
    sqlfluff-nixpkgs,
    playwright-nixpkgs,
    flyctl-nixpkgs,
    litestream-nixpkgs,
  } @ inputs:
    flake-utils.lib.eachDefaultSystem (system: let
      gopkg = go-nixpkgs.legacyPackages.${system};
      go = gopkg.go_1_26;
      buildGoModule = gopkg.buildGoModule.override {inherit go;};
      sqlite = sqlite-nixpkgs.legacyPackages.${system}.sqlite;
      nodepkgs = nodejs-nixpkgs.legacyPackages.${system};
      nodejs = nodepkgs.nodejs_24;
      pnpm = nodepkgs.pnpm_10.override {inherit nodejs;};
      shellcheck = shellcheck-nixpkgs.legacyPackages.${system}.shellcheck;
      sqlfluff = sqlfluff-nixpkgs.legacyPackages.${system}.sqlfluff;
      playwright = playwright-nixpkgs.legacyPackages.${system}.playwright-driver.browsers;
      flyctl = flyctl-nixpkgs.legacyPackages.${system}.flyctl;
      litestream = litestream-nixpkgs.legacyPackages.${system}.litestream;
      git = gopkg.git;
      dockerTools = gopkg.dockerTools;

      # Fonts for Playwright browser tests.
      fontsConf = nodepkgs.makeFontsConf {
        fontDirectories = [nodepkgs.dejavu_fonts];
      };

      # Go static analysis tools.
      go-tools = gopkg.go-tools.override {inherit buildGoModule;};
      errcheck = gopkg.errcheck.override {inherit buildGoModule;};
      go-critic = gopkg.go-critic.override {inherit buildGoModule;};

      goVendorHash = "sha256-4I/5uWpHrOnICE32g8ssvLGS5F1TPKzXVUjUCrYgiUE=";

      pnpmDepsHash = "sha256-6fgI3SaUkEvMOPUu9gpZ+xGMZCoumHdJ45usg9R1Rgs=";

      appName = "screenjournal";
      appNameDev = "${appName}-dev";

      pnpmDeps = pnpm.fetchDeps {
        pname = "${appName}-pnpm-deps";
        version = "0.0.0";
        src = nodepkgs.lib.cleanSource ./.;
        fetcherVersion = 2;
        hash = pnpmDepsHash;
      };

      appPackage = buildGoModule {
        pname = appName;
        version = "0.0.1";
        src = gopkg.lib.cleanSource ./.;
        vendorHash = goVendorHash;
        subPackages = ["cmd/screenjournal"];
        env.CGO_ENABLED = "0";
        tags = ["netgo" "sqlite_omit_load_extension"];
        ldflags = ["-s" "-w"];
      };

      appPackageDev = buildGoModule {
        pname = appNameDev;
        version = "0.0.1";
        src = gopkg.lib.cleanSource ./.;
        vendorHash = goVendorHash;
        subPackages = ["cmd/screenjournal"];
        env.CGO_ENABLED = "0";
        tags = ["netgo" "sqlite_omit_load_extension" "dev"];
        ldflags = ["-s" "-w"];
        postInstall = ''
          mv "$out/bin/screenjournal" "$out/bin/${appNameDev}"
        '';
      };

      mkBuildStep = {
        name,
        command,
        extraInputs ? [],
        setup ? "",
        extraAttrs ? {},
      }:
        gopkg.stdenvNoCC.mkDerivation ({
            pname = name;
            version = "0.0.0";
            src = gopkg.lib.cleanSource ./.;
            nativeBuildInputs = [gopkg.bash] ++ extraInputs;
            buildPhase = ''
              runHook preBuild

              export HOME="$TMPDIR/home"
              mkdir -p "$HOME"

              export CI=1

              patchShebangs ./dev-scripts
              ${setup}
              ${command}

              runHook postBuild
            '';
            installPhase = ''
              mkdir -p "$out"
              echo "${name}" > "$out/done"
            '';
          }
          // extraAttrs);
    in {
      packages = {
        "${appName}" = appPackage;
        "${appNameDev}" = appPackageDev;

        backend = appPackage;
        backend-dev = appPackageDev;

        go-tests = mkBuildStep {
          name = "go-tests";
          command = "./dev-scripts/run-go-tests";
          extraInputs = [
            go
            sqlite
            gopkg.gcc
            gopkg.binutils
            go-tools
            errcheck
            go-critic
          ];
          setup = ''
            # Use pre-fetched Go modules in vendor format to avoid network access.
            cp -r ${appPackage.goModules} vendor
            chmod -R u+w vendor
            export GOFLAGS="-mod=vendor"

            # Create symlinks where run-go-tests expects Go tools.
            GO_BIN_DIR="$(go env GOPATH)/bin"
            mkdir -p "$GO_BIN_DIR"
            ln -sf ${go-critic}/bin/gocritic "$GO_BIN_DIR/go-critic"
            ln -sf ${go-tools}/bin/staticcheck "$GO_BIN_DIR/staticcheck"
            ln -sf ${errcheck}/bin/errcheck "$GO_BIN_DIR/errcheck"
          '';
        };

        check-bash = mkBuildStep {
          name = "check-bash";
          command = "./dev-scripts/check-bash";
          extraInputs = [git shellcheck];
          setup = ''
            git init -q
            git add -A
          '';
        };

        check-frontend = mkBuildStep {
          name = "check-frontend";
          command = "./dev-scripts/check-frontend";
          extraInputs = [nodejs pnpm pnpm.configHook];
          extraAttrs = {inherit pnpmDeps;};
        };

        check-go-formatting = mkBuildStep {
          name = "check-go-formatting";
          command = "./dev-scripts/check-go-formatting";
          extraInputs = [go];
        };

        check-trailing-newline = mkBuildStep {
          name = "check-trailing-newline";
          command = "./dev-scripts/check-trailing-newline";
          extraInputs = [git gopkg.coreutils gopkg.findutils gopkg.gnugrep];
          setup = ''
            git init -q
            git add -A
          '';
        };

        check-trailing-whitespace = mkBuildStep {
          name = "check-trailing-whitespace";
          command = "./dev-scripts/check-trailing-whitespace";
          extraInputs = [git gopkg.coreutils gopkg.findutils gopkg.gnugrep];
          setup = ''
            git init -q
            git add -A
          '';
        };

        docker-image = dockerTools.buildLayeredImage {
          name = appName;
          tag = "latest";
          contents = [
            gopkg.bashInteractive
            gopkg.coreutils
            gopkg.tzdata
            appPackage
            litestream
          ];
          extraCommands = ''
            mkdir -p app etc data
            cp ${./docker-entrypoint} app/docker-entrypoint
            chmod +x app/docker-entrypoint
            cp ${./litestream.yml} etc/litestream.yml

            ln -s ${appPackage}/bin/${appName} app/${appName}
            ln -s ${litestream}/bin/litestream app/litestream
          '';
          config = {
            Env = ["DB_PATH=/data/store.db"];
            Entrypoint = ["/app/docker-entrypoint"];
            WorkingDir = "/app";
          };
        };

        e2e-tests = mkBuildStep {
          name = "e2e-tests";
          command = ''
            pnpm exec playwright test \
              --workers=4 \
              --timeout=60000 \
              --grep-invert 'adds a new|views a TV show with an existing review|HTML tags in reviews|editing another user'
          '';
          extraInputs = [nodejs pnpm pnpm.configHook playwright appPackageDev];
          extraAttrs = {inherit pnpmDeps;};
          setup = ''
            export PLAYWRIGHT_BROWSERS_PATH=${playwright}
            export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true
            export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
            export SJ_TMDB_API=dummy

            # Configure fonts for headless browser rendering.
            export FONTCONFIG_FILE=${fontsConf}

            mkdir -p ./bin
            cp ${appPackageDev}/bin/${appNameDev} ./bin/${appNameDev}
          '';
        };

        lint-sql = mkBuildStep {
          name = "lint-sql";
          command = "./dev-scripts/lint-sql";
          extraInputs = [sqlfluff];
        };
      };

      devShells.default = gopkg.mkShell {
        packages = [
          gopkg.gotools
          gopkg.gopls
          gopkg.go-outline
          gopkg.gopkgs
          gopkg.gocode-gomod
          gopkg.godef
          gopkg.golint
          go
          sqlite
          nodejs
          pnpm
          shellcheck
          sqlfluff
          playwright
          flyctl
          litestream
        ];

        shellHook = ''
          # Avoid sharing GOPATH with other projects.
          PROJECT_NAME="$(basename "$PWD")"
          export GOPATH="$HOME/.local/share/go-workspaces/$PROJECT_NAME"

          # Override GOROOT set by Go tools (built with older Go) to match our
          # Go version.
          export GOROOT='${go}/share/go'

          export PLAYWRIGHT_BROWSERS_PATH=${playwright}
          export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true

          echo "shellcheck" "$(shellcheck --version | grep '^version:')"
          sqlfluff --version
          fly version | cut -d ' ' -f 1-3
          echo "litestream" "$(litestream version)"
          echo "node" "$(node --version)"
          echo "pnpm" "$(pnpm --version)"
          echo "sqlite" "$(sqlite3 --version | cut -d ' ' -f 1-2)"
          go version
        '';
      };

      formatter = gopkg.alejandra;
    });
}
