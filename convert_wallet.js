// Node.js script to convert Phantom wallet private key to Nosana format
const fs = require('fs');
const path = require('path');
const os = require('os');

// Simple base58 decode function
function base58Decode(str) {
    const alphabet = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz';
    let decoded = BigInt(0);
    let multi = BigInt(1);
    
    for (let i = str.length - 1; i >= 0; i--) {
        const char = str[i];
        const index = alphabet.indexOf(char);
        if (index === -1) {
            throw new Error(`Invalid base58 character: ${char}`);
        }
        decoded += BigInt(index) * multi;
        multi *= BigInt(58);
    }
    
    // Convert to bytes
    const bytes = [];
    while (decoded > 0) {
        bytes.unshift(Number(decoded & BigInt(0xff)));
        decoded >>= BigInt(8);
    }
    
    // Add leading zeros for '1' characters
    let leadingOnes = 0;
    for (let i = 0; i < str.length && str[i] === '1'; i++) {
        leadingOnes++;
    }
    
    return new Uint8Array([...Array(leadingOnes).fill(0), ...bytes]);
}

// Get private key from command line argument
const privateKey = process.argv[2];
if (!privateKey) {
    console.error('Usage: node convert_wallet.js <private_key>');
    process.exit(1);
}

try {
    console.log('Converting private key...');
    
    // Decode the private key
    const keyBytes = base58Decode(privateKey);
    console.log(`Private key length: ${keyBytes.length} bytes`);
    
    if (keyBytes.length !== 64) {
        console.warn(`Warning: Expected 64 bytes for Solana private key, got ${keyBytes.length}`);
    }
    
    // Convert to array format
    const keyArray = Array.from(keyBytes);
    
    // Paths
    const nosanaDir = path.join(os.homedir(), '.nosana');
    const keypairPath = path.join(nosanaDir, 'nosana_key.json');
    const backupPath = path.join(nosanaDir, 'nosana_key.json.backup');
    
    // Create directory if it doesn't exist
    if (!fs.existsSync(nosanaDir)) {
        fs.mkdirSync(nosanaDir, { recursive: true });
    }
    
    // Backup existing keypair
    if (fs.existsSync(keypairPath)) {
        fs.copyFileSync(keypairPath, backupPath);
        console.log(`Backed up existing keypair to: ${backupPath}`);
    }
    
    // Write new keypair
    fs.writeFileSync(keypairPath, JSON.stringify(keyArray));
    
    console.log('Keypair updated successfully!');
    console.log(`New keypair saved to: ${keypairPath}`);
    console.log('You can now use your Phantom wallet with Nosana CLI');
    
} catch (error) {
    console.error('Error:', error.message);
    console.error('Failed to convert private key. Please check the format.');
    process.exit(1);
}
