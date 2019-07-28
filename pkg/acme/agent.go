package acme

import "github.com/fhofherr/acmeproxy/pkg/acme/internal/challenge"

// AgentConfig contains the configuration for the ACME agent.
type AgentConfig struct {
	DirectoryURL string
}

// Agent takes care of obtaining and renewing ACME certificates for its
// clients.
//
// Agent supports three operation modes:
//
// 1. acme-client: Agent obtains and stores certificates locally. Once it
// successfully obtained or renewed a certificate it executes a
// certificate-obtained hook which delivers the certificate to its consumer.
//
// 2. certificate-agent: just as with the acme-client operation mode Agent
// obtains and renews certificates for its client. However, in contrast to
// acme-client Agent does not execute a certificate-obtained hook. The final
// consumers of the certificate have to actively retrieve the certificate from
// Agent.
//
// TODO(fh): implement certificate-agent mode
//
// 3. acme-gateway: Agent does not store the certificates. Neither does it
// take care of automatically renewing certificates. Instead it allows clients
// to offload obtaining or renewing certificates. Clients have to actively
// request an action from Agent. Once it received a new or renewed certificate
// it directly sends it to its client and forgets all about it afterwards.
//
// TODO(fh): implement acme-gateway mode
type Agent struct {
	acmeClient *Client
}

// NewAgent creates a new instance of Agent.
func NewAgent(cfg AgentConfig) *Agent {
	return &Agent{
		acmeClient: &Client{
			DirectoryURL: cfg.DirectoryURL,
			HTTP01Solver: challenge.NewHTTP01Solver(),
		},
	}
}

// HTTP01ChallengeSolver returns a pointer to the HTTP01Solver used by Agent.
func (a *Agent) HTTP01ChallengeSolver() *challenge.HTTP01Solver {
	return a.acmeClient.HTTP01Solver
}
