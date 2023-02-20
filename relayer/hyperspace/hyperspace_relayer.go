// Package rly provides an interface to the cosmos relayer running in a Docker container.
package hyperspace

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/pelletier/go-toml/v2"
	"go.uber.org/zap"
)

var _ ibc.Relayer = &HyperspaceRelayer{}
// ******* DockerRelayer methods that will panic in hyperspace commander, no overrides yet *******
// FlushAcknowledgements() - no hyperspace implementation yet
// FlushPackets() - no hypersapce implementation yet
// UpdatePath() - hyperspace doesn't understand paths, may not be needed.
// LinkPath() - doesn't make sense for hyperspace, interchaintest.Build() will fail because of parachain setup time (if used with parachain)
// UpdateClients() - no hyperspace implementation yet
// AddKey() - no hyperspace implementation yet


// HyperspaceRelayer is the ibc.Relayer implementation for github.com/ComposableFi/hyperspace.
type HyperspaceRelayer struct {
	// Embedded DockerRelayer so commands just work.
	*relayer.DockerRelayer
	hc *hyperspaceCommander
}

func NewHyperspaceRelayer(log *zap.Logger, testName string, cli *client.Client, networkID string, options ...relayer.RelayerOption) *HyperspaceRelayer {
	c := hyperspaceCommander{log: log}
	for _, opt := range options {
		switch o := opt.(type) {
		case relayer.RelayerOptionExtraStartFlags:
			c.extraStartFlags = o.Flags
		}
	}
	dr, err := relayer.NewDockerRelayer(context.TODO(), log, testName, cli, networkID, &c, options...)
	if err != nil {
		panic(err) // TODO: return
	}

	coreConfig := HyperspaceRelayerCoreConfig{
		PrometheusEndpoint: "",
	}
	bytes, err := toml.Marshal(coreConfig)
	if err != nil {
		panic(err) // TODO: return
	}
	err = dr.WriteFileToHomeDir(context.TODO(), "core.config", bytes)
	if err != nil {
		panic(err) // TODO: return
	}

	r := &HyperspaceRelayer{
		DockerRelayer: dr,
		hc: &c,
	}

	return r
}

// HyperspaceCapabilities returns the set of capabilities of the Cosmos relayer.
//
// Note, this API may change if the rly package eventually needs
// to distinguish between multiple rly versions.
func HyperspaceCapabilities() map[relayer.Capability]bool {
	// RC1 matches the full set of capabilities as of writing.
	return nil // relayer.FullCapabilities()
}

func (r *HyperspaceRelayer) RestoreKey(ctx context.Context, rep ibc.RelayerExecReporter, cfg ibc.ChainConfig, keyName, mnemonic string) error {
	addrBytes := ""
	chainID := cfg.ChainID
	coinType := cfg.CoinType
	chainType := cfg.Type

	chainConfig := make(testutil.Toml)
	switch chainType {
	case "cosmos":
		//chainConfig["private_key"] = mnemonic
		bech32Prefix := cfg.Bech32Prefix
		keyEntry := GenKeyEntry(bech32Prefix, coinType, mnemonic)
		keyEntryOverrides := make(testutil.Toml)
		keyEntryOverrides["account"] = keyEntry.Account
		keyEntryOverrides["private_key"] = keyEntry.PrivateKey
		keyEntryOverrides["public_key"] = keyEntry.PublicKey
		keyEntryOverrides["address"] = keyEntry.Address
		chainConfig["keybase"] = keyEntryOverrides
	case "polkadot":
		//chainConfig["private_key"] = "//Alice"
		chainConfig["private_key"] = mnemonic
	}
	chainConfigFile := chainID + ".config"
	err := r.DockerRelayer.ModifyTomlConfigFile(ctx, chainConfigFile, chainConfig)
	if err != nil {
		return err
	}

	r.AddWallet(chainID, NewWallet(chainID, addrBytes, mnemonic))

	return nil
}

func (r *HyperspaceRelayer) SetClientContractHash(ctx context.Context, rep ibc.RelayerExecReporter, cfg ibc.ChainConfig, hash string) error {
	chainConfig := make(testutil.Toml)
	chainConfig["wasm_code_id"] = hash
	chainConfigFile := cfg.ChainID + ".config"
	err := r.ModifyTomlConfigFile(ctx, chainConfigFile, chainConfig)
	if err != nil {
		return err
	}

	return nil
}

func (r *HyperspaceRelayer) PrintCoreConfig(ctx context.Context, rep ibc.RelayerExecReporter) error {
	cmd := []string{
		"cat",
		path.Join(r.HomeDir(), "core.config"),
	}
		
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	res := r.Exec(ctx, rep, cmd, nil)
	if res.Err != nil {
		return res.Err
	}
	fmt.Println(string(res.Stdout))
	return nil
}

func (r *HyperspaceRelayer) PrintConfigs(ctx context.Context, rep ibc.RelayerExecReporter, chainID string) error {
	cmd := []string{
		"cat",
		path.Join(r.HomeDir(), chainID+".config"),
	}
		
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	res := r.Exec(ctx, rep, cmd, nil)
	if res.Err != nil {
		return res.Err
	}
	fmt.Println(string(res.Stdout))
	return nil
}