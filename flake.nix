{
  description = "Dev environment for ScreenJournal";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    devenv = {
      url = "github:cachix/devenv";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    sqlite-nixpkgs.url = "github:NixOS/nixpkgs/5ad9903c16126a7d949101687af0aa589b1d7d3d";
    shellcheck-nixpkgs.url = "github:NixOS/nixpkgs/4ae2e647537bcdbb82265469442713d066675275";
    sqlfluff-nixpkgs.url = "github:NixOS/nixpkgs/7cf5ccf1cdb2ba5f08f0ac29fc3d04b0b59a07e4";
    playwright-nixpkgs.url = "github:NixOS/nixpkgs/f5c27c6136db4d76c30e533c20517df6864c46ee";
    flyctl-nixpkgs.url = "github:NixOS/nixpkgs/09dc04054ba2ff1f861357d0e7e76d021b273cd7";
    litestream-nixpkgs.url = "github:NixOS/nixpkgs/a343533bccc62400e8a9560423486a3b6c11a23b";
  };

  outputs = {
    self,
    nixpkgs,
    devenv,
    flake-utils,
    sqlite-nixpkgs,
    shellcheck-nixpkgs,
    sqlfluff-nixpkgs,
    playwright-nixpkgs,
    flyctl-nixpkgs,
    litestream-nixpkgs,
  } @ inputs:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
      sqlite = sqlite-nixpkgs.legacyPackages.${system}.sqlite;
      shellcheck = shellcheck-nixpkgs.legacyPackages.${system}.shellcheck;
      sqlfluff = sqlfluff-nixpkgs.legacyPackages.${system}.sqlfluff;
      playwright = playwright-nixpkgs.legacyPackages.${system}.playwright-driver.browsers;
      flyctl = flyctl-nixpkgs.legacyPackages.${system}.flyctl;
      litestream = litestream-nixpkgs.legacyPackages.${system}.litestream;
    in {
      devShells.default = devenv.lib.mkShell {
        inherit inputs pkgs;
        modules = [
          {
            packages = [
              sqlite
              shellcheck
              sqlfluff
              playwright
              flyctl
              litestream
            ];

            languages.go = {
              enable = true;
              package = pkgs.go;
            };

            languages.javascript = {
              enable = true;
              package = pkgs.nodejs-slim;
            };

            enterShell = ''
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
          }
        ];
      };

      formatter = nixpkgs.legacyPackages.${system}.alejandra;
    });
}
