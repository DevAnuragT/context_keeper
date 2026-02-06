#!/usr/bin/env node

/**
 * Simple static file server for the MCP Context Engine landing page
 * For development and testing purposes
 * 
 * Usage: node server.js [port]
 * Default port: 3000
 */

const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = process.argv[2] || 3000;
const PUBLIC_DIR = __dirname;

// MIME types for different file extensions
const MIME_TYPES = {
    '.html': 'text/html',
    '.css': 'text/css',
    '.js': 'text/javascript',
    '.json': 'application/json',
    '.png': 'image/png',
    '.jpg': 'image/jpeg',
    '.jpeg': 'image/jpeg',
    '.gif': 'image/gif',
    '.svg': 'image/svg+xml',
    '.ico': 'image/x-icon',
    '.woff': 'font/woff',
    '.woff2': 'font/woff2',
    '.ttf': 'font/ttf',
    '.eot': 'application/vnd.ms-fontobject',
    '.otf': 'font/otf',
    '.webp': 'image/webp',
    '.mp4': 'video/mp4',
    '.webm': 'video/webm',
    '.txt': 'text/plain',
    '.md': 'text/markdown'
};

// Route mappings for clean URLs
const ROUTES = {
    '/': '/index.html',
    '/signup': '/signup.html',
    '/login': '/login.html',
    '/reset-password': '/reset-password.html',
    '/reset-password-confirm': '/reset-password-confirm.html',
    '/verify-email': '/verify-email.html',
    '/auth/callback': '/auth-callback.html'
};

const server = http.createServer((req, res) => {
    // Parse URL
    let filePath = req.url.split('?')[0]; // Remove query string
    
    // Handle route mappings
    if (ROUTES[filePath]) {
        filePath = ROUTES[filePath];
    }
    
    // Default to index.html for root
    if (filePath === '/') {
        filePath = '/index.html';
    }
    
    // Construct full file path
    const fullPath = path.join(PUBLIC_DIR, filePath);
    
    // Security check: prevent directory traversal
    if (!fullPath.startsWith(PUBLIC_DIR)) {
        res.writeHead(403, { 'Content-Type': 'text/plain' });
        res.end('403 Forbidden');
        return;
    }
    
    // Get file extension
    const ext = path.extname(fullPath).toLowerCase();
    const contentType = MIME_TYPES[ext] || 'application/octet-stream';
    
    // Read and serve file
    fs.readFile(fullPath, (err, content) => {
        if (err) {
            if (err.code === 'ENOENT') {
                // File not found - serve 404
                res.writeHead(404, { 'Content-Type': 'text/html' });
                res.end(`
                    <!DOCTYPE html>
                    <html>
                    <head>
                        <title>404 - Page Not Found</title>
                        <style>
                            body {
                                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
                                display: flex;
                                align-items: center;
                                justify-content: center;
                                height: 100vh;
                                margin: 0;
                                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                                color: white;
                            }
                            .container {
                                text-align: center;
                            }
                            h1 {
                                font-size: 72px;
                                margin: 0;
                            }
                            p {
                                font-size: 24px;
                                margin: 20px 0;
                            }
                            a {
                                color: white;
                                text-decoration: none;
                                border: 2px solid white;
                                padding: 12px 24px;
                                border-radius: 8px;
                                display: inline-block;
                                margin-top: 20px;
                                transition: all 0.2s ease;
                            }
                            a:hover {
                                background: white;
                                color: #667eea;
                            }
                        </style>
                    </head>
                    <body>
                        <div class="container">
                            <h1>404</h1>
                            <p>Page not found</p>
                            <a href="/">Go Home</a>
                        </div>
                    </body>
                    </html>
                `);
            } else {
                // Server error
                res.writeHead(500, { 'Content-Type': 'text/plain' });
                res.end('500 Internal Server Error');
                console.error('Server error:', err);
            }
        } else {
            // Success - serve file
            res.writeHead(200, { 
                'Content-Type': contentType,
                'Cache-Control': 'public, max-age=3600' // Cache for 1 hour
            });
            res.end(content);
        }
    });
    
    // Log request
    const timestamp = new Date().toISOString();
    console.log(`[${timestamp}] ${req.method} ${req.url} - ${res.statusCode || 'pending'}`);
});

server.listen(PORT, () => {
    console.log('='.repeat(60));
    console.log('MCP Context Engine - Landing Page Server');
    console.log('='.repeat(60));
    console.log(`Server running at http://localhost:${PORT}/`);
    console.log(`Serving files from: ${PUBLIC_DIR}`);
    console.log('');
    console.log('Available pages:');
    console.log(`  - http://localhost:${PORT}/                      (Landing page)`);
    console.log(`  - http://localhost:${PORT}/signup                (Signup page)`);
    console.log(`  - http://localhost:${PORT}/login                 (Login page)`);
    console.log(`  - http://localhost:${PORT}/reset-password        (Password reset)`);
    console.log(`  - http://localhost:${PORT}/verify-email          (Email verification)`);
    console.log('');
    console.log('Press Ctrl+C to stop the server');
    console.log('='.repeat(60));
});

// Handle server errors
server.on('error', (err) => {
    if (err.code === 'EADDRINUSE') {
        console.error(`Error: Port ${PORT} is already in use`);
        console.error('Try a different port: node server.js <port>');
    } else {
        console.error('Server error:', err);
    }
    process.exit(1);
});

// Graceful shutdown
process.on('SIGINT', () => {
    console.log('\n\nShutting down server...');
    server.close(() => {
        console.log('Server stopped');
        process.exit(0);
    });
});

process.on('SIGTERM', () => {
    console.log('\n\nShutting down server...');
    server.close(() => {
        console.log('Server stopped');
        process.exit(0);
    });
});
