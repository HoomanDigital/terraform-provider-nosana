#!/usr/bin/env node

/**
 * Nosana Validation Script
 * Validates wallet setup and network connectivity using the official SDK
 * 
 * Usage: node nosana-validate.js <private_key> <network>
 */

import { Client } from '@nosana/sdk';

async function main() {
    try {
        // Parse command line arguments
        const args = process.argv.slice(2);
        if (args.length !== 2) {
            console.error('Usage: node nosana-validate.js <private_key> <network>');
            process.exit(1);
        }

        const [privateKey, network] = args;

        // Validate network
        if (!['mainnet', 'devnet'].includes(network)) {
            throw new Error(`Invalid network: ${network}. Must be 'mainnet' or 'devnet'`);
        }

        // Initialize Nosana client
        console.log(`Connecting to ${network}...`);
        const nosana = new Client(network, privateKey);
        
        // Get wallet info
        const publicKey = nosana.solana.wallet.publicKey.toString();
        console.log(`Wallet public key: ${publicKey}`);

        // Check balances
        const solBalance = await nosana.solana.getSolBalance();
        const nosBalance = await nosana.solana.getNosBalance();

        // Prepare validation result
        const result = {
            success: true,
            network: network,
            wallet_address: publicKey,
            sol_balance: solBalance,
            nos_balance: nosBalance?.amount?.toString() || '0',
            has_sufficient_sol: solBalance >= 0.01, // Minimum SOL for transactions
            has_nos_tokens: nosBalance && nosBalance.amount > 0,
            timestamp: new Date().toISOString()
        };

        // Add warnings if balances are low
        const warnings = [];
        if (solBalance < 0.01) {
            warnings.push('Low SOL balance. You need at least 0.01 SOL for transaction fees.');
        }
        if (!nosBalance || nosBalance.amount <= 0) {
            warnings.push('No NOS tokens found. You need NOS tokens to submit jobs.');
        }

        if (warnings.length > 0) {
            result.warnings = warnings;
        }

        console.log('VALIDATION_JSON:' + JSON.stringify(result));
        process.exit(0);

    } catch (error) {
        const errorResult = {
            success: false,
            error: error.message,
            stack: error.stack,
            network: process.argv[3] || null
        };
        
        console.error('VALIDATION_ERROR_JSON:' + JSON.stringify(errorResult));
        process.exit(1);
    }
}

main();