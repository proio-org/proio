# Status [![Travis CI Build Status](https://travis-ci.org/decibelcooper/proio.svg?branch=master)](https://travis-ci.org/decibelcooper/proio)
* [Go](go-proio)
  * Mostly complete
  * Still needs to write file descriptor protos to bucket headers
  * Need to add back in examples.
* [Python](py-proio)
  * Mostly complete
  * Still needs to write file descriptor protos to bucket headers
  * Need to add back in examples
* [C++](cpp-proio)
  * Has been cleared with recent rewrite of Go and Python libraries
  * Needs to be rewritten based on new scheme developed in Go and Python
  * High priority
* [Java](java-proio)
  * Has been cleared with recent rewrite of Go and Python libraries
  * Needs to be rewritten based on new scheme developed in Go and Python
  * Lower priority
  
# What is proio?
Proio is a library and set of tools that provide a simple but powerful and performant IO for physics events.  Proio uses Google's protocol buffer libraries and aims simply to add the concept of an event to a protocol buffer IO.  This work was inspired and influenced by ProMC (Sergei Chekanov) and EicMC (Alexander Kiselev).
