{
  description = "Dev environment for ScreenJournal";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";

    # Use https://www.nixhub.io/ to find the exact nixpkgs reference for exact
    # package versions.

    # 1.26.1 release
    go-nixpkgs.url = "github:NixOS/nixpkgs/e607cb5360ff1234862ac9f8839522becb853bb9";

    # 3.44.2 release
    sqlite-nixpkgs.url = "github:NixOS/nixpkgs/5ad9903c16126a7d949101687af0aa589b1d7d3d";

    # 24.15.0 release
    nodejs-nixpkgs.url = "github:NixOS/nixpkgs/9eac87a12312b8f60dd52e1c6e1a265f6fc7f5fc";

    # 10.25.0 release
    pnpm-nixpkgs.url = "github:NixOS/nixpkgs/af84f9d270d404c17699522fab95bbf928a2d92f";

    # 0.10.0 release
    shellcheck-nixpkgs.url = "github:NixOS/nixpkgs/4ae2e647537bcdbb82265469442713d066675275";

    # 0.14.0 release
    gci-nixpkgs.url = "github:NixOS/nixpkgs/29916453413845e54a65b8a1cf996842300cd299";

    # 3.3.0 release
    sqlfluff-nixpkgs.url = "github:NixOS/nixpkgs/bf689c40d035239a489de5997a4da5352434632e";

    # 1.58.2
    playwright-nixpkgs.url = "github:NixOS/nixpkgs/9cded172058da9dfa6b3227d46ff30ea2023698a";

    # 0.4.59 release
    flyctl-nixpkgs.url = "github:NixOS/nixpkgs/5a722a7155bfc9fbe657f28d26b71860d95324bc";

    # 0.3.13 release
    litestream-nixpkgs.url = "github:NixOS/nixpkgs/a343533bccc62400e8a9560423486a3b6c11a23b";

    stapelberg-nix = {
      url = "github:stapelberg/nix";
      inputs.nixpkgs.follows = "go-nixpkgs";
    };
  };

  outputs = {
    self,
    flake-utils,
    go-nixpkgs,
    sqlite-nixpkgs,
    nodejs-nixpkgs,
    pnpm-nixpkgs,
    shellcheck-nixpkgs,
    gci-nixpkgs,
    sqlfluff-nixpkgs,
    playwright-nixpkgs,
    flyctl-nixpkgs,
    litestream-nixpkgs,
    stapelberg-nix,
  } @ inputs:
    flake-utils.lib.eachDefaultSystem (system: let
      # Use Stapelberg's overlay so Nix builds stamp Go binaries with the
      # flake revision and modification time, matching Go's VCS build info.
      gopkg = import go-nixpkgs {
        inherit system;
        overlays = [
          (final: prev: {
            buildGoModule = prev.buildGoModule.override {go = final.go_1_26;};
          })
          stapelberg-nix.overlays.goVcsStamping
        ];
      };
      go = gopkg.go_1_26;
      buildGoModule = gopkg.buildGoModule;
      sqlite = sqlite-nixpkgs.legacyPackages.${system}.sqlite;
      nodepkgs = nodejs-nixpkgs.legacyPackages.${system};
      nodejs = nodepkgs.nodejs_24;
      pnpm = pnpm-nixpkgs.legacyPackages.${system}.pnpm_10.override {inherit nodejs;};
      shellcheck = shellcheck-nixpkgs.legacyPackages.${system}.shellcheck;
      gci = gci-nixpkgs.legacyPackages.${system}.gci;
      sqlfluff = sqlfluff-nixpkgs.legacyPackages.${system}.sqlfluff;
      playwright = playwright-nixpkgs.legacyPackages.${system}.playwright-driver.browsers;
      flyctl = flyctl-nixpkgs.legacyPackages.${system}.flyctl;
      litestream = litestream-nixpkgs.legacyPackages.${system}.litestream;
      git = gopkg.git;

      # Go static analysis tools.
      go-tools = gopkg.go-tools.override {inherit buildGoModule;}; # includes staticcheck
      errcheck = gopkg.errcheck.override {inherit buildGoModule;};
      go-critic = gopkg.go-critic.override {inherit buildGoModule;};

      # Fonts for Playwright browser tests.
      fontsConf = nodepkgs.makeFontsConf {
        fontDirectories = [nodepkgs.dejavu_fonts];
      };

      goVendorHash = "sha256-KEl0DtmecaSsWlejBhsAnQ6cXsb/1Vq1L6aQHH9nXJk=";

      goTestVendorHash = "sha256-qRMbjyi7YDBunlUJTokBJzodU7UIs54WeoHbo2xWstI=";

      pnpmDepsHash = "sha256-PD6Donpph1DJaQvuiurrhHOe4u2x4acVLygYDk1aIkY=";

      appName = "screenjournal";
      appNameDev = "${appName}-dev";

      backendSrc = gopkg.lib.fileset.toSource {
        root = ./.;
        fileset = gopkg.lib.fileset.unions [
          ./go.mod
          ./go.sum
          (gopkg.lib.fileset.fileFilter (
              file: file.hasExt "go" && !(gopkg.lib.hasSuffix "_test.go" file.name)
            )
            ./.)
          ./handlers/static
          ./handlers/templates
          ./store/sqlite/migrations
        ];
      };

      backendVcsMetadata =
        if self ? rev
        then {
          inherit (self) rev lastModified;
        }
        else if self ? dirtyRev
        then {
          rev = gopkg.lib.removeSuffix "-dirty" self.dirtyRev;
          inherit (self) lastModified;
          dirty = true;
        }
        else null;

      pnpmDeps = pnpm.fetchDeps {
        pname = "${appName}-pnpm-deps";
        version = "0.0.0";
        src = gopkg.lib.fileset.toSource {
          root = ./.;
          fileset = gopkg.lib.fileset.unions [
            ./package.json
            ./pnpm-lock.yaml
          ];
        };
        fetcherVersion = 2;
        hash = pnpmDepsHash;
      };

      appPackage = buildGoModule {
        pname = appName;
        version = "0.0.1";
        src = backendSrc;
        vcsMetadata = backendVcsMetadata;
        vendorHash = goVendorHash;
        subPackages = ["cmd/screenjournal"];
        env.CGO_ENABLED = "0";
        tags = ["netgo" "sqlite_omit_load_extension"];
        ldflags = ["-s" "-w"];
        postInstall = ''
          mv "$out/bin/screenjournal" "$out/bin/${appName}"
        '';
      };

      appPackageDev = buildGoModule {
        pname = appNameDev;
        version = "0.0.1";
        src = backendSrc;
        vcsMetadata = backendVcsMetadata;
        vendorHash = goVendorHash;
        subPackages = ["cmd/screenjournal"];
        env.CGO_ENABLED = "0";
        tags = ["netgo" "sqlite_omit_load_extension" "dev"];
        ldflags = ["-s" "-w"];
        postInstall = ''
          mv "$out/bin/screenjournal" "$out/bin/${appNameDev}"
        '';
      };

      testGoModules = buildGoModule {
        pname = "${appName}-test-modules";
        version = "0.0.0";
        src = gopkg.lib.cleanSource ./.;
        vendorHash = goTestVendorHash;
        subPackages = [];
        doCheck = false;
      };

      mkBuildStep = {
        name,
        command,
        src ? gopkg.lib.cleanSource ./. ,
        extraInputs ? [],
        setup ? "",
        extraAttrs ? {},
      }:
        gopkg.stdenvNoCC.mkDerivation ({
            pname = name;
            version = "0.0.0";
            inherit src;
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
            gci
          ];
          setup = ''
            # Use pre-fetched Go modules (vendor format) to avoid network access.
            cp -r ${testGoModules.goModules} vendor
            chmod -R u+w vendor
            export GOFLAGS="-mod=vendor"

            # Create symlinks where run-go-tests expects Go tools.
            export GOBIN="$(go env GOPATH)/bin"
            mkdir -p "$GOBIN"
            ln -sf ${go-critic}/bin/gocritic "$GOBIN/go-critic"
            ln -sf ${go-tools}/bin/staticcheck "$GOBIN/staticcheck"
            ln -sf ${errcheck}/bin/errcheck "$GOBIN/errcheck"
            ln -sf ${gci}/bin/gci "$GOBIN/gci"
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

        lint-sql = mkBuildStep {
          name = "lint-sql";
          command = "./dev-scripts/lint-sql";
          src = gopkg.lib.fileset.toSource {
            root = ./.;
            fileset = gopkg.lib.fileset.unions [
              ./dev-scripts/lint-sql
              (gopkg.lib.fileset.fileFilter (file: file.hasExt "sql") ./.)
            ];
          };
          extraInputs = [sqlfluff];
        };

        backend = appPackage;
        backend-dev = appPackageDev;

        check-frontend = mkBuildStep {
          name = "check-frontend";
          command = "./dev-scripts/check-frontend";
          src = gopkg.lib.fileset.toSource {
            root = ./.;
            fileset = gopkg.lib.fileset.intersection
              (gopkg.lib.fileset.gitTracked ./.)
              (gopkg.lib.fileset.unions [
                ./dev-scripts/check-frontend
                ./.prettierignore
                ./.prettierrc
                ./eslint.config.js
                ./package.json
                ./pnpm-lock.yaml
                (gopkg.lib.fileset.fileFilter (file:
                  file.hasExt "md"
                  || file.hasExt "js"
                  || file.hasExt "html"
                  || file.hasExt "css"
                  || file.hasExt "json"
                  || file.hasExt "yaml"
                  || file.hasExt "yml"
                  || file.hasExt "ts"
                ) ./.)
              ]);
          };
          extraInputs = [git nodejs pnpm pnpm.configHook];
          extraAttrs = {inherit pnpmDeps;};
          setup = ''
            git init -q
            git add -A
          '';
        };

        check-go-formatting = mkBuildStep {
          name = "check-go-formatting";
          command = "./dev-scripts/check-go-formatting";
          src = gopkg.lib.fileset.toSource {
            root = ./.;
            fileset = gopkg.lib.fileset.intersection
              (gopkg.lib.fileset.gitTracked ./.)
              (gopkg.lib.fileset.unions [
                ./dev-scripts/check-go-formatting
                (gopkg.lib.fileset.fileFilter (file: file.hasExt "go") ./.)
              ]);
          };
          extraInputs = [go];
        };

        check-go-test-packages = mkBuildStep {
          name = "check-go-test-packages";
          command = "./dev-scripts/check-go-test-packages";
          src = gopkg.lib.fileset.toSource {
            root = ./.;
            fileset = gopkg.lib.fileset.intersection
              (gopkg.lib.fileset.gitTracked ./.)
              (gopkg.lib.fileset.unions [
                ./dev-scripts/check-go-test-packages
                (gopkg.lib.fileset.fileFilter
                  (file: gopkg.lib.hasSuffix "_test.go" file.name)
                  ./.)
              ]);
          };
          extraInputs = [git gopkg.gawk];
          setup = ''
            git init -q
            git add -A
          '';
        };

        check-trailing-newline = mkBuildStep {
          name = "check-trailing-newline";
          command = "./dev-scripts/check-trailing-newline";
          src = gopkg.lib.fileset.toSource {
            root = ./.;
            fileset = gopkg.lib.fileset.gitTracked ./.;
          };
          extraInputs = [git gopkg.coreutils gopkg.findutils gopkg.gnugrep];
          setup = ''
            git init -q
            git add -A
          '';
        };

        check-trailing-whitespace = mkBuildStep {
          name = "check-trailing-whitespace";
          command = "./dev-scripts/check-trailing-whitespace";
          src = gopkg.lib.fileset.toSource {
            root = ./.;
            fileset = gopkg.lib.fileset.gitTracked ./.;
          };
          extraInputs = [git gopkg.coreutils gopkg.findutils gopkg.gnugrep];
          setup = ''
            git init -q
            git add -A
          '';
        };

        e2e-tests = mkBuildStep {
          name = "e2e-tests";
          command = "pnpm exec playwright test";
          src = gopkg.lib.fileset.toSource {
            root = ./.;
            fileset = gopkg.lib.fileset.intersection
              (gopkg.lib.fileset.gitTracked ./.)
              (gopkg.lib.fileset.unions [
                ./e2e
                ./playwright.config.ts
                ./package.json
                ./pnpm-lock.yaml
              ]);
          };
          extraInputs = [nodejs pnpm pnpm.configHook playwright appPackageDev];
          extraAttrs = {inherit pnpmDeps;};
          setup = ''
            export PLAYWRIGHT_BROWSERS_PATH=${playwright}
            export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true
            export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

            # Configure fonts for headless browser rendering.
            export FONTCONFIG_FILE=${fontsConf}

            # Use pre-built binary from ${appNameDev}.
            mkdir -p ./bin
            cp ${appPackageDev}/bin/${appNameDev} ./bin/${appNameDev}
          '';
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
          # Ignore user Go settings so Nix's pinned toolchain is authoritative.
          export GOENV='off'
          export GOTOOLCHAIN='local'

          # Isolate `go install`ed binaries per checkout, while sharing Go's
          # default, content-addressed module cache across projects.
          export GOBIN="$PWD/bin/.go"
          export PATH="$GOBIN:$PATH"

          export PLAYWRIGHT_BROWSERS_PATH=${playwright}
          export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true

          # Auto-install pnpm packages if needed.
          if [ -f package.json ]; then
            if [ ! -d node_modules ] || \
                [ package.json -nt node_modules ] || \
                [ pnpm-lock.yaml -nt node_modules ]; then
              echo "Installing pnpm packages..."
              CI=true pnpm install --frozen-lockfile
              touch node_modules
            fi
          fi

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
