package netio

const (
	AttachTunnel      = 1
	TunnelCreated     = 2
	DisconnectTunnel  = 3
	ListTunnel        = 4
	CreateCertificate = 5
)

type (
	ReqCmd struct {
		Code    int
		Payload interface{}
	}
	PortTuple struct {
		RelayServerPort, ClientPort int
	}
	CreateCertificateReq  struct{}
	CreateCertificateResp struct {
		Bytes []byte
	}
	AttachTunnelReq struct {
		PortTuple PortTuple
		Uid       int
	}
	CreateTunnelResp struct {
		Port int
	}
	DisconnectTunnelReq struct {
		Port int
		Uid  int
	}
	ListTunnelsReq struct {
		Uid int
	}
	ListTunnelResp struct {
		Ports []PortTuple
	}
)
