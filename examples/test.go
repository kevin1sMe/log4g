package main

import (
    /*"github.com/cuixin/log4g"*/
    "github.com/kevin1sMe/log4g"
    /*"os"*/
    "time"
)

func main() {

    log4g.InitLogger(log4g.LDebug, ".")
    defer log4g.Close()

    logBytes := make([]byte, 26)
    for i := 0; i < 26; i++ {
        logBytes[i] = 'a' + byte(i)
    }
    logString := string(logBytes)
    start := time.Now()
    /*for i := 0; i < 250000; i++ {*/
    for i := 0; i < 10; i++ {
        log4g.Debug(logString)
        log4g.Info(logString)
        log4g.Error(logString)
        log4g.Fatal(logString)
    }

    log4g.Info(time.Since(start))

    time.Sleep(time.Second * 1)
}
