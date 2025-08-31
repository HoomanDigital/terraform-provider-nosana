#!/usr/bin/env node

/**
 * Nosana Job Get Script
 * Gets job status and details from the Nosana network using the official SDK
 * 
 * Usage: node nosana-job-get.js <private_key> <network> <job_id>
 */

import { PublicKey } from '@solana/web3.js';
import { Client } from '@nosana/sdk';

async function main() {
    try {
        // Parse command line arguments
        const args = process.argv.slice(2);
        if (args.length !== 3) {
            console.error('Usage: node nosana-job-get.js <private_key> <network> <job_id>');
            process.exit(1);
        }

        const [privateKey, network, jobId] = args;

        // Validate network
        if (!['mainnet', 'devnet'].includes(network)) {
            throw new Error(`Invalid network: ${network}. Must be 'mainnet' or 'devnet'`);
        }

        // Initialize Nosana client
        const nosana = new Client(network, privateKey);
        
        // Get job details
        console.log(`Fetching job details for: ${jobId}`);
        const job = await nosana.jobs.get(new PublicKey(jobId));

        if (!job) {
            throw new Error(`Job not found: ${jobId}`);
        }

        // Map job state to consistent status
        let status = 'UNKNOWN';
        switch (job.state) {
            case 'QUEUED':
                status = 'PENDING';
                break;
            case 'RUNNING':
                status = 'RUNNING';
                break;
            case 'COMPLETED':
                status = 'COMPLETED';
                break;
            case 'FAILED':
                status = 'FAILED';
                break;
            default:
                status = job.state || 'UNKNOWN';
        }

        // Prepare result
        const result = {
            success: true,
            job_id: jobId,
            status: status,
            state: job.state,
            ipfs_job: job.ipfsJob,
            ipfs_result: job.ipfsResult,
            market: job.market?.toString(),
            node: job.node?.toString(),
            price: job.price?.toString(),
            time_start: job.timeStart ? new Date(job.timeStart * 1000).toISOString() : null,
            time_end: job.timeEnd ? new Date(job.timeEnd * 1000).toISOString() : null,
            dashboard_url: `https://dashboard.nosana.com/jobs/${jobId}`
        };

        // If job is completed and has results, try to fetch them
        if (job.state === 'COMPLETED' && job.ipfsResult) {
            try {
                const jobResult = await nosana.ipfs.retrieve(job.ipfsResult);
                result.result_data = jobResult;
            } catch (error) {
                console.warn(`Failed to fetch job results: ${error.message}`);
                result.result_error = error.message;
            }
        }

        console.log('JOB_STATUS_JSON:' + JSON.stringify(result));
        process.exit(0);

    } catch (error) {
        const errorResult = {
            success: false,
            error: error.message,
            stack: error.stack,
            job_id: process.argv[4] || null
        };
        
        console.error('JOB_ERROR_JSON:' + JSON.stringify(errorResult));
        process.exit(1);
    }
}

main();