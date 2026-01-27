{ pkgs, ... }: {
  channel = "stable-23.11"; 
  packages = [
    pkgs.go         # 必须有这个，否则报 go not found
    pkgs.gnumake    # 必须有这个，否则报 make not found
    pkgs.gcc        # 必须有这个，否则报 gmp.h not found
  ];
  env = {
    CGO_ENABLED = "1";
  };
}