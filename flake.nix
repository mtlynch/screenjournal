{
  description = "Dev environment for ScreenJournal";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";

    # 1.23.0 release
    go-nixpkgs.url = "github:NixOS/nixpkgs/0cb2fd7c59fed0cd82ef858cbcbdb552b9a33465";

    # 20.6.1 release
    nodejs-nixpkgs.url = "github:NixOS/nixpkgs/78058d810644f5ed276804ce7ea9e82d92bee293";

    # 0.9.0 release
    shellcheck-nixpkgs.url = "github:NixOS/nixpkgs/8b5ab8341e33322e5b66fb46ce23d724050f6606";

    # 1.2.1 release
    sqlfluff-nixpkgs.url = "github:NixOS/nixpkgs/7cf5ccf1cdb2ba5f08f0ac29fc3d04b0b59a07e4";

    # 1.40.0
    playwright-nixpkgs.url = "github:NixOS/nixpkgs/f5c27c6136db4d76c30e533c20517df6864c46ee";

    # 0.1.131 release
    flyctl-nixpkgs.url = "github:NixOS/nixpkgs/09dc04054ba2ff1f861357d0e7e76d021b273cd7";

    # 0.3.13 release
    litestream-nixpkgs.url = "github:NixOS/nixpkgs/a343533bccc62400e8a9560423486a3b6c11a23b";
  };

  outputs = { self, flake-utils, go-nixpkgs, nodejs-nixpkgs, shellcheck-nixpkgs, sqlfluff-nixpkgs, playwright-nixpkgs, flyctl-nixpkgs, litestream-nixpkgs }@inputs :
    flake-utils.lib.eachDefaultSystem (system:
    let
      go-nixpkgs = inputs.go-nixpkgs.legacyPackages.${system};
      nodejs-nixpkgs = inputs.nodejs-nixpkgs.legacyPackages.${system};
      shellcheck-nixpkgs = inputs.shellcheck-nixpkgs.legacyPackages.${system};
      sqlfluff-nixpkgs = inputs.sqlfluff-nixpkgs.legacyPackages.${system};
      playwright-nixpkgs = inputs.playwright-nixpkgs.legacyPackages.${system};
      flyctl-nixpkgs = inputs.flyctl-nixpkgs.legacyPackages.${system};
      litestream-nixpkgs = inputs.litestream-nixpkgs.legacyPackages.${system};
    in
    {
      devShells.default = go-nixpkgs.mkShell.override { stdenv = go-nixpkgs.pkgsStatic.stdenv; } {
        packages = [
          go-nixpkgs.gotools
          go-nixpkgs.gopls
          go-nixpkgs.go-outline
          go-nixpkgs.gopkgs
          go-nixpkgs.gocode-gomod
          go-nixpkgs.godef
          go-nixpkgs.golint
          go-nixpkgs.go_1_23
          nodejs-nixpkgs.nodejs_20
          shellcheck-nixpkgs.shellcheck
          sqlfluff-nixpkgs.sqlfluff
          playwright-nixpkgs.playwright-driver.browsers
          flyctl-nixpkgs.flyctl
          litestream-nixpkgs.litestream
        ];

        shellHook = ''
          export GOROOT="${go-nixpkgs.go_1_23}/share/go"

          export PLAYWRIGHT_BROWSERS_PATH=${playwright-nixpkgs.playwright-driver.browsers}
          export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true

          echo "shellcheck" "$(shellcheck --version | grep '^version:')"
          sqlfluff --version
          fly version | cut -d ' ' -f 1-3
          echo "litestream" "$(litestream version)"
          echo "node" "$(node --version)"
          echo "npm" "$(npm --version)"
          go version
        '';
      };
    });
}
