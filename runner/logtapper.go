// +build ignore
package runner

import (
	"io"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/dmcgowan/streams/spdy"
)

func TapServer(l net.Listener, lr *LogRouter) {
	for {
		c, err := l.Accept()
		if err != nil {
			if err != io.EOF {
				logrus.Errorf("Listen error: %#v", err)
			}
			return
		}

		p, err := spdy.NewStreamProvider(c, true)
		if err != nil {
			logrus.Errorf("Error creating stream provider: %#v", err)
			continue
		}
		t := spdy.NewTransport(p)
		go func() {
			r, err := t.WaitReceiveChannel()
			if err != nil {
				// TODO(dmcgowan): log this
				return
			}
			// Start receiving messages
			// Parse messages
			// Tap writers
		}()
	}
}
