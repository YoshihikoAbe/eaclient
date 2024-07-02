# eaclient

A client for e-amusement (XRPC) services 

# Usage

This document assumes that you already have a basic understanding of the protocol used by e-amusement services.
Mostly accurate documentation of the protocol can be found elsewhere.

## Configuration

In order to use a service, it must first be defined in a config file.
A simple config file containing a single service named `test` can be written like so:
```
client:
  model: "EAM:J:A:A"
  srcid: "1000"

services:
  test:            
    url: "http://test/"
    obfuscate: true
    compress: "lz77"
    encoding: "UTF-8"
```
See `sample/sample.yml` in this repository for a sample config with more detailed documentation.

## Requests

A simple request file looks like this:
```
<call>
    <MODULE method="METHOD"/>
</call>
```
The `model` and `srcid` attributes of the `call` node will be automatically filled in using their respective config values.

## CLI

```
Usage: eaclient [OPTIONS] CONFIG SERVICE REQUEST
List of available options:
  -m string
        Override the client's model
  -p string
        Override the client's PCBID
  -u string
        Override the value of the client's User-Agent header
```
Once you have written your config file, you may use the services defined in it by running `eaclient CONFIG SERVICE REQUEST`.
