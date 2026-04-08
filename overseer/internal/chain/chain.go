package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const auditLogABI = `[{"name":"logViolation","type":"function","inputs":[{"name":"agentId","type":"address"},{"name":"actionHash","type":"bytes32"},{"name":"reason","type":"string"}]}]`

const punishmentABI = `[{"name":"suspend","type":"function","inputs":[{"name":"agentId","type":"address"},{"name":"durationSeconds","type":"uint256"}]},{"name":"shutdown","type":"function","inputs":[{"name":"agentId","type":"address"}]}]`

type Client struct {
	eth        *ethclient.Client
	key        *ecdsa.PrivateKey
	chainID    *big.Int
	auditLog   common.Address
	auditABI   abi.ABI
	punishment common.Address
	punishABI  abi.ABI
}

func New(ctx context.Context) (*Client, error) {
	rpcURL := os.Getenv("ALCHEMY_RPC_URL")
	if rpcURL == "" {
		return nil, fmt.Errorf("ALCHEMY_RPC_URL is not set")
	}

	privKeyHex := os.Getenv("PRIVATE_KEY")
	if privKeyHex == "" {
		return nil, fmt.Errorf("PRIVATE_KEY is not set")
	}
	privKeyHex = strings.TrimPrefix(privKeyHex, "0x")

	auditAddr := os.Getenv("AUDIT_LOG_ADDRESS")
	if auditAddr == "" {
		return nil, fmt.Errorf("AUDIT_LOG_ADDRESS is not set")
	}

	punishAddr := os.Getenv("PUNISHMENT_ADDRESS")
	if punishAddr == "" {
		return nil, fmt.Errorf("PUNISHMENT_ADDRESS is not set")
	}

	eth, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	chainID, err := eth.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	key, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	parsedAudit, err := abi.JSON(strings.NewReader(auditLogABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit log ABI: %w", err)
	}

	parsedPunish, err := abi.JSON(strings.NewReader(punishmentABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse punishment ABI: %w", err)
	}

	return &Client{
		eth:        eth,
		key:        key,
		chainID:    chainID,
		auditLog:   common.HexToAddress(auditAddr),
		auditABI:   parsedAudit,
		punishment: common.HexToAddress(punishAddr),
		punishABI:  parsedPunish,
	}, nil
}

func (c *Client) transactor(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(c.key, c.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	opts.Context = ctx
	return opts, nil
}

func (c *Client) LogViolation(ctx context.Context, agentID string, actionHash [32]byte, reason string) (string, error) {
	data, err := c.auditABI.Pack("logViolation", common.HexToAddress(agentID), actionHash, reason)
	if err != nil {
		return "", fmt.Errorf("failed to pack logViolation: %w", err)
	}

	opts, err := c.transactor(ctx)
	if err != nil {
		return "", err
	}

	contract := bind.NewBoundContract(c.auditLog, c.auditABI, c.eth, c.eth, c.eth)
	tx, err := contract.RawTransact(opts, data)
	if err != nil {
		return "", fmt.Errorf("logViolation tx failed: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Printf("logViolation tx sent: %s", txHash)
	return txHash, nil
}

func (c *Client) Suspend(ctx context.Context, agentID string, durationSeconds uint64) (string, error) {
	data, err := c.punishABI.Pack("suspend", common.HexToAddress(agentID), new(big.Int).SetUint64(durationSeconds))
	if err != nil {
		return "", fmt.Errorf("failed to pack suspend: %w", err)
	}

	opts, err := c.transactor(ctx)
	if err != nil {
		return "", err
	}

	contract := bind.NewBoundContract(c.punishment, c.punishABI, c.eth, c.eth, c.eth)
	tx, err := contract.RawTransact(opts, data)
	if err != nil {
		return "", fmt.Errorf("suspend tx failed: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Printf("suspend tx sent: %s", txHash)
	return txHash, nil
}

func (c *Client) Shutdown(ctx context.Context, agentID string) (string, error) {
	data, err := c.punishABI.Pack("shutdown", common.HexToAddress(agentID))
	if err != nil {
		return "", fmt.Errorf("failed to pack shutdown: %w", err)
	}

	opts, err := c.transactor(ctx)
	if err != nil {
		return "", err
	}

	contract := bind.NewBoundContract(c.punishment, c.punishABI, c.eth, c.eth, c.eth)
	tx, err := contract.RawTransact(opts, data)
	if err != nil {
		return "", fmt.Errorf("shutdown tx failed: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Printf("shutdown tx sent: %s", txHash)
	return txHash, nil
}
