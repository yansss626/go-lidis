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

func (c *Connection) writeRequest(args []string) error {

	command, err := c.buildCommand(args)
	if err != nil {
		return err
	}
	if command == nil {
		return fmt.Errorf("failed to build lidis command")
	}

	totalWriten := 0
	length := len(command)

	for totalWriten < length {
		n, err := c.conn.Write(command[totalWriten:])
		if err != nil {
			return err
		}

		totalWriten += n
	}

	return nil
}

func (c *Connection) buildCommand(args []string) ([]byte, error) {

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

func (c *Connection) HashGet(key string) (string, error) {

	err := c.writeRequest([]string{HashGet, key})
	if err != nil {
		c.markError(err)
		return "", err
	}

	Reply, err := c.readReply()
	if err != nil {
		c.markError(err)
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

func (c *Connection) HashSet(key, value string) (bool, error) {

	err := c.writeRequest([]string{HashSet, key, value})
	if err != nil {
		c.markError(err)
		return false, err
	}

	Reply, err := c.readReply()
	if err != nil {
		c.markError(err)
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

func (c *Connection) HashMod(key, value string) (bool, error) {

	err := c.writeRequest([]string{HashMod, key, value})
	if err != nil {
		c.markError(err)
		return false, err
	}

	Reply, err := c.readReply()
	if err != nil {
		c.markError(err)
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

func (c *Connection) HashDel(key string) (bool, error) {

	err := c.writeRequest([]string{HashDel, key})
	if err != nil {
		c.markError(err)
		return false, err
	}

	Reply, err := c.readReply()
	if err != nil {
		c.markError(err)
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
