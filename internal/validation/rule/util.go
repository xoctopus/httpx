package rule

import "bytes"

func Slash(data []byte) []byte {
	buf := &bytes.Buffer{}
	buf.WriteRune('/')
	for _, b := range data {
		if b == '/' {
			buf.WriteRune('\\')
		}
		buf.WriteByte(b)
	}
	buf.WriteRune('/')
	return buf.Bytes()
}

func SingleQuote(data []byte) []byte {
	buf := &bytes.Buffer{}
	buf.WriteRune('\'')
	for _, b := range data {
		if b == '\'' {
			buf.WriteRune('\\')
		}
		buf.WriteByte(b)
	}
	buf.WriteRune('\'')
	return buf.Bytes()
}
