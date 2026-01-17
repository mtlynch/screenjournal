{
  description = "Dev environment for ScreenJournal";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";

    # 1.25.4 release
    go-nixpkgs.url = "github:NixOS/nixpkgs/ee09932cedcef15aaf476f9343d1dea2cb77e261";

    # 3.44.2 release
    sqlite-nixpkgs.url = "github:NixOS/nixpkgs/5ad9903c16126a7d949101687af0aa589b1d7d3d";

    # 20.6.1 release
    nodejs-nixpkgs.url = "github:NixOS/nixpkgs/78058d810644f5ed276804ce7ea9e82d92bee293";

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
    shellcheck-nixpkgs,
    sqlfluff-nixpkgs,
    playwright-nixpkgs,
    flyctl-nixpkgs,
    litestream-nixpkgs,
  } @ inputs:
    flake-utils.lib.eachDefaultSystem (system: let
      gopkg = go-nixpkgs.legacyPackages.${system};
      go = gopkg.go;
      sqlite = sqlite-nixpkgs.legacyPackages.${system}.sqlite;
      nodejs = nodejs-nixpkgs.legacyPackages.${system}.nodejs_20;
      shellcheck = shellcheck-nixpkgs.legacyPackages.${system}.shellcheck;
      sqlfluff = sqlfluff-nixpkgs.legacyPackages.${system}.sqlfluff;
      playwright = playwright-nixpkgs.legacyPackages.${system}.playwright-driver.browsers;
      flyctl = flyctl-nixpkgs.legacyPackages.${system}.flyctl;
      litestream = litestream-nixpkgs.legacyPackages.${system}.litestream;
    in {
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

          export PLAYWRIGHT_BROWSERS_PATH=${playwright}
          export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true

          echo "shellcheck" "$(shellcheck --version | grep '^version:')"
          sqlfluff --version
          fly version | cut -d ' ' -f 1-3
          echo "litestream" "$(litestream version)"
          echo "node" "$(node --version)"
          echo "npm" "$(npm --version)"
          echo "sqlite" "$(sqlite3 --version | cut -d ' ' -f 1-2)"
          go version
        '';
      };

      formatter = gopkg.alejandra;
    });
}
