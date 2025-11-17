{
  description = "XMTP v2 and v3 go node implementation";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    # go 1.20.14 https://lazamar.co.uk/nix-versions/
    old-nixpkgs.url = "github:NixOs/nixpkgs/087f43a1fa052b17efd441001c4581813c34bc19";
  };

  outputs = inputs@{ flake-parts, nixpkgs, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ];
      perSystem = { pkgs, lib, system, ... }:
        let
          pkgs-go120 = import inputs.old-nixpkgs { inherit system; };
        in
        {
          _module.args.pkgs = import nixpkgs {
            inherit system;
            overlays = [
              (_: prev: {
                inherit (pkgs-go120) go_1_20;
              })
            ];
          };
          devShells.default = pkgs.mkShell {
            nativeBuildInputs = [ pkgs.pkg-config ];
            buildInputs = with pkgs; [
              go_1_20
              mockgen
              moreutils
              protoc-gen-go
              gopls
              go-mockery
              golangci-lint
              moreutils
              shellcheck
              protobuf
              protolint

            ] ++ lib.optionals pkgs.stdenv.isDarwin [ pkgs.darwin.cctools ];
          };
        };
    };
}
