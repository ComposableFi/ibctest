package ibc

import (
	"context"

	"github.com/docker/docker/client"
)

type Chain interface {
	// Config fetches the chain configuration.
	Config() ChainConfig

	// Initialize initializes node structs so that things like initializing keys can be done before starting the chain
	Initialize(ctx context.Context, testName string, cli *client.Client, networkID string) error

	// Start sets up everything needed (validators, gentx, fullnodes, peering, additional accounts) for chain to start from genesis.
	Start(testName string, ctx context.Context, additionalGenesisWallets ...WalletAmount) error

	// Exec runs an arbitrary command using Chain's docker environment.
	// Whether the invoked command is run in a one-off container or execing into an already running container
	// is up to the chain implementation.
	//
	// "env" are environment variables in the format "MY_ENV_VAR=value"
	Exec(ctx context.Context, cmd []string, env []string) (stdout, stderr []byte, err error)

	// ExportState exports the chain state at specific height.
	ExportState(ctx context.Context, height int64) (string, error)

	// GetRPCAddress retrieves the rpc address that can be reached by other containers in the docker network.
	GetRPCAddress() string

	// GetGRPCAddress retrieves the grpc address that can be reached by other containers in the docker network.
	GetGRPCAddress() string

	// GetHostRPCAddress returns the rpc address that can be reached by processes on the host machine.
	// Note that this will not return a valid value until after Start returns.
	GetHostRPCAddress() string

	// GetHostGRPCAddress returns the grpc address that can be reached by processes on the host machine.
	// Note that this will not return a valid value until after Start returns.
	GetHostGRPCAddress() string

	// HomeDir is the home directory of a node running in a docker container. Therefore, this maps to
	// the container's filesystem (not the host).
	HomeDir() string

	// CreateKey creates a test key in the "user" node (either the first fullnode or the first validator if no fullnodes).
	CreateKey(ctx context.Context, keyName string) error

	// RecoverKey recovers an existing user from a given mnemonic.
	RecoverKey(ctx context.Context, name, mnemonic string) error

	// GetAddress fetches the bech32 address for a test key on the "user" node (either the first fullnode or the first validator if no fullnodes).
	GetAddress(ctx context.Context, keyName string) ([]byte, error)

	// SendFunds sends funds to a wallet from a user account.
	SendFunds(ctx context.Context, keyName string, amount WalletAmount) error

	// SendIBCTransfer sends an IBC transfer returning a transaction or an error if the transfer failed.
	SendIBCTransfer(ctx context.Context, channelID, keyName string, amount WalletAmount, timeout *IBCTimeout) (Tx, error)

	// Height returns the current block height or an error if unable to get current height.
	Height(ctx context.Context) (uint64, error)

	// GetBalance fetches the current balance for a specific account address and denom.
	GetBalance(ctx context.Context, address string, denom string) (int64, error)

	// GetGasFeesInNativeDenom gets the fees in native denom for an amount of spent gas.
	GetGasFeesInNativeDenom(gasPaid int64) int64

	// Acknowledgements returns all acknowledgements in a block at height.
	Acknowledgements(ctx context.Context, height uint64) ([]PacketAcknowledgement, error)

	// Timeouts returns all timeouts in a block at height.
	Timeouts(ctx context.Context, height uint64) ([]PacketTimeout, error)
}
