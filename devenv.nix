{
  pkgs,
  lib,
  inputs,
  ...
}:

let
  system = pkgs.stdenv.system;
in
{
  cachix.enable = false;

  languages.go = {
    enable = true;
    package = pkgs.go_1_25;
  };

  packages = [
    inputs.gomod2nix.packages.${system}.default
    pkgs.cruft
    pkgs.golangci-lint
    pkgs.govulncheck
    pkgs.just
    pkgs.nixfmt-rfc-style
  ]
  ++ lib.optionals pkgs.stdenv.isDarwin [
    pkgs.apple-sdk_15
  ];

  env.CGO_ENABLED = "1";

  enterTest = ''
    cruft check
    just check
  '';
}
