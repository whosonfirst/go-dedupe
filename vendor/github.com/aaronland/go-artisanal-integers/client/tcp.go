package client

// EXPERIMENTAL

import (
	"bufio"
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type TCPClient struct {
	Client
	url *url.URL
}

func NewTCPClient(u *url.URL) (*TCPClient, error) {

	cl := TCPClient{
		url: u,
	}

	return &cl, nil
}

func (cl *TCPClient) NextInt(ctx context.Context) (int64, error) {

	conn, err := net.Dial("tcp", cl.url.Host)

	if err != nil {
		return -1, err
	}

	str_i, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		return -1, err
	}

	str_i = strings.Trim(str_i, "\n")

	i, err := strconv.ParseInt(str_i, 10, 64)

	if err != nil {
		return -1, err
	}

	return i, err
}
