package chord

import (
	"math/big"
	"net"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

const (
	FindSuccessorMethod   RPCMethod = "FindSuccessor"
	FindPredecessorMethod RPCMethod = "FindPredecessor"
	NotifyMethod          RPCMethod = "Notify"
	LeaveMethod           RPCMethod = "Leave"
	StablizeMethod        RPCMethod = "Stablize"
)

// types used for remote procedure calls on the network
type (
	RPCMethod string

	RPCRequest struct {
		Method RPCMethod
		Data   []byte
	}
)

func NewRPCRequest(method RPCMethod, data []byte) *RPCRequest {
	return &RPCRequest{
		Method: method,
		Data:   data,
	}
}

// Transport is the interface for the network transport.
// It listens for incoming connections and invokes remote procedure
type (
	Transport interface {
		listen(node *Node) error
		invokeRPC(address string, req *RPCRequest) ([]byte, error)
	}

	transportImpl struct{}
)

func NewTransport() Transport {
	return &transportImpl{}
}

func (t *transportImpl) listen(node *Node) error {
	log.Info().Str("node_id", node.ID.String()).Msgf("%s: Listening for connections...", node.Address)
	listener, err := net.Listen("tcp", node.Address)
	if err != nil {
		log.Fatal().
			Err(err).Str("node_id", node.ID.String()).
			Msgf("%s: Failed to listen", node.Address)
		return err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Error().
					Err(err).Str("node_id", node.ID.String()).
					Msgf("%s: Failed to accept connection", node.Address)
				continue
			}

			go t.handleConnection(conn, node)
		}
	}()

	return nil
}

// handleConnection handles an incoming connection and processes the request dynamically.
func (t *transportImpl) handleConnection(conn net.Conn, node *Node) {
	defer conn.Close()

	req := &RPCRequest{}
	if err := jsoniter.NewDecoder(conn).Decode(req); err != nil {
		log.Error().
			Err(err).Str("node_id", node.ID.String()).
			Msgf("%s: Failed to decode request", node.Address)
		return
	}

	var res []byte
	switch req.Method {
	case FindSuccessorMethod:
		var position big.Int
		position.SetBytes(req.Data)
		log.Info().
			Str("successor_pos", position.String()).
			Msgf("%s: Handling FindSuccessor request", node.Address)

		successor, err := node.findSuccessor(&position)
		if err != nil {
			log.Error().
				Err(err).Str("node_id", node.ID.String()).
				Msgf("%s: Failed to find successor", node.Address)
			return
		}

		res, err = jsoniter.Marshal(successor)
		if err != nil {
			log.Error().
				Err(err).Str("node_id", node.ID.String()).
				Msgf("%s: Failed to encode successor", node.Address)
		}
	case FindPredecessorMethod:
		var id big.Int
		id.SetBytes(req.Data)
		log.Info().
			Str("predecessor_pos", id.String()).
			Msgf("%s: Handling FindPredecessor request", node.Address)

		predecessor, err := node.findPredecessor(&id)
		if err != nil {
			log.Error().
				Err(err).Str("node_id", node.ID.String()).
				Msgf("%s: Failed to find predecessor", node.Address)
			return
		}

		res, err = jsoniter.Marshal(predecessor)
		if err != nil {
			log.Error().
				Err(err).Str("node_id", node.ID.String()).
				Msgf("%s: Failed to encode predecessor", node.Address)
			return
		}
	case NotifyMethod:
		pred := &NodeInfo{}
		if err := jsoniter.Unmarshal(req.Data, pred); err != nil {
			log.Error().
				Err(err).Str("node_id", node.ID.String()).
				Msgf("%s: Failed to decode predecessor", node.Address)
			return
		}

		// write the response
		res = []byte("OK")

		log.Info().
			Str("predecessor_id", pred.ID.String()).
			Msgf("%s: Handling Notify request", node.Address)
		node.notify(pred)
	}

	if _, err := conn.Write(res); err != nil {
		log.Error().
			Err(err).Str("node_id", node.ID.String()).
			Msgf("%s: Failed to write response", node.Address)
	}
}

// invokeRPC sends a remote procedure call to the given address and returns the response
func (t *transportImpl) invokeRPC(address string, req *RPCRequest) ([]byte, error) {
	log.Info().Msgf("invoking RPC %s on %s", req.Method, address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to dial %s", address)
		return nil, err
	}
	defer conn.Close()

	// marshal the request and write to the connection
	data, err := jsoniter.Marshal(req)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to marshal request %s", address)
		return nil, err
	}

	_, err = conn.Write(data)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to write request %s", address)
		return nil, err
	}

	// read the response from the connection and return
	res := make([]byte, 1024)
	if _, err = conn.Read(res); err != nil {
		log.Error().Err(err).Msgf("Failed to read response %s", address)
		return nil, err
	}

	return res, nil
}
