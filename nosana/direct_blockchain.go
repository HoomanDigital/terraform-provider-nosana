package nosana

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
)

// DirectJobSubmission represents a direct blockchain job submission (like CLI does)
type DirectJobSubmission struct {
	PrivateKey    solana.PrivateKey
	PublicKey     solana.PublicKey
	RPCClient     *rpc.Client
	MarketAddress string
	PriorityFee   uint64 // in microlamports
}

// JobResult represents the result of a direct job submission
type JobResult struct {
	JobID     string `json:"job"`
	RunID     string `json:"run"`
	TxHash    string `json:"tx"`
	IPFSHash  string `json:"ipfs_hash"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// NewDirectJobSubmission creates a new direct blockchain job submission client
func NewDirectJobSubmission(privateKeyBase58, rpcURL, marketAddress string, priorityFee uint64) (*DirectJobSubmission, error) {
	privateKeyBytes, err := base58.Decode(privateKeyBase58)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length: expected %d bytes, got %d", ed25519.PrivateKeySize, len(privateKeyBytes))
	}

	privateKey := solana.PrivateKey(privateKeyBytes)
	publicKey := privateKey.PublicKey()

	var rpcClient *rpc.Client
	if rpcURL != "" {
		rpcClient = rpc.New(rpcURL)
	} else {
		rpcClient = rpc.New("https://api.mainnet-beta.solana.com") // fallback
	}

	return &DirectJobSubmission{
		PrivateKey:    privateKey,
		PublicKey:     publicKey,
		RPCClient:     rpcClient,
		MarketAddress: marketAddress,
		PriorityFee:   priorityFee,
	}, nil
}

// SubmitJobDirectly submits a job directly to the Nosana blockchain via Solana RPC
// This replicates what nosana.jobs.list() does: direct blockchain transaction to Nosana Jobs program
func (d *DirectJobSubmission) SubmitJobDirectly(ctx context.Context, ipfsHash string, timeout int) (*JobResult, error) {
	log.Printf("[INFO] Submitting job DIRECTLY to Nosana Jobs program on Solana blockchain")
	log.Printf("[INFO] IPFS hash: %s", ipfsHash)
	log.Printf("[INFO] Priority fee: %d microlamports", d.PriorityFee)
	log.Printf("[INFO] Market: %s", d.MarketAddress)

	// STEP 1: Generate job and run keypairs (like CLI does)
	jobKeypair := solana.NewWallet()
	runKeypair := solana.NewWallet()
	
	log.Printf("[INFO] Generated Job ID: %s", jobKeypair.PublicKey().String())
	log.Printf("[INFO] Generated Run ID: %s", runKeypair.PublicKey().String())

	// STEP 2: Build Nosana Jobs program transaction
	// Based on nosana-sdk: this.jobs!.methods.list([...bs58.decode(ipfsHash).subarray(2)], new BN(jobTimeout))
	txResult, err := d.buildAndSubmitJobTransaction(ctx, ipfsHash, timeout, jobKeypair, runKeypair)
	if err != nil {
		return nil, fmt.Errorf("failed to submit job transaction: %w", err)
	}

	// STEP 3: Return real job result (same format as CLI)
	result := &JobResult{
		JobID:     jobKeypair.PublicKey().String(),
		RunID:     runKeypair.PublicKey().String(), 
		TxHash:    txResult,
		IPFSHash:  ipfsHash,
		Status:    "QUEUED", // Same as CLI
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	log.Printf("[INFO] ✅ Job submitted to Nosana blockchain! Job ID: %s, TX: %s", result.JobID, result.TxHash)
	return result, nil
}

// buildAndSubmitJobTransaction builds and submits the REAL Nosana Jobs program transaction
// This replicates the exact Anchor transaction that nosana.jobs.list() creates
func (d *DirectJobSubmission) buildAndSubmitJobTransaction(ctx context.Context, ipfsHash string, timeout int, jobKeypair, runKeypair *solana.Wallet) (string, error) {
	log.Printf("[INFO] 🔧 Building REAL Nosana Jobs program transaction...")

	// Nosana program addresses (mainnet)
	nosanaJobsProgram := solana.MustPublicKeyFromBase58("nosJhNRqr2bc9g1nfGDcXXTXvYUmxD4cVwy2pMWhrYM")
	nosTokenMint := solana.MustPublicKeyFromBase58("nosXBVoaCTtYdLvKY6Csb4AC8JCdQKKAaWYtx2ZMoo7")
	marketPublicKey := solana.MustPublicKeyFromBase58(d.MarketAddress)
	
	log.Printf("[INFO] 📋 Jobs Program: %s", nosanaJobsProgram.String())
	log.Printf("[INFO] 🪙 NOS Token: %s", nosTokenMint.String())
	log.Printf("[INFO] 🏪 Market: %s", marketPublicKey.String())

	// Get user's associated token account for NOS token
	userTokenAccount, _, err := solana.FindAssociatedTokenAddress(d.PublicKey, nosTokenMint)
	if err != nil {
		return "", fmt.Errorf("failed to find user token account: %w", err)
	}

	// Build vault PDA (market + mint)
	vaultPDA, _, err := solana.FindProgramAddress(
		[][]byte{marketPublicKey.Bytes(), nosTokenMint.Bytes()},
		nosanaJobsProgram,
	)
	if err != nil {
		return "", fmt.Errorf("failed to find vault PDA: %w", err)
	}

	// Get latest blockhash
	resp, err := d.RPCClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("failed to get blockhash: %w", err)
	}

	log.Printf("[INFO] 🔐 User Token Account: %s", userTokenAccount.String())
	log.Printf("[INFO] 🏦 Vault PDA: %s", vaultPDA.String())
	log.Printf("[INFO] 🧱 Blockhash: %s", resp.Value.Blockhash.String())

	// Build the instruction for Nosana Jobs program's "list" method
	instruction, err := d.buildNosanaListInstruction(
		ipfsHash, timeout, jobKeypair.PublicKey(), runKeypair.PublicKey(),
		marketPublicKey, userTokenAccount, vaultPDA, nosanaJobsProgram, nosTokenMint,
	)
	if err != nil {
		return "", fmt.Errorf("failed to build instruction: %w", err)
	}

	// Create transaction
	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		resp.Value.Blockhash,
		solana.TransactionPayer(d.PublicKey),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	// Sign transaction with all required signers
	signers := []solana.PrivateKey{
		d.PrivateKey,
		jobKeypair.PrivateKey,
		runKeypair.PrivateKey,
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		for _, signer := range signers {
			if signer.PublicKey().Equals(key) {
				return &signer
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Submit transaction to blockchain
	sig, err := d.RPCClient.SendTransaction(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Printf("[INFO] ✅ Transaction submitted to Nosana blockchain: %s", sig.String())
	log.Printf("[INFO] 🌐 Job will appear on dashboard.nosana.com/account/deployer")

	return sig.String(), nil
}

// buildNosanaListInstruction builds the instruction for Nosana Jobs program's "list" method
func (d *DirectJobSubmission) buildNosanaListInstruction(
	ipfsHash string, timeout int,
	jobAccount, runAccount, marketAccount, userTokenAccount, vaultAccount, programID, mintAccount solana.PublicKey,
) (*solana.GenericInstruction, error) {
	// Decode IPFS hash to bytes (skip first 2 bytes as per SDK)
	ipfsBytes, err := base58.Decode(ipfsHash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IPFS hash: %w", err)
	}
	
	if len(ipfsBytes) < 34 {
		return nil, fmt.Errorf("invalid IPFS hash length")
	}
	
	// Take 32 bytes after the first 2 (as per SDK: subarray(2))
	ipfsHashBytes := ipfsBytes[2:34]

	// Convert timeout to little-endian bytes (i64)
	timeoutBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeoutBytes, uint64(timeout))

	// Build instruction data: method discriminator + args
	// For Anchor programs, method discriminator is first 8 bytes of sha256("global:list")
	methodHash := sha256.Sum256([]byte("global:list"))
	instructionData := make([]byte, 0, 8+32+8)
	instructionData = append(instructionData, methodHash[:8]...)  // method discriminator
	instructionData = append(instructionData, ipfsHashBytes...)   // ipfs hash (32 bytes)
	instructionData = append(instructionData, timeoutBytes...)    // timeout (8 bytes)

	// Build accounts list (order matters for Anchor programs)
	accounts := solana.AccountMetaSlice{
		&solana.AccountMeta{PublicKey: jobAccount, IsWritable: true, IsSigner: true},        // job
		&solana.AccountMeta{PublicKey: marketAccount, IsWritable: true, IsSigner: false},    // market  
		&solana.AccountMeta{PublicKey: runAccount, IsWritable: true, IsSigner: true},        // run
		&solana.AccountMeta{PublicKey: userTokenAccount, IsWritable: true, IsSigner: false}, // user
		&solana.AccountMeta{PublicKey: vaultAccount, IsWritable: true, IsSigner: false},     // vault
		&solana.AccountMeta{PublicKey: d.PublicKey, IsWritable: true, IsSigner: true},       // payer
		&solana.AccountMeta{PublicKey: d.PublicKey, IsWritable: false, IsSigner: true},      // authority
		&solana.AccountMeta{PublicKey: token.ProgramID, IsWritable: false, IsSigner: false}, // token program
		&solana.AccountMeta{PublicKey: system.ProgramID, IsWritable: false, IsSigner: false}, // system program
	}

	return solana.NewInstruction(
		programID,
		accounts,
		instructionData,
	), nil
}