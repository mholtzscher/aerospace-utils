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

        version = "0.2.0";

        # macOS-specific build inputs for CoreGraphics CGO bindings
        darwinBuildInputs = pkgs.lib.optionals pkgs.stdenv.isDarwin [
          pkgs.apple-sdk_15
        ];
        # Dev scripts (replaces justfile)
        scripts = {
          dev-build = pkgs.writeShellScriptBin "dev-build" ''
            echo "Building (development)..."
            go build
          '';

          dev-build-release = pkgs.writeShellScriptBin "dev-build-release" ''
            echo "Building (release)..."
            go build -ldflags="-s -w"
          '';

          dev-run = pkgs.writeShellScriptBin "dev-run" ''
            go run . "$@"
          '';

          dev-test = pkgs.writeShellScriptBin "dev-test" ''
            go test ./...
          '';

          dev-test-verbose = pkgs.writeShellScriptBin "dev-test-verbose" ''
            go test -v ./...
          '';

          dev-fmt = pkgs.writeShellScriptBin "dev-fmt" ''
            go fmt ./...
          '';

          dev-vet = pkgs.writeShellScriptBin "dev-vet" ''
            go vet ./...
          '';

          dev-lint = pkgs.writeShellScriptBin "dev-lint" ''
            golangci-lint run
          '';

          dev-check = pkgs.writeShellScriptBin "dev-check" ''
            echo "Formatting..."
            go fmt ./...
            echo "Vetting..."
            go vet ./...
            echo "Linting..."
            golangci-lint run
            echo "Testing..."
            go test ./...
          '';

          dev-tidy = pkgs.writeShellScriptBin "dev-tidy" ''
            go mod tidy
          '';
        };
      in
      {
        packages.default = pkgs.buildGoApplication {
          pname = "aerospace-utils";
          inherit version;
          src = ./.;
          modules = ./gomod2nix.toml;

          buildInputs = darwinBuildInputs;

          # CGO required for CoreGraphics display detection on macOS
          CGO_ENABLED = 1;

          ldflags = [
            "-s"
            "-w"
            "-X github.com/mholtzscher/aerospace-utils/cmd.Version=${version}"
          ];

          meta = with pkgs.lib; {
            description = "CLI for managing Aerospace workspace gaps";
            homepage = "https://github.com/mholtzscher/aerospace-utils";
            license = licenses.mit;
            mainProgram = "aerospace-utils";
            platforms = platforms.darwin ++ platforms.linux;
          };
        };

        formatter = pkgs.nixfmt-rfc-style;

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.gopls
            pkgs.golangci-lint
            pkgs.gotools
            pkgs.gomod2nix
          ]
          ++ darwinBuildInputs
          ++ (builtins.attrValues scripts);

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
