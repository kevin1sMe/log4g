package log4g

//by kevin at 2014-07-30

import (
    "os"
    "fmt"
    "io"
    "runtime"
    "sync"
    "time"
)

type logChan chan []byte

var (
    stdFlags  = lstdFlags | lshortfile | lmicroseconds
    loggerStd = &logger{out: os.Stderr, flag: stdFlags}
    logDir string
    recvOver  = make(chan bool)
    logChans = make([]logChan, nFatal + 1)
    logWriters   = make([]*os.File, nFatal + 1)
    subLog   = make([]string, nFatal + 1)
)

var (
    Debugf logfType
    Debug  logType
    Infof  logfType
    Info   logType
    Errorf logfType
    Error  logType
    Fatalf logfType
    Fatal  logType
)

/*const (*/
    /*infoStr  = "[Info ]- "*/
    /*debugStr = "[Debug]- "*/
    /*errorStr = "[Error]- "*/
    /*fatalStr = "[Fatal]- "*/
/*)*/

const (
    nDebug = iota
    nInfo
    nError
    nFatal
)

const (
    LDebug = 1 << iota
    LInfo
    LError
    LFatal
)
const (
    _debug = 1<<(iota+1) - 1
    _info
    _error
    _fatal
)
const (
    ldate = 1 << iota
    ltime
    lmicroseconds
    llongfile
    lshortfile
    lstdFlags = ldate | ltime
)

type logfType func(format string, v ...interface{})
type logType func(v ...interface{})

func Close() {
    recvOver <- true

    for _, c := range logChans {
        close(c)
    }

    for _, o := range logWriters {
        o.Close()
    }
}

func InitLogger(lvl int, dir string) {
    if dir != "" {
        logDir = dir
        //创建日志目录
        createLogDir(logDir)

        //初始化各日志文件
        for i := nDebug ; i <= nFatal ; i++ {
            subLog[i] = ""
        }

        //初始化chan
        initLogChans()

        //循环读取各日志的channel 并写入到文件中
        go func() {
            for {
                select {
                    case v, ok := <-logChans[nDebug]:
                        if ok  {
                            if logWriters[nDebug] != nil {
                                logWriters[nDebug].Write(v) 
                            }else {
                                fmt.Printf("debug logwriter is nil")
                            }
                        }
                    case v, ok := <-logChans[nInfo]:
                        if ok {
                            if logWriters[nInfo] != nil {
                                logWriters[nInfo].Write(v)
                            }else {
                                fmt.Printf("info logwriter is nil")
                            }
                        }
                    case v, ok := <-logChans[nError]:
                        if ok {
                            if logWriters[nError] != nil {
                                /*fmt.Printf("error write bytes: %s\n", string(v))*/
                                logWriters[nError].Write(v)
                            }else {
                                fmt.Printf("error logwriter is nil")
                            }
                        }
                    case v, ok := <-logChans[nFatal]: 
                        if ok {
                            if logWriters[nFatal] != nil {
                                /*fmt.Printf("fatal write bytes: %s\n", string(v))*/
                                logWriters[nFatal].Write(v)
                            }else {
                                fmt.Printf("fatal logwriter is nil")
                            }
                        }
                    case <-recvOver:
                            return

                }
            }
        }()
    }

    var prefix string
    prefix = ""

    if lvl&_debug != 0 {
        Debugf, Debug = makeLog(prefix, nDebug)
    } else {
        Debugf, Debug = emptyLogf, emptyLog
    }
    if lvl&_info != 0 {
        Infof, Info = makeLog(prefix, nInfo)
    } else {
        Infof, Info = emptyLogf, emptyLog
    }
    if lvl&_error != 0 {
        Errorf, Error = makeLog(prefix, nError)
    } else {
        Errorf, Error = emptyLogf, emptyLog
    }
    if lvl&_fatal != 0 {
        Fatalf, Fatal = makeLog(prefix, nFatal)
    } else {
        Fatalf, Fatal = emptyLogf, emptyLog
    }
}

func makeLog(prefix string, lvl int) (x logfType, y logType) {
    return func(format string, v ...interface{}) {
            loggerStd.output(prefix, 2, lvl, fmt.Sprintf(format, v...))
        },
        func(v ...interface{}) {
            loggerStd.output(prefix, 2, lvl, fmt.Sprintln(v...))
        }
}

func emptyLogf(format string, v ...interface{}) {}
func emptyLog(v ...interface{})                 {}

type logger struct {
    mu   sync.Mutex
    flag int
    out  io.Writer
    buf  []byte
}

