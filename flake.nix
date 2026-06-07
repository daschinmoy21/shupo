{
  description = "shupo — polyglot video processing pipeline dev shell";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    rust-overlay.url = "github:oxalica/rust-overlay";
  };

  outputs = {
    self,
    nixpkgs,
    rust-overlay,
    ...
  }: let
    supportedSystems = ["x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin"];
    forEachSupportedSystem = f:
      nixpkgs.lib.genAttrs supportedSystems (system:
        f {
          pkgs = import nixpkgs {
            inherit system;
            overlays = [(import rust-overlay)];
          };
        });
  in {
    devShells = forEachSupportedSystem ({pkgs}: {
      default = pkgs.mkShell {
        packages = with pkgs; [
          rust-bin.stable.latest.default
          go
          lua5_4

          rust-analyzer
          gopls
          lua-language-server
          nixd
          bash-language-server
          yaml-language-server
          vscode-langservers-extracted

          rustfmt
          clippy
          gotools
          nixfmt-rfc-style
          stylua
          statix
          deadnix
          shellcheck
          golangci-lint
          actionlint

          ffmpeg-full
          docker-compose
          minio-client
          redis
          jq
          just
          git
          pkg-config
          openssl
        ];

        shellHook = ''
          echo "shupo dev shell ready"
          if [ ! -f .env ] && [ -f .env.example ]; then
            cp .env.example .env
            echo "created .env from .env.example"
          fi
        '';
      };
    });
  };
}
