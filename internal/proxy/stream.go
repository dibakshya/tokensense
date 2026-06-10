package proxy

import (
	"bytes"
	"io"
	"net"
)

// readAndForward reads from upstream and forwards to client, returning the full response data.
func readAndForward(upstream, client net.Conn) ([]byte, error) {
	var responseBuf bytes.Buffer
	buf := make([]byte, 32*1024)

	for {
		n, err := upstream.Read(buf)
		if n > 0 {
			responseBuf.Write(buf[:n])
			if _, writeErr := client.Write(buf[:n]); writeErr != nil {
				return responseBuf.Bytes(), writeErr
			}
		}
		if err != nil {
			if err == io.EOF {
				return responseBuf.Bytes(), nil
			}
			return responseBuf.Bytes(), err
		}
	}
}
