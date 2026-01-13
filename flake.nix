{
  description = "CLI for managing Aerospace workspace gaps";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };

        # macOS-specific build inputs for CoreGraphics CGO bindings
        darwinBuildInputs = pkgs.lib.optionals pkgs.stdenv.isDarwin [
          pkgs.darwin.apple_sdk.frameworks.CoreGraphics
          pkgs.darwin.apple_sdk.frameworks.IOKit
          pkgs.darwin.apple_sdk.frameworks.CoreFoundation
        ];
      in
      {
        packages.default = pkgs.buildGoApplication {
          pname = "aerospace-utils";
          version = "0.2.0";
          src = ./.;
          modules = ./gomod2nix.toml;

          buildInputs = darwinBuildInputs;

          # CGO required for CoreGraphics display detection on macOS
          CGO_ENABLED = 1;

          ldflags = [
            "-s"
            "-w"
            "-X github.com/mholtzscher/aerospace-utils/cmd.Version=0.2.0"
          ];

          meta = with pkgs.lib; {
            description = "CLI for managing Aerospace workspace gaps";
            homepage = "https://github.com/mholtzscher/aerospace-utils";
            license = licenses.mit;
            maintainers = [ ];
            mainProgram = "aerospace-utils";
            platforms = platforms.darwin ++ platforms.linux;
          };
        };

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };

        checks.default = self.packages.${system}.default;

        formatter = pkgs.nixfmt-rfc-style;

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.gopls
            pkgs.golangci-lint
            pkgs.gotools
            pkgs.gomod2nix
          ]
          ++ darwinBuildInputs;

          CGO_ENABLED = "1";
        };

        devShells.ci = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.golangci-lint
          ]
          ++ darwinBuildInputs;

          CGO_ENABLED = "1";
        };
      }
    );
}
