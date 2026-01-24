/**
 * Simple mock server for BugIt development
 * Run with: node server/index.js
 */

const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = 3001;
const mockData = JSON.parse(fs.readFileSync(path.join(__dirname, 'mock-data.json'), 'utf8'));

// Generate mock input events
function generateInputs(durationMs) {
  const keyboard = [];
  const mouse = [];
  const gamepad = [];
  
  // Generate keyboard events (WASD movement)
  const keys = ['W', 'A', 'S', 'D', 'Space', 'Shift'];
  let t = 0;
  while (t < durationMs) {
    const key = keys[Math.floor(Math.random() * keys.length)];
    const duration = 100 + Math.random() * 500;
    keyboard.push({ timestampMs: t, type: 'down', key, keyCode: key.charCodeAt(0) });
    keyboard.push({ timestampMs: t + duration, type: 'up', key, keyCode: key.charCodeAt(0) });
    t += duration + Math.random() * 200;
  }
  
  // Generate mouse clicks
  t = 0;
  while (t < durationMs) {
    mouse.push({
      timestampMs: t,
      type: 'down',
      button: Math.random() > 0.7 ? 2 : 0,
      x: Math.floor(Math.random() * 1920),
      y: Math.floor(Math.random() * 1080),
    });
    t += 500 + Math.random() * 3000;
  }
  
  return { keyboard, mouse, gamepad };
}

// Generate mock frame timing
function generateFrames(durationMs) {
  const samples = [];
  let t = 0;
  const targetFps = 60;
  
  while (t < durationMs) {
    // Simulate occasional FPS drops
    let fps = targetFps + (Math.random() - 0.5) * 10;
    if (Math.random() < 0.05) {
      fps = 20 + Math.random() * 20; // FPS drop
    }
    
    samples.push({
      timestampMs: t,
      frameTimeMs: 1000 / fps,
      fps: fps,
    });
    
    t += 1000 / targetFps;
  }
  
  const fpsList = samples.map(s => s.fps);
  const summary = {
    avgFps: fpsList.reduce((a, b) => a + b, 0) / fpsList.length,
    minFps: Math.min(...fpsList),
    maxFps: Math.max(...fpsList),
    p99FrameTimeMs: 1000 / Math.min(...fpsList),
    stutterCount: samples.filter(s => s.fps < 30).length,
  };
  
  return { samples, summary };
}

// Generate mock logs
function generateLogs(durationMs) {
  const categories = ['PhysicsEngine', 'CharacterMovement', 'AnimationSystem', 'Renderer', 'Audio'];
  const levels = ['verbose', 'log', 'warning', 'error'];
  const logs = [];
  
  let t = 0;
  while (t < durationMs) {
    const level = levels[Math.floor(Math.random() * levels.length)];
    const category = categories[Math.floor(Math.random() * categories.length)];
    
    logs.push({
      timestampMs: t,
      level,
      category,
      message: `Sample ${level} message from ${category} at ${t}ms`,
    });
    
    t += 100 + Math.random() * 500;
  }
  
  return { logs, categories };
}

// Router
function handleRequest(req, res) {
  const url = new URL(req.url, `http://localhost:${PORT}`);
  const pathname = url.pathname.replace('/api/v1', '');
  
  res.setHeader('Content-Type', 'application/json');
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  
  if (req.method === 'OPTIONS') {
    res.writeHead(204);
    res.end();
    return;
  }
  
  // GET /repros
  if (pathname === '/repros' && req.method === 'GET') {
    const page = parseInt(url.searchParams.get('page') || '1');
    const limit = parseInt(url.searchParams.get('limit') || '20');
    const start = (page - 1) * limit;
    
    res.writeHead(200);
    res.end(JSON.stringify({
      repros: mockData.repros.slice(start, start + limit),
      total: mockData.repros.length,
      page,
      pageSize: limit,
    }));
    return;
  }
  
  // GET /repros/:id
  const reproMatch = pathname.match(/^\/repros\/([^\/]+)$/);
  if (reproMatch && req.method === 'GET') {
    const repro = mockData.repros.find(r => r.id === reproMatch[1]);
    if (repro) {
      res.writeHead(200);
      res.end(JSON.stringify(repro));
    } else {
      res.writeHead(404);
      res.end(JSON.stringify({ error: 'Repro not found' }));
    }
    return;
  }
  
  // GET /repros/:id/inputs
  const inputsMatch = pathname.match(/^\/repros\/([^\/]+)\/inputs$/);
  if (inputsMatch && req.method === 'GET') {
    const repro = mockData.repros.find(r => r.id === inputsMatch[1]);
    if (repro) {
      res.writeHead(200);
      res.end(JSON.stringify(generateInputs(repro.durationMs)));
    } else {
      res.writeHead(404);
      res.end(JSON.stringify({ error: 'Repro not found' }));
    }
    return;
  }
  
  // GET /repros/:id/frames
  const framesMatch = pathname.match(/^\/repros\/([^\/]+)\/frames$/);
  if (framesMatch && req.method === 'GET') {
    const repro = mockData.repros.find(r => r.id === framesMatch[1]);
    if (repro) {
      res.writeHead(200);
      res.end(JSON.stringify(generateFrames(repro.durationMs)));
    } else {
      res.writeHead(404);
      res.end(JSON.stringify({ error: 'Repro not found' }));
    }
    return;
  }
  
  // GET /repros/:id/logs
  const logsMatch = pathname.match(/^\/repros\/([^\/]+)\/logs$/);
  if (logsMatch && req.method === 'GET') {
    const repro = mockData.repros.find(r => r.id === logsMatch[1]);
    if (repro) {
      res.writeHead(200);
      res.end(JSON.stringify(generateLogs(repro.durationMs)));
    } else {
      res.writeHead(404);
      res.end(JSON.stringify({ error: 'Repro not found' }));
    }
    return;
  }
  
  // GET /filters
  if (pathname === '/filters' && req.method === 'GET') {
    res.writeHead(200);
    res.end(JSON.stringify(mockData.filters));
    return;
  }
  
  // 404
  res.writeHead(404);
  res.end(JSON.stringify({ error: 'Not found' }));
}

const server = http.createServer(handleRequest);

server.listen(PORT, () => {
  console.log(`Mock BugIt API running at http://localhost:${PORT}`);
  console.log('Endpoints:');
  console.log('  GET /api/v1/repros');
  console.log('  GET /api/v1/repros/:id');
  console.log('  GET /api/v1/repros/:id/inputs');
  console.log('  GET /api/v1/repros/:id/frames');
  console.log('  GET /api/v1/repros/:id/logs');
  console.log('  GET /api/v1/filters');
});
