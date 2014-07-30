log4g
=======

Simple go logger, such as log4j, too simple to use.

fork from github.com/cuixin/log4g

１）根据日志等级，产生不同的日志文件夹
２）日志文件按小时划分，方便查找

installation
------------

    go get github.com/kevin1sMe/log4g

usage
-----

Add main entry to start log4g, specify the output file or not(No
output set the nil argument, just only print to console)


```
    import "github.com/kevin1sMe/log4g"
    import "time"

    func main() {
        log4g.InitLogger(log4g.LDebug, ".")
        defer log4g.Close()
        // begin to output
        log4g.Info("hello world")

        time.Sleep(time.Second * 1)
    }
```

examples&benchmark
-------
    test 1000k lines to print:
    $>go run examples/test.go
    $>tail -f logging.log

    or directly test the speed of writing bytes.
    ```
    $>nohup go run examples/test.go >/dev/null 2>&1
    $>tail -f logging.log
    ```
