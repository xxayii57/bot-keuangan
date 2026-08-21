import dotenv from 'dotenv';
dotenv.config();

const LOCAL_ACCESS_TOKEN = process.env.LOCAL_ACCESS_TOKEN || 'local-access-token-12345';

export function getLocalToken() {
  return LOCAL_ACCESS_TOKEN;
}

export function authGuard(req, res, next) {
  // Allow health check without token
  if (req.path === '/api/health' || req.path === '/health' || req.path === '/' || req.path.startsWith('/api/auth/gemini/callback')) {
    return next();
  }

  const tokenHeader = req.headers['authorization'];
  const tokenCustom = req.headers['x-local-token'];
  const tokenQuery = req.query.token;

  let requestToken = tokenCustom || tokenQuery;

  if (tokenHeader && tokenHeader.startsWith('Bearer ')) {
    requestToken = tokenHeader.split(' ')[1];
  }

  if (!requestToken) {
    return res.status(401).json({
      error: 'Missing access token. Please provide your LOCAL_ACCESS_TOKEN in headers or query parameters.'
    });
  }

  if (requestToken !== LOCAL_ACCESS_TOKEN) {
    return res.status(403).json({
      error: 'Invalid local access token.'
    });
  }

  next();
}
