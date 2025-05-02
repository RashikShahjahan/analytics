const http = require('http');
const https = require('https');
const url = require('url');

const PORT = 3000;
const TARGET_HOST = 'analytics.rashik.sh';

const server = http.createServer((req, res) => {
  // Set CORS headers
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  
  // Handle preflight OPTIONS request
  if (req.method === 'OPTIONS') {
    res.writeHead(204);
    res.end();
    return;
  }
  
  // Only handle GET requests to /api
  if (req.method === 'GET' && req.url.startsWith('/api')) {
    const parsedUrl = url.parse(req.url, true);
    const targetPath = parsedUrl.path;
    
    console.log(`Proxying request to: https://${TARGET_HOST}${targetPath}`);
    
    const options = {
      hostname: TARGET_HOST,
      port: 443,
      path: targetPath,
      method: 'GET',
      headers: {
        'Host': TARGET_HOST,
        'User-Agent': req.headers['user-agent'] || 'Analytics Visualizer Proxy'
      }
    };
    
    const proxyReq = https.request(options, (proxyRes) => {
      res.writeHead(proxyRes.statusCode, proxyRes.headers);
      proxyRes.pipe(res);
    });
    
    proxyReq.on('error', (error) => {
      console.error(`Proxy error: ${error.message}`);
      res.writeHead(500);
      res.end(`Proxy error: ${error.message}`);
    });
    
    req.pipe(proxyReq, { end: true });
  } else {
    // Serve the visualizer.html file for the root path
    if (req.url === '/' || req.url === '/index.html') {
      const fs = require('fs');
      fs.readFile('visualizer.html', (err, data) => {
        if (err) {
          res.writeHead(500);
          res.end('Error loading visualizer.html');
          return;
        }
        res.writeHead(200, { 'Content-Type': 'text/html' });
        res.end(data);
      });
    } else {
      // Handle 404 for other paths
      res.writeHead(404);
      res.end('Not Found');
    }
  }
});

server.listen(PORT, () => {
  console.log(`CORS Proxy server running at http://localhost:${PORT}`);
  console.log(`Open http://localhost:${PORT} in your browser to view the visualizer`);
}); 