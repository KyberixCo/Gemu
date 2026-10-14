//go:build !windows || !amd64

package overlay

import "fmt"

type Client struct{}

func HasD3D11(uint32) bool { return false }
func Attach(uint32, string, string, string) (*Client, error) {
	return nil, fmt.Errorf("native overlay requires Windows x64")
}
func (*Client) Ready() bool                  { return false }
func (*Client) Update(string, string, bool)  {}
func (*Client) Command() uint32              { return CommandNone }
func (*Client) Dimensions() (uint32, uint32) { return 0, 0 }
func (*Client) Close()                       {}
