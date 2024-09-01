# JR Plugins

Lists of plugin for JR.
Currently the following plugins are included:


- `awsdynamodb`
- `azblobstorage`
- `azcosmosdb`
- `cassandra`
- `elastic`
- `gcs`
- `http`
- `luascript`
- `mongodb`
- `redis`
- `s3`


# Building the plugins

Launch the `make`command with the target `compile`  and the plugins will be built in the `build/` folder


# Creating a plugin

The JR plugins should be in the `internal/plugin` package since they are not meant to be exposed externally.

To build a plugin `someplugin` the following steps are needed:

1. create the package `internal/plugin/someplugin`
2. implement the plugin in a file (e.g. `plugin.go`) with the following requirements:
  - the `plugin.go` file should have conditional build directives:
  ```golang
  //go:build someplugin
  // +build someplugin
  ```
  - a `doc.go` without conditional build directives must be included (with the plugin documentation)
  - the plugin should implement the ´plugin.Plugin´ interface type:
  ```golang
  type Plugin interface {
    jrpc.Producer
    Init(context.Context, []byte) error
}
```
  - in the `plugin.go` file register the plugin:

  ```golang
package someplugin
...
import (
        "github.com/jrnd-io/jr-plugins/internal/plugin"
)
const (
    Name = "someplugin"
)
func init() {
    plugin.RegisterPlugin(Name, &Plugin{})
}
type Plugin struct{
...
}
func (p *Plugin) Init(ctx context.Context, cfgBytes []byte) error{
    ...
}
func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string) (*jrpc.ProduceResponse, error) {
    ...
}

```

  - add the `someplugin` package  to the import in the `run.go` file:
  ```golang
  package main

  import(
    ...
    _ "github.com/jrnd-io/jr-plugins/internal/plugin/someplugin"
  )
  ```
   - add `someplugin`  to the list of plugins in `Makefile``
```make
PLUGINS=mongodb \
        azblobstorage \
        azcosmosdb \
        luascript \
        awsdynamodb \
        s3 \
        cassandra \
        gcs \
        elastic \
        redis \
        http \
        someplugin
```
