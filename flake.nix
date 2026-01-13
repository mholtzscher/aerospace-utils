{
  description = "CLI for managing Aerospace workspace gaps";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        # macOS-specific build inputs for CoreGraphics CGO bindings
        darwinBuildInputs = pkgs.lib.optionals pkgs.stdenv.isDarwin [
          pkgs.darwin.apple_sdk.frameworks.CoreGraphics
          pkgs.darwin.apple_sdk.frameworks.IOKit
          pkgs.darwin.apple_sdk.frameworks.CoreFoundation
        ];
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "aerospace-utils";
          version = "0.2.0";
          src = ./.;

          vendorHash = null;

          # CGO required for CoreGraphics display detection on macOS
          CGO_ENABLED = "1";

          buildInputs = darwinBuildInputs;

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
            platforms = platforms.darwin;
          };
        };

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };

        checks.default = self.packages.${system}.default;

        formatter = pkgs.nixfmt-rfc-style;

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go_1_26
            pkgs.gopls
            pkgs.golangci-lint
            pkgs.gotools
          ]
          ++ darwinBuildInputs;

          CGO_ENABLED = "1";
        };

        devShells.ci = pkgs.mkShell {
          buildInputs = [
            pkgs.go_1_22
            pkgs.golangci-lint
          ]
          ++ darwinBuildInputs;

          CGO_ENABLED = "1";
        };
      }
    );
}
