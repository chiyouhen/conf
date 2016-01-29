# conf
A configure parser library for go.

# Brief
Conf can work with many conf syntax. Developer can write `SyntaxHandler` for their own configure syntax. This library provided a `SyntaxHandler` for Hdf configure format, `SyntaxHandlerHdf`.

# Using of Conf
## Constructors
```
func Read(h SyntaxHandler, rd io.Reader) (cf *Conf, err error)       // read from reader with SyntaxHandler h
func Parse(h SyntaxHandler, b []byte) (cf *Conf, err error)          // parse from []byte with SyntaxHandler h
func Load(h SyntaxHandler, filename string) (cf *Conf, err error)    // load a single file with SyntaxHandler h
func LoadDir(h SyntaxHandler, confRoot string) (cf *Conf, err error) // load whole directory with SyntaxHandler h
```

## Access Configure Node
No Syntax defined in Conf, all things about configure file syntax are defined in `SyntaxHandler`. Data structure must math the operation of Conf.
### Data Structure
1. The root of data **can only** be a `map[string]interface{}`. It is similar to json, but root cannot be an array.
1. A map **cannot** contain an array nor a map directly. Each map or array **must** have a key name.
1. In an array, data type must be the same.
1. Values can only be string, map or array(slice). Key can only be string.

### Access Data Node
#### Access value in a map:
```
{
    "name": "test",
    "location": "china"
}
```
If we want the value of `name`.
```
s := cf.Find("name").String()
```
`s` would be a string `test`.

#### Access value in a deep level of maps:
```
{
    "server": {
        "name": "localhost",
        "listen": ":80"
    },
    "log": "/var/log/hello.log"
}
```
If we want the value of `name`.
```
s := cf.Find("server.name").String()
```
or
```
s := cf.Find("server", "name").String()
```
`s` would be the string `localhost`.

#### Access array:
```
{
    "rewrite": [
        {
            "from": "/",
            "to": "/index.php"
        },
        {
            "from": "/s",
            "to": "/search.php"
        }
    ]
}
```
If we want to access each `rewrite` rule.
```
c := cf.Find("rewrite")
for _, i := range c.Array() {
    fr := i.Find("from").String()
    /* other processing code */
}
```

# Using of `SyntaxHandler`
A `SyntaxHandler` must implement the following interface.
```
type SyntaxHandler interface {
    Read(rd io.Reader) (interface{}, error)
    Load(filename string) (interface{}, error)
    Parse(b []byte) (interface{}, error)
    LoadDir(dirname string) (interface{}, error)
    Write(wr io.Writer, data interface{}) (int, error)
    Dump(data interface{}) ([]byte, error)
}
```
