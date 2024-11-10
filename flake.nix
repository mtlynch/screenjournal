{
  description = "Dev environment for ScreenJournal";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";

    # 1.23.2 release
    # We temporarily mismatch the version in Docker (1.23.3) because Nix doesn't
    # have the latest version yet, but we need to fix:
    # https://github.com/golang/go/issues/68976
    go-nixpkgs.url = "github:NixOS/nixpkgs/4ae2e647537bcdbb82265469442713d066675275";

    # 3.44.2 release
    sqlite-nixpkgs.url = "github:NixOS/nixpkgs/5ad9903c16126a7d949101687af0aa589b1d7d3d";

    # 20.6.1 release
    nodejs-nixpkgs.url = "github:NixOS/nixpkgs/78058d810644f5ed276804ce7ea9e82d92bee293";

    # 0.10.0 release
    shellcheck-nixpkgs.url = "github:NixOS/nixpkgs/4ae2e647537bcdbb82265469442713d066675275";

    # 1.2.1 release
    sqlfluff-nixpkgs.url = "github:NixOS/nixpkgs/7cf5ccf1cdb2ba5f08f0ac29fc3d04b0b59a07e4";

    # 1.40.0
    playwright-nixpkgs.url = "github:NixOS/nixpkgs/f5c27c6136db4d76c30e533c20517df6864c46ee";

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
      go = go-nixpkgs.legacyPackages.${system}.go_1_23;
      sqlite = sqlite-nixpkgs.legacyPackages.${system}.sqlite;
      nodejs = nodejs-nixpkgs.legacyPackages.${system}.nodejs_20;
      shellcheck = shellcheck-nixpkgs.legacyPackages.${system}.shellcheck;
      sqlfluff = sqlfluff-nixpkgs.legacyPackages.${system}.sqlfluff;
      playwright = playwright-nixpkgs.legacyPackages.${system}.playwright-driver.browsers;
      flyctl = flyctl-nixpkgs.legacyPackages.${system}.flyctl;
      litestream = litestream-nixpkgs.legacyPackages.${system}.litestream;
    in {
      devShells.default =
        go-nixpkgs.legacyPackages.${system}.mkShell.override
        {
          stdenv = go-nixpkgs.legacyPackages.${system}.pkgsStatic.stdenv;
        }
        {
          packages = [
            go-nixpkgs.legacyPackages.${system}.gotools
            go-nixpkgs.legacyPackages.${system}.gopls
            go-nixpkgs.legacyPackages.${system}.go-outline
            go-nixpkgs.legacyPackages.${system}.gopkgs
            go-nixpkgs.legacyPackages.${system}.gocode-gomod
            go-nixpkgs.legacyPackages.${system}.godef
            go-nixpkgs.legacyPackages.${system}.golint
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
            export GOROOT="${go}/share/go"

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

      formatter = go-nixpkgs.legacyPackages.${system}.alejandra;
    });
}
