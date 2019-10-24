package smtp

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/pkg/errors"
)

type verify struct {
	address string
}

type VerifyInterface interface {
	Check() error
}

func Verify(address string) VerifyInterface {
	return &verify{
		address: address,
	}
}

func (v *verify) Check() error {
	check, err := v.email(v.address)
	if err != nil {
		if strings.Contains(err.Error(), `no such host`) {
			err = errors.New(`Server not found`)
		} else {
			err = errors.New(`Server not answer with timeout`)
		}
	} else if !check {
		err = errors.New(`Address not found into server`)
	}

	return err
}

func (v *verify) email(address string) (check bool, err error) {
	sliceAddress := strings.Split(address, `@`)

	ips, err := net.LookupMX(sliceAddress[1])
	if err != nil {
		return false, errors.WithStack(err)
	}

	for _, ip := range ips {
		status, err :=  v.item(address, ip)
		if err != nil {
			return false, errors.WithStack(err)
		}

		if status[:3] == `250` {
			check = true
			break
		}
	}

	return check, nil
}

func (v *verify) item(address string, ip *net.MX) (status string, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
		if err != nil {
			status = ``
			err = errors.WithStack(err)
		}
	}()

	conn, _ := net.Dial("tcp", ip.Host[:len(ip.Host)-1]+":25")
	err = conn.SetDeadline(time.Now().Add(time.Second * 3))
	if err != nil {
		return ``, err
	}

	_, _, err = bufio.NewReader(conn).ReadLine()
	if err != nil {
		return ``, err
	}

	fmt.Fprintf(conn, "helo hi\n")
	_, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return ``, err
	}

	gofakeit.Seed(time.Now().UnixNano())
	fmt.Fprintf(conn, "mail from: <%s>\n", gofakeit.Email())
	_, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return ``, err
	}

	fmt.Fprintf(conn, "rcpt to: <%s>\n", address)
	status, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return ``, err
	}

	fmt.Fprintf(conn, "quit\n")
	_, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return ``, err
	}

	conn.Close()
	return status, nil
}
