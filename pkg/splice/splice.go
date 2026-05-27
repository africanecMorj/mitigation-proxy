package pkg

import (
    "io"
    "os"

    "golang.org/x/sys/unix"
)

const spliceSize = 1 << 20

func Splice(src, dst *os.File) error {

    pipefds := make([]int, 2)

    err = unix.Pipe(pipefds)
    if err != nil {
        return err
    }

    defer unix.Close(pipefds[0])
    defer unix.Close(pipefds[1])

    srcFD := int(src.Fd())
    dstFD := int(dst.Fd())

    for {
        
        n, err := unix.Splice(
            srcFD,
            nil,
            pipefds[1],
            nil,
            spliceSize,
            unix.SPLICE_F_MOVE,
        )

        if err != nil {
            return err
        }

        if n == 0 {
            return io.EOF
        }

        remaining := n

        for remaining > 0 {
            written, err := unix.Splice(
                pipefds[0],
                nil,
                dstFD,
                nil,
                remaining,
                unix.SPLICE_F_MOVE,
            )

            if err != nil {
                return err
            }

            remaining -= written
        }
    }
}