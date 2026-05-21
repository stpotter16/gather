{
    description = "Dev environment for Gather";

    inputs = {
      flake-utils.url = "github:numtide/flake-utils";

      # 1.25.2 release
      go-nixpkgs.url = "github:NixOS/nixpkgs/01b6809f7f9d1183a2b3e081f0a1e6f8f415cb09";

      # 0.11.0 release
      shellcheck-nixpkgs.url = "github:NixOS/nixpkgs/ee09932cedcef15aaf476f9343d1dea2cb77e261";

      # 3.5.0 release
      sqlfluff-nixpkgs.url = "github:NixOS/nixpkgs/ee09932cedcef15aaf476f9343d1dea2cb77e261";

      # 24.5.0
      nodejs-nixpkgs.url = "github:NixOS/nixpkgs/281aac132f6cd84252a5a242cde14c183f600cbc";

      # 0.4.36 release
      flyctl-nixpkgs.url = "github:NixOS/nixpkgs/01fbdeef22b76df85ea168fbfe1bfd9e63681b30";
    };

    outputs = {
      self,
      flake-utils,
      go-nixpkgs,
      shellcheck-nixpkgs,
      sqlfluff-nixpkgs,
      nodejs-nixpkgs,
      flyctl-nixpkgs
    } @inputs:
      flake-utils.lib.eachDefaultSystem (system: let
        gopkg = go-nixpkgs.legacyPackages.${system};
        go = gopkg.go_1_25;
        postgresql = gopkg.postgresql;
        shellcheck = shellcheck-nixpkgs.legacyPackages.${system}.shellcheck;
        sqlfluff = sqlfluff-nixpkgs.legacyPackages.${system}.sqlfluff;
        nodejs = nodejs-nixpkgs.legacyPackages.${system}.nodejs_24;
        flyctl = flyctl-nixpkgs.legacyPackages.${system}.flyctl;
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
              postgresql
              shellcheck
              sqlfluff
              nodejs
              flyctl
            ];

            shellHook = ''
              PROJECT_NAME="$(basename "$PWD")"
              export GOPATH="$HOME/.local/share/go-workspaces/$PROJECT_NAME"
              export GOROOT="${go}/share/go"

              go version
              echo "node" "$(node --version)"
              echo "npm" "$(npm --version)"
              fly version | cut -d ' ' -f 1-3
              echo "psql" "$(psql --version)"
              echo "shellcheck" "$(shellcheck --version | grep '^version:')"
              sqlfluff --version
            '';
        };
      });
}