func itoa(buf *[]byte, i int, wid int) {
    var u uint = uint(i)
    if u == 0 && wid <= 1 {
        *buf = append(*buf, '0')
        return
    }

    var b [32]byte
    bp := len(b)
    for ; u > 0 || wid > 0; u /= 10 {
        bp--
        wid--
        b[bp] = byte(u%10) + '0'
    }
    *buf = append(*buf, b[bp:]...)
}

func (l *logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {
    if l.flag&(ldate|ltime|lmicroseconds) != 0 {
        if l.flag&ldate != 0 {
            year, month, day := t.Date()
            itoa(buf, year, 4)
            *buf = append(*buf, '-')
            itoa(buf, int(month), 2)
            *buf = append(*buf, '-')
            itoa(buf, day, 2)
            *buf = append(*buf, ' ')
        }
        if l.flag&(ltime|lmicroseconds) != 0 {
            hour, min, sec := t.Clock()
            itoa(buf, hour, 2)
            *buf = append(*buf, ':')
            itoa(buf, min, 2)
            *buf = append(*buf, ':')
            itoa(buf, sec, 2)
            if l.flag&lmicroseconds != 0 {
                *buf = append(*buf, '.')
                itoa(buf, t.Nanosecond()/1e3, 6)
            }
            /**buf = append(*buf, ' ')*/
        }

        *buf = append(*buf, '|')
    }
    if l.flag&(lshortfile|llongfile) != 0 {
        if l.flag&lshortfile != 0 {
            short := file
            for i := len(file) - 1; i > 0; i-- {
                if file[i] == '/' {
                    short = file[i+1:]
                    break
                }
            }
            file = short
        }
        *buf = append(*buf, file...)
        *buf = append(*buf, ':')
        itoa(buf, line, -1)
        *buf = append(*buf, "|"...)
    }
}

//lvl 是日志等级。　
func (l *logger) output(prefix string, calldepth int, lvl int, s string) error {
    now := time.Now()
    var file string
    var line int
    l.mu.Lock()
    defer l.mu.Unlock()
    if l.flag&(lshortfile|llongfile) != 0 {
        l.mu.Unlock()
        var ok bool
        _, file, line, ok = runtime.Caller(calldepth)
        if !ok {
            file = "???"
            line = 0
        }
        l.mu.Lock()
    }
    l.buf = l.buf[:0]
    l.buf = append(l.buf, prefix...)

    logfile := getLogFileName(now)
    if subLog[lvl] != logfile {
        //创建此日志文件  
        o, err := os.OpenFile(getLogFullName(lvl, logfile),
                        os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s\n",err.Error())
            os.Exit(-1)
        }

        //更新iowriter
        updateIOWriter(lvl, o)

        subLog[lvl] = logfile
    }

    l.formatHeader(&l.buf, now, file, line)
    l.buf = append(l.buf, s...)
    if len(s) > 0 && s[len(s)-1] != '\n' {
        l.buf = append(l.buf, '\n')
    }
    /*_, err := l.out.Write(l.buf)*/
    newSlice := make([]byte, len(l.buf))

    //获取对应日志级别的chan
    recvChan := logChans[lvl]
    if recvChan != nil {
        copy(newSlice, l.buf)
        /*fmt.Printf("send to recvChan succ: newSlice:%v \n", newSlice)*/
        recvChan <- newSlice
    }else{
        fmt.Fprintf(os.Stderr, "getRecvChan failed: lvl:%d\n", lvl)
    }

    return nil
}

func createLogDir(dir string) {
    subDir := []string{"debug", "info", "error", "fatal"}
    for _, v := range(subDir) {
        //make sub dir
        path := dir + "/" + v
        err := os.MkdirAll(path, 0775)
        if err != nil {
             fmt.Fprintf(os.Stderr, "make dir failed: %s\n", err.Error())
             os.Exit(-1)
         }
    }
}

func getLogFileName(t time.Time) string{
    var buf []byte
    year, month, day := t.Date()
    itoa(&buf, year, 4)
    itoa(&buf, int(month), 2)
    itoa(&buf, day, 2)

    hour, _, _ := t.Clock()
    itoa(&buf, hour, 2)
    return string(buf) + ".log"
}

func getLogFullName(lvl int, logfile string) string {
    subDir := []string{"debug", "info", "error", "fatal"}
    return logDir + "/" + subDir[lvl] + "/" + logfile
}

func updateIOWriter(lvl int, o *os.File) {
    if lvl >= nDebug && lvl <= nFatal {
        if logWriters[lvl] != nil {
            logWriters[lvl].Close()
        }
        logWriters[lvl] = o
    }else {
        fmt.Fprintf(os.Stderr, "lvl err:%d\n", lvl)
    }
}


//初始化chan
func initLogChans() {
    for i:=nDebug; i <= nFatal; i++ {
        logChans[i] = make(chan []byte, 128)
    }
}
