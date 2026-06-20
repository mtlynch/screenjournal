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

    # 3.3.0 release
    sqlfluff-nixpkgs.url = "github:NixOS/nixpkgs/bf689c40d035239a489de5997a4da5352434632e";

    # 1.57.0
    playwright-nixpkgs.url = "github:NixOS/nixpkgs/5f02c91314c8ba4afe83b256b023756412218535";

    # 0.1.131 release
    flyctl-nixpkgs.url = "github:NixOS/nixpkgs/09dc04054ba2ff1f861357d0e7e76d021b273cd7";

    # 0.3.13 release
    litestream-nixpkgs.url = "github:NixOS/nixpkgs/a343533bccc62400e8a9560423486a3b6c11a23b";
  };

  outputs = {
    self,
    flake-utils,
    go-nixpkgs,
    sqlite-nixpkgs,
    nodejs-nixpkgs,
    pnpm-nixpkgs,
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
      pnpm = pnpm-nixpkgs.legacyPackages.${system}.pnpm_10.override {inherit nodejs;};
      shellcheck = shellcheck-nixpkgs.legacyPackages.${system}.shellcheck;
      sqlfluff = sqlfluff-nixpkgs.legacyPackages.${system}.sqlfluff;
      playwright = playwright-nixpkgs.legacyPackages.${system}.playwright-driver.browsers;
      flyctl = flyctl-nixpkgs.legacyPackages.${system}.flyctl;
      litestream = litestream-nixpkgs.legacyPackages.${system}.litestream;

      # Fonts for Playwright browser tests.
      fontsConf = nodepkgs.makeFontsConf {
        fontDirectories = [nodepkgs.dejavu_fonts];
      };

      goVendorHash = "sha256-J7KOCiad1xcAbKU6nz4HOYn1xcnjuCpocXFIhkeN23w=";

      pnpmDepsHash = "sha256-4KVX/YzoLYxu3Cr7hYAaL8LovuEvWyzT7srHhLIpfbU=";

      appName = "screenjournal";
      appNameDev = "${appName}-dev";

      source = gopkg.lib.cleanSourceWith {
        src = ./.;
        filter = path: type:
          ! builtins.elem (builtins.baseNameOf path) [
            ".direnv"
            ".pnpm-store"
            "e2e-results"
            "node_modules"
            "playwright-report"
            "reference"
            "result"
          ];
      };

      pnpmDeps = pnpm.fetchDeps {
        pname = "${appName}-pnpm-deps";
        version = "0.0.0";
        src = source;
        fetcherVersion = 2;
        hash = pnpmDepsHash;
      };

      appPackageDev = buildGoModule {
        pname = appNameDev;
        version = "0.0.1";
        src = source;
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
            src = source;
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
        e2e-tests = mkBuildStep {
          name = "e2e-tests";
          command = "pnpm exec playwright test";
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
