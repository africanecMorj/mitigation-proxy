package pkg

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

const spliceSize = 1 << 20

func setNonblock(fd int) error {
	return unix.SetNonblock(fd, true)
}

func Splice(src, dst *os.File) error {
	srcFD := int(src.Fd())
	dstFD := int(dst.Fd())

	if srcFD == dstFD {
		return errors.New("same fd")
	}

	if err := setNonblock(srcFD); err != nil {
		return err
	}
	if err := setNonblock(dstFD); err != nil {
		return err
	}

	pipefds := make([]int, 2)
	if err := unix.Pipe(pipefds); err != nil {
		return err
	}
	defer unix.Close(pipefds[0])
	defer unix.Close(pipefds[1])

	epfd, err := unix.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer unix.Close(epfd)

	// 🔥 тільки EPOLLIN для src
	if err := unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, srcFD, &unix.EpollEvent{
		Events: unix.EPOLLIN | unix.EPOLLHUP | unix.EPOLLERR,
		Fd:     int32(srcFD),
	}); err != nil {
		return err
	}

	// dst додаємо, але без постійного EPOLLOUT
	if err := unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, dstFD, &unix.EpollEvent{
		Events: unix.EPOLLHUP | unix.EPOLLERR,
		Fd:     int32(dstFD),
	}); err != nil {
		return err
	}

	events := make([]unix.EpollEvent, 4)

	var pipeBuffered int

	for {
		nEvents, err := unix.EpollWait(epfd, events, -1)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return err
		}

		for i := 0; i < nEvents; i++ {
			ev := events[i]

			// 🔥 обробка помилок
			if ev.Events&(unix.EPOLLERR|unix.EPOLLHUP) != 0 {
				return nil
			}

			switch int(ev.Fd) {

			case srcFD:
				for {
					// не читаємо якщо pipe повний
					if pipeBuffered >= spliceSize {
						break
					}

					n, err := unix.Splice(
						srcFD, nil,
						pipefds[1], nil,
						spliceSize-pipeBuffered,
						unix.SPLICE_F_MOVE|unix.SPLICE_F_NONBLOCK,
					)

					if err != nil {
						if errors.Is(err, unix.EAGAIN) {
							break
						}
						if errors.Is(err, unix.EINTR) {
							continue
						}
						return err
					}

					if n == 0 {
						// EOF
						if pipeBuffered == 0 {
							return nil
						}
						break
					}

					pipeBuffered += int(n)
				}

				// 🔥 включаємо EPOLLOUT тільки коли є що писати
				if pipeBuffered > 0 {
					_ = unix.EpollCtl(epfd, unix.EPOLL_CTL_MOD, dstFD, &unix.EpollEvent{
						Events: unix.EPOLLOUT | unix.EPOLLHUP | unix.EPOLLERR,
						Fd:     int32(dstFD),
					})
				}

			case dstFD:
				for pipeBuffered > 0 {
					n, err := unix.Splice(
						pipefds[0], nil,
						dstFD, nil,
						pipeBuffered,
						unix.SPLICE_F_MOVE|unix.SPLICE_F_NONBLOCK,
					)

					if err != nil {
						if errors.Is(err, unix.EAGAIN) {
							break
						}
						if errors.Is(err, unix.EINTR) {
							continue
						}
						return err
					}

					if n == 0 {
						return errors.New("write returned 0")
					}

					pipeBuffered -= int(n)
				}

				// 🔥 якщо все записали — прибираємо EPOLLOUT
				if pipeBuffered == 0 {
					_ = unix.EpollCtl(epfd, unix.EPOLL_CTL_MOD, dstFD, &unix.EpollEvent{
						Events: unix.EPOLLHUP | unix.EPOLLERR,
						Fd:     int32(dstFD),
					})
				}
			}
		}
	}
}