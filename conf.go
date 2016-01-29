package conf

import (
    "io"
    "strings"
    "path/filepath"
    "strconv"
    "time"
)

var (
    boolMap = map[string]bool{
        "on": true,
        "off": false,
        "0": false,
        "1": true,
        "yes": true,
        "no": false,
    }
)

type Conf struct {
    Handler SyntaxHandler
    confRoot string
    prefix string
    data interface{}
    iterIdx int
    items [][]interface{}
}

func Read(h SyntaxHandler, rd io.Reader) (cf *Conf, err error) {
    cf = &Conf{Handler: h}
    err = cf.Read(rd)
    return
}

func Parse(h SyntaxHandler, b []byte) (cf *Conf, err error) {
    cf = &Conf{Handler: h}
    err = cf.Parse(b)
    return
}

func Load(h SyntaxHandler, filename string) (cf *Conf, err error) {
    cf = &Conf{Handler: h}
    err = cf.Load(filename)
    return
}

func LoadDir(h SyntaxHandler, confRoot string) (cf *Conf, err error) {
    cf = &Conf{Handler: h}
    err = cf.LoadDir(confRoot)
    return
}

func (cf *Conf) Read(rd io.Reader) (err error) {
    cf.data, err = cf.Handler.Read(rd)
    return
}

func (cf *Conf) Parse(b []byte) (err error) {
    cf.data, err = cf.Handler.Parse(b)
    return
}

func (cf *Conf) Load(filename string) (err error) {
    cf.data, err = cf.Handler.Load(filename)
    if err != nil {
        cf.confRoot = filepath.Dir(filename)
    }
    return
}

func (cf *Conf) LoadDir(confRoot string) (err error) {
    cf.data, err = cf.Handler.LoadDir(confRoot)
    cf.confRoot = confRoot
    return
}

func (cf *Conf) Dump() (b []byte, err error) {
    b, err = cf.Handler.Dump(cf.data)
    return
}

func (cf *Conf) Write(wr io.Writer) (n int, err error) {
    n, err = cf.Handler.Write(wr, cf.data)
    return
}

func (cf *Conf) draw() *Conf {
    var ret = new(Conf)
    ret.confRoot = cf.confRoot
    ret.prefix = cf.prefix
    return ret
}

func (cf *Conf) SetConfRoot(confRoot string) {
    cf.confRoot = confRoot
}

func (cf *Conf) SetPrefix(prefix string) {
    cf.prefix = prefix
}

func (cf *Conf) Find(p... string) *Conf {
    var pattern = strings.Join(p, ".")
    pattern = strings.Trim(pattern, ".")
    var k string
    var i int
    var v interface{}
    var ok bool
    var err error
    var curr = cf.data
    var ret = cf.draw()
    for _, k = range strings.Split(pattern, ".") {
        switch curr.(type) {
            case map[string]interface{}:
                if v, ok = curr.(map[string]interface{})[k]; ok {
                    curr = v
                } else {
                    return ret
                }
            case []interface{}:
                if i, err = strconv.Atoi(k); err == nil {
                    if i < len(curr.([]interface{})) {
                        curr = curr.([]interface{})[i]
                    } else {
                        return ret
                    }
                } else {
                    return ret
                }
            default:
                return ret
        }
    }
    ret.data = curr
    return ret
}

func (cf *Conf) String(a...string) string {
    var fallback string
    var ok bool
    var ret string
    if len(a) == 0 {
        fallback = ""
    } else {
        fallback = a[0]
    }
    ret, ok = cf.data.(string)
    if ! ok {
        return fallback
    }
    return ret
}

func (cf *Conf) Int64(a...int64) int64 {
    var fallback int64
    var ret int64
    var datastr string
    var err error
    var ok bool
    if len(a) == 0 {
        fallback = 0
    } else {
        fallback = a[0]
    }
    datastr, ok = cf.data.(string)
    if ! ok {
        return fallback
    }
    ret, err = strconv.ParseInt(datastr, 10, 64)
    if err != nil {
        return fallback
    }
    return ret
}

func (cf *Conf) Int(a...int) int {
    var aa =make([]int64, 0, 1)
    if len(a) > 0 {
        aa = append(aa, int64(a[0]))
    }
    return int(cf.Int64(aa...))
}

func (cf *Conf) Path(a...string) string {
    var ret = cf.String(a...)
    if ret[0] != '/' {
        ret = filepath.Join(cf.prefix, ret)
    }
    return ret
}

func (cf *Conf) ConfPath(a...string) string {
    var ret = cf.String(a...)
    if ret[0] != '/' {
        ret = filepath.Join(cf.confRoot, ret)
    }
    return ret
}

func (cf *Conf) Bool(a...bool) bool {
    var fallback = false
    if len(a) > 0 {
        fallback = a[0]
    }
    var datastr = strings.ToLower(cf.String())
    var ret, ok = boolMap[datastr]
    if ! ok {
        return fallback
    }
    return ret
}

func (cf *Conf) Duration(a...time.Duration) time.Duration {
    var fallback, _ = time.ParseDuration("0s")
    var datastr = cf.String()
    if datastr == "" {
        return fallback
    }
    var ret, err = time.ParseDuration(datastr)
    if err != nil {
        return fallback
    }
    return ret
}

func (cf *Conf) Data() interface{} {
    return cf.data
}

func (cf *Conf) Next() bool {
    if cf.items == nil {
        cf.items = make([][]interface{}, 0, 32)
        cf.iterIdx = 0
        switch cf.data.(type) {
            case []interface{}:
                for i, d := range cf.data.([]interface{}) {
                    cf.items = append(cf.items, []interface{}{interface{}(i), d})
                }
            case map[string]interface{}:
                for k, d := range cf.data.(map[string]interface{}) {
                    cf.items = append(cf.items, []interface{}{interface{}(k), d})
                }
            default:
                return false
        }
        if len(cf.items) == 0 {
            return false
        }
        return true
    }
    cf.iterIdx ++
    if cf.iterIdx >= len(cf.items) {
        cf.iterIdx = -1
        return false
    }
    return true
}

func (cf *Conf) Value() *Conf {
    var ret = cf.draw()
    ret.data = cf.items[cf.iterIdx][1]
    return ret
}

func (cf *Conf) Key() string {
    var d = cf.items[cf.iterIdx][0]
    var ret string
    switch d.(type) {
        case string:
            ret = d.(string)
        case int:
            ret = strconv.FormatInt(int64(d.(int)), 10)
    }
    return ret
}

func (cf *Conf) Idx() int {
    var d = cf.items[cf.iterIdx][0]
    var ret, ok = d.(int)
    if ! ok {
        return 0
    }
    return ret
}

func (cf *Conf) Item() (string, *Conf) {
    return cf.Key(), cf.Value()
}

func (cf *Conf) Array() []*Conf {
    var ret = make([]*Conf, 0, 64)
    var d interface{}
    switch cf.data.(type) {
        case []interface{}:
            for _, d = range cf.data.([]interface{}) {
                var c = cf.draw()
                c.data = d
                ret = append(ret, c)
            }
    }
    return ret
}
