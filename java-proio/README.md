# Building
Building the code requires Maven.
```shell
mvn install
```

# Running the "Ls" tool
This is a tool that serves as an example for a browser tool.  This one is simple and only dumps text to the terminal.
```shell
java --illegal-access=deny -cp target/proio-*-jar-with-dependencies.jar proio.Ls ../samples/muons-withmeta.proio | less
```
