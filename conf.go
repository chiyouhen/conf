package conf
//
import (
    "os"
    "io"
    "io/ioutil"
    "bytes"
    "bufio"
    "strings"
    "regexp"
    "path/filepath"
    "strconv"
    "time"
    "fmt"
)

var (
    commentString = map[string]bool{
        "#": true,
        ";": true,
    }
    validKeyPattern, _ = regexp.Compile("^[A-Za-z][A-Za-z0-9-_]*$")
    boolMap = map[string]bool{
        "on": true,
        "off": false,
        "0": false,
        "1": true,
        "yes": true,
        "no": false,
    }
)

type ConfError struct {
    msg string
    filename string
    lineno int
}

func (e *ConfError) Error() string {
    return fmt.Sprintf("%s in '%s' line %d", e.msg, e.filename, e.lineno)
}

type Conf struct {
    confRoot string
    prefix string
    data interface{}
    iterIdx int
    items [][]interface{}
}

func ReadData(rd io.Reader) (data map[string]interface{}, err error) {
    err = nil
    data = make(map[string]interface{})
    var sc = bufio.NewScanner(rd)
    var l string;
    var s string;
    var g []string;
    var stack = make([]map[string]interface{}, 0, 512)
    var curr = data
    var ok bool
    var lg int
    var lv int
    var k string
    var v interface{}
    var fp *os.File

    var lineno = 0
    var filename = "<io.Reader>"

    fp, ok = rd.(*os.File)
    if ok {
        filename = fp.Name()
    }

    stack = append(stack, curr)
    for sc.Scan() {
        lineno ++
        l = sc.Text()
        l = strings.TrimSpace(l)
        if len(l) == 0 {
            continue
        }

        // comment
        if _, ok = commentString[string(l[0])]; ok {
            continue
        }

        // for group
        if l[0] == '[' && l[len(l) - 1] == ']' {
            s = strings.Trim(l, "[]")
            g = strings.Split(s, ".")
            curr = data
            lg = len(g)
            for lv, k = range g {
                if k == "" {
                    // we get . here
                    curr = stack[lv + 1]
                } else {
                    // . but with a string key
                    if _, ok = curr[k]; ok {
                        curr = curr[k].(map[string]interface{})
                        stack = stack[0:lv + 1]
                        stack = append(stack, curr)
                    }
                }
            }
            stack = stack[0:lg]
            k = g[lg - 1]
            v = make(map[string]interface{})
            // check tail key
            if k[0] == '@' {
                // array here
                k = string(k[1:])
                if ! validKeyPattern.Match([]byte(k)) {
                    err = &ConfError{"invalid key", filename, lineno}
                    return
                }
                if _, ok = curr[k]; ! ok {
                    curr[k] = make([]interface{}, 0, 0)
                }
                curr[k] = append(curr[k].([]interface{}), v)
            } else {
                // map here
                if ! validKeyPattern.Match([]byte(k)) {
                    err = &ConfError{"invalid key", filename, lineno}
                    return
                }
                if _, ok = curr[k]; ok {
                    err = &ConfError{"duplicate key", filename, lineno}
                    return
                }
                curr[k] = v
            }
            curr = v.(map[string]interface{})
            stack = append(stack, curr)
        } else {
            g = strings.SplitN(l, ":", 2)
            if len(g) != 2 {
                err = &ConfError{"syntax error", filename, lineno}
                return
            }
            k = strings.TrimSpace(g[0])
            v = strings.TrimSpace(g[1])
            if k[0] == '@' {
                k = string(k[1:])
                if ! validKeyPattern.Match([]byte(k)) {
                    err = &ConfError{"invalid key", filename, lineno}
                    return
                }
                if _, ok = curr[k]; ! ok {
                    curr[k] = make([]interface{}, 0, 0)
                }
                curr[k] = append(curr[k].([]interface{}), v)
            } else {
                if ! validKeyPattern.Match([]byte(k)) {
                    err = &ConfError{"invalid key", filename, lineno}
                    return
                }
                curr[k] = v
            }
        }
    }
    return
}

func ParseData(b []byte) (data map[string]interface{}, err error) {
    var buf = bytes.NewBuffer(b)
    return ReadData(buf)
}

func LoadFileData(filename string) (data map[string]interface{}, err error) {
    var rd *os.File
    rd, err = os.OpenFile(filename, os.O_RDONLY, 0644)
    if err != nil {
        return
    }
    defer rd.Close()
    data, err = ReadData(rd)
    return
}

func LoadDirData(dir string) (data map[string]interface{}, err error) {
    err = nil
    data = make(map[string]interface{})
    var dirs = make([]os.FileInfo, 0, 512)
    var files = make([]os.FileInfo, 0, 512)
    var dirContent []os.FileInfo
    var hasGlobal = false
    var f os.FileInfo
    var k string
    var ok bool

    dirContent, err = ioutil.ReadDir(dir)
    if err != nil {
        return
    }
    for _, f = range dirContent {
        if f.Name()[0] == '.' {
            continue
        } else if f.IsDir() {
            dirs = append(dirs, f)
        } else if f.Name() == "_global.conf" {
            hasGlobal = true
        } else if filepath.Ext(f.Name()) == ".conf" {
            files = append(files, f)
        }
    }
    if hasGlobal {
        data, err = LoadFileData(filepath.Join(dir, "_global.conf"))
        if err != nil {
            return
        }
    }
    for _, f = range files {
        k = strings.TrimSuffix(f.Name(), ".conf")
        if _, ok = data[k]; ! ok {
            data[k], err = LoadFileData(filepath.Join(dir, f.Name()))
            if err != nil {
                return
            }
        }
    }
    for _, f = range dirs {
        k = f.Name()
        if _, ok = data[k]; ! ok {
            data[k], err = LoadDirData(filepath.Join(dir, f.Name()))
            if err != nil {
                return
            }
        }
    }
    return
}

func Read(rd io.Reader) (cf *Conf, err error) {
    cf = &Conf{}
    err = cf.Read(rd)
    return
}

func Parse(b []byte) (cf *Conf, err error) {
    cf = &Conf{}
    err = cf.Parse(b)
    return
}

func LoadFile(filename string) (cf *Conf, err error) {
    cf = &Conf{}
    err = cf.LoadFile(filename)
    return
}

func LoadDir(confRoot string) (cf *Conf, err error) {
    cf = &Conf{}
    err = cf.LoadDir(confRoot)
    return
}

func (cf *Conf) Read(rd io.Reader) (err error) {
    cf.data, err = ReadData(rd)
    return
}

func (cf *Conf) Parse(b []byte) (err error) {
    cf.data, err = ParseData(b)
    return
}

func (cf *Conf) LoadFile(filename string) (err error) {
    cf.data, err = LoadFileData(filename)
    if err != nil {
        cf.confRoot = filepath.Dir(filename)
    }
    return
}

func (cf *Conf) LoadDir(confRoot string) (err error) {
    cf.data, err = LoadDirData(confRoot)
    cf.confRoot = confRoot
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
