# Go Development
# go
# Go development nix-shell with go compiler and tools
{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    gotools
  ];
}
