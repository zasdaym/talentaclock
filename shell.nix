with (import <nixpkgs> {});

mkShell {
    buildInputs = [
        chromedriver
        git
        go_1_19
     ];
}
