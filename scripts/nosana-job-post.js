#!/usr/bin/env node

/**
 * Nosana Job Post Script
 * Posts a job to the Nosana network using the official SDK
 * 
 * Usage: node nosana-job-post.js <private_key> <network> <market_address> <job_json>
 */

import { PublicKey } from '@solana/web3.js';
import { Client } from '@nosana/sdk';
import { promises as fs } from 'fs';

async function main() {
    try {
        // Parse command line arguments
        const args = process.argv.slice(2);
        if (args.length !== 4) {
            console.error('Usage: node nosana-job-post.js <private_key> <network> <market_address> <job_json_file_or_string>');
            process.exit(1);
        }

        const [privateKey, network, marketAddress, jobInput] = args;

        // Validate network
        if (!['mainnet', 'devnet'].includes(network)) {
            throw new Error(`Invalid network: ${network}. Must be 'mainnet' or 'devnet'`);
        }

        // Parse job definition
        let jobDefinition;
        try {
            // Try to read as file first, then as JSON string
            if (jobInput.endsWith('.json') || jobInput.startsWith('/') || jobInput.startsWith('./')) {
                const fileContent = await fs.readFile(jobInput, 'utf8');
                jobDefinition = JSON.parse(fileContent);
            } else {
                jobDefinition = JSON.parse(jobInput);
            }
        } catch (error) {
            throw new Error(`Failed to parse job definition: ${error.message}`);
        }

        // Initialize Nosana client
        const nosana = new Client(network, privateKey);
        
        // Log connection info
        console.log(`Connected with wallet: ${nosana.solana.wallet.publicKey.toString()}`);
        
        // Check balances
        const solBalance = await nosana.solana.getSolBalance();
        const nosBalance = await nosana.solana.getNosBalance();
        console.log(`SOL balance: ${solBalance} SOL`);
        console.log(`NOS balance: ${nosBalance?.amount.toString() || '0'} NOS`);

        // Upload job definition to IPFS
        console.log('Uploading job definition to IPFS...');
        const ipfsHash = await nosana.ipfs.pin(jobDefinition);
        console.log(`IPFS hash: ${ipfsHash}`);

        // Post job to market
        console.log(`Posting job to market: ${marketAddress}`);
        const market = new PublicKey(marketAddress);
        const response = await nosana.jobs.list(ipfsHash, market);

        // Output result as JSON for Go to parse
        const result = {
            success: true,
            job_id: response.job.toString(),
            transaction_id: response.transaction_id || response.signature,
            ipfs_hash: ipfsHash,
            market_address: marketAddress,
            dashboard_url: `https://dashboard.nosana.com/jobs/${response.job}`,
            market_url: `https://dashboard.nosana.com/markets/${market.toBase58()}`
        };

        console.log('JOB_RESULT_JSON:' + JSON.stringify(result));
        process.exit(0);

    } catch (error) {
        const errorResult = {
            success: false,
            error: error.message,
            stack: error.stack
        };
        
        console.error('JOB_ERROR_JSON:' + JSON.stringify(errorResult));
        process.exit(1);
    }
}

main();