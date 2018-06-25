package run

import (
    "os"
    "wio/cmd/wio/log"
)

func performArgumentCheck(args []string) string {
    var directory string
    var err error

    // check directory
    if len(args) <= 0 {
        directory, err = os.Getwd()
        if err != nil {
            log.WriteErrorlnExit(err)
        }
    } else {
        directory = args[0]
    }

    return directory
}
