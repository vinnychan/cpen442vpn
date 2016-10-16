package logger

import "fmt"

func Log(msg string, server bool) {
    var stamp string
    if server {
        stamp = "[SERVER] "
    } else {
        stamp = "[CLIENT] "
    }

    fmt.Println(stamp + msg)
}