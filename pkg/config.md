# README

Config package is here to provide a prioritized configuration registry for all subsystems/packages. Based on `github.com/spf13/viper` it provides:

* A separate package to expose by value all loaded configs to any node package
  * No cyclic dependency issues
  * No config changes after the initial loading by main pkg
* A prioritized registry for different config sources \(flags, config file, env\)
* Support for a minimal config language \(TOML\)
* Support for other config formats like JSON, YAML, HCL
  * _.json,_ .toml, _.yaml,_ .yml, _.properties,_ .props, _.prop,_ .hcl
* Live-watching and re-reading of config files
* Unmarshal-to-structs for better code readability and easier maintenance
* config examples - `samples/default.dusk.toml`
* Multiple config file searchPaths
  * current working directory
  * $HOME/.dusk/
* Remote Key/Value Store \(not yet in-use\)
* List of general constants defined in config/consts.go 

Example usage:

```bash
# Try to load a config file from any of the searchPaths
# and overwrite general.network setting
user$ ./testnet --general.network=testnet

# with shorthand letter
user$ ./testnet -n=testnet

# Load a specified config file and overwrite logger.level
# config file name can be in form of dusk.toml, dusk.json, dusk.yaml, dusk.tf
user$ ./testnet --config=./pkg/config/default.dusk.toml --logger.level=error

# Load config file found in $searchPaths and overwrite general.network value
user$ DUSK_GENERAL_NETWORK=mainnet; ./testnet

# Load config where a file config value is overwritten by both ENV var and CLI flag but CLI flag has higher priority
user$ DUSK_LOGGER_LEVEL=WARN; ./testnet --logger.level=error

# with shorthand letter
user$ ./testnet -l=error
```

More detailed and up-to-date examples about supported flags, and ENV vars can be found in `loader_test.go`

## Viper

```text
Viper is a prioritized configuration registry. It maintains a set of configuration sources, fetches values to populate those, and provides them according to the source's priority.

The priority of the sources is the following:
1. overrides
2. flags
3. env. variables
4. config file
5. key/value store
6. defaults
```

