#!/usr/bin/env node

/**
 * Setup Script for Terraform Nosana Provider
 * Installs dependencies and validates Node.js environment
 */

import { promises as fs } from 'fs';
import { exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';
import { fileURLToPath } from 'url';

const execAsync = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function main() {
    try {
        console.log('Setting up Terraform Nosana Provider SDK bridge...');

        // Check Node.js version
        const nodeVersion = process.version;
        const majorVersion = parseInt(nodeVersion.slice(1).split('.')[0]);
        
        if (majorVersion < 16) {
            throw new Error(`Node.js version ${nodeVersion} is not supported. Please upgrade to Node.js 16 or higher.`);
        }
        
        console.log(`✓ Node.js version: ${nodeVersion}`);

        // Check if package.json exists
        const packagePath = path.join(__dirname, 'package.json');
        try {
            await fs.access(packagePath);
            console.log('✓ Package.json found');
        } catch {
            throw new Error('package.json not found. Please ensure you\'re running this from the scripts directory.');
        }

        // Install dependencies
        console.log('Installing Nosana SDK dependencies...');
        try {
            const { stdout, stderr } = await execAsync('npm install', { cwd: __dirname });
            console.log('✓ Dependencies installed successfully');
            if (stderr && !stderr.includes('npm WARN')) {
                console.warn('Installation warnings:', stderr);
            }
        } catch (error) {
            throw new Error(`Failed to install dependencies: ${error.message}`);
        }

        // Validate SDK installation
        console.log('Validating SDK installation...');
        try {
            const { default: Client } = await import('@nosana/sdk');
            console.log('✓ Nosana SDK imported successfully');
        } catch (error) {
            throw new Error(`Failed to import Nosana SDK: ${error.message}`);
        }

        // Make scripts executable
        const scripts = ['nosana-job-post.js', 'nosana-job-get.js', 'nosana-validate.js'];
        for (const script of scripts) {
            const scriptPath = path.join(__dirname, script);
            try {
                await fs.chmod(scriptPath, 0o755);
                console.log(`✓ Made ${script} executable`);
            } catch (error) {
                console.warn(`Warning: Could not make ${script} executable: ${error.message}`);
            }
        }

        console.log('\n🎉 Setup completed successfully!');
        console.log('\nThe Terraform Nosana Provider is now ready to use the official Nosana SDK.');
        console.log('\nNext steps:');
        console.log('1. Ensure you have a Solana wallet with SOL and NOS tokens');
        console.log('2. Set your private key in the Terraform configuration');
        console.log('3. Run terraform plan/apply');

        process.exit(0);

    } catch (error) {
        console.error('\n❌ Setup failed:', error.message);
        console.error('\nPlease resolve the issue and run setup again.');
        process.exit(1);
    }
}

main();