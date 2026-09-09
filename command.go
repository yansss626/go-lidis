package lidis

import (
	"fmt"
	"strings"
)

const (
	HashSet = "HSET"
	HashGet = "HGET"
	HashMod = "HMOD"
	HashDel = "HDEL"

	RbtreeSet = "RSET"
	RbtreeGet = "RGET"
	RbtreeMod = "RMOD"
	RbtreeDel = "RDEL"
)

func (c *Client) WriteRequest(args []string, conn *connection) error {

	command, err := c.BuildCommand(args)
	if err != nil {
		return err
	}
	if command == nil {
		return fmt.Errorf("failed to build lidis command")
	}

	totalWriten := 0
	length := len(command)

	for totalWriten < length {
		n, err := conn.Conn.Write(command[totalWriten:])
		if err != nil {
			return err
		}

		totalWriten += n
	}

	return nil
}

func (c *Client) BuildCommand(args []string) ([]byte, error) {

	count := len(args)
	if count == 0 {
		return nil, fmt.Errorf("lack of argument")
	}

	// build body
	var body strings.Builder

	for i := 0; i < count; i++ {
		argLen := len(args[i])
		body.WriteString(fmt.Sprintf("^%d&%s", argLen, args[i]))

		if i == count-1 {
			body.WriteString("\r\n")
		}
	}

	bodyLen := body.Len()

	// build head
	var command strings.Builder
	command.WriteString(fmt.Sprintf("#%d\r\n", bodyLen))

	// build command
	command.WriteString(body.String())

	return []byte(command.String()), nil
}

func (c *Client) HashGet(key string) (string, error) {
	conn, release, err := c.acquireConnection()
	if err != nil {
		return "", err
	}
	defer release()

	err = c.WriteRequest([]string{HashGet, key}, conn)
	if err != nil {
		return "", err
	}

	Reply, err := c.readReply(conn)
	if err != nil {
		return "", err
	}

	if Reply.Type != StringReply {
		if Reply.Type == ErrorReply {
			return "", fmt.Errorf("%s", Reply.Bytes)
		}
		return "", fmt.Errorf("reply type mismatch: expected string, got %c", Reply.Type)
	}

	if Reply.IsNull {
		return "", nil

	}

	return string(Reply.Bytes), nil
}

func (c *Client) HashSet(key, value string) (bool, error) {

	conn, release, err := c.acquireConnection()
	if err != nil {
		return false, err
	}
	defer release()

	err = c.WriteRequest([]string{HashSet, key, value}, conn)
	if err != nil {
		return false, err
	}

	Reply, err := c.readReply(conn)
	if err != nil {
		return false, err
	}

	if Reply.Type != StatusReply {
		if Reply.Type == ErrorReply {
			return false, fmt.Errorf("%s", Reply.Bytes)
		}

		if !Reply.IsNull {
			return false, nil
		}
		return false, fmt.Errorf("reply type mismatch: expected string, got %c", Reply.Type)
	}

	return true, nil

}

func (c *Client) HashMod(key, value string) (bool, error) {
	conn, release, err := c.acquireConnection()
	if err != nil {
		return false, err
	}
	defer release()

	err = c.WriteRequest([]string{HashMod, key, value}, conn)
	if err != nil {
		return false, err
	}

	Reply, err := c.readReply(conn)
	if err != nil {
		return false, err
	}

	if Reply.Type != StatusReply {
		if Reply.Type == ErrorReply {
			return false, fmt.Errorf("%s", Reply.Bytes)
		}

		if Reply.Type == IntegerReply {
			return false, nil
		}
		return false, fmt.Errorf("reply type mismatch: expected string, got %c", Reply.Type)
	}

	return true, nil
}

func (c *Client) HashDel(key, value string) (bool, error) {
	conn, release, err := c.acquireConnection()
	if err != nil {
		return false, err
	}
	defer release()

	err = c.WriteRequest([]string{HashDel, key, value}, conn)
	if err != nil {
		return false, err
	}

	Reply, err := c.readReply(conn)
	if err != nil {
		return false, err
	}

	if Reply.Type != StatusReply {
		if Reply.Type == ErrorReply {
			return false, fmt.Errorf("%s", Reply.Bytes)
		}

		if !Reply.IsNull {
			return false, nil
		}
		return false, fmt.Errorf("reply type mismatch: expected string, got %c", Reply.Type)
	}

	return true, nil

}
