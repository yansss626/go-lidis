package lidis

import (
	"fmt"
	"io"
	"strconv"
)

type ReplyType byte

const (
	StatusReply  ReplyType = '+'
	ErrorReply   ReplyType = '-'
	IntegerReply ReplyType = ':'
	StringReply  ReplyType = '$'
)

type Reply struct {
	Type   ReplyType
	Int    int64
	Bytes  []byte
	IsNull bool
}

func (c *Client) readReply(conn *connection) (*Reply, error) {

	// 按行读取
	line, err := conn.Reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	// 检查行尾结束符是否为"\r\n"
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return nil, fmt.Errorf("unknown reply format %q", line)
	}

	payload := line[1 : len(line)-2]
	replyType := ReplyType(line[0])

	switch replyType {

	case StatusReply: // '+'
		return &Reply{
			Type:  StatusReply,
			Bytes: payload,
		}, nil

	case StringReply: // '$'
		return c.parseBulkString(string(payload), conn)

	case ErrorReply: // '-'
		return &Reply{
			Type:  ErrorReply,
			Bytes: []byte(fmt.Sprintf("(error) %s", payload)),
		}, nil

	case IntegerReply: // ':'
		val, err := strconv.ParseInt(string(payload), 10, 64)
		if err != nil {
			return nil, err
		}
		return &Reply{
			Type: IntegerReply,
			Int:  val,
		}, nil

	default:
		return nil, fmt.Errorf("unknown reply type: '%c'", replyType)
	}

}

func (c *Client) parseBulkString(payload string, conn *connection) (*Reply, error) {

	// 先判断是否为空值
	valueLength, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return nil, err
	}

	if valueLength == -1 {
		return &Reply{
			Type:   StringReply,
			IsNull: true,
		}, nil
	}

	if valueLength < -1 {
		return nil, fmt.Errorf("(error) invalid value length: %d", valueLength)
	}

	// 精确读取 valuelength 个字节
	buffer := make([]byte, valueLength)
	_, err = io.ReadFull(conn.Reader, buffer)
	if err != nil {
		return nil, err
	}

	// 判断结束字符是否为 "\r\n"
	crlf := make([]byte, 2)
	_, err = io.ReadFull(conn.Reader, crlf)
	if err != nil {
		return nil, err
	}

	if string(crlf) != "\r\n" {
		return nil, fmt.Errorf("expecetd /\r/\n at end of bulk string")
	}

	return &Reply{
		Type:   StringReply,
		IsNull: false,
		Bytes:  buffer,
	}, nil

}
