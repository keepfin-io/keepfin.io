import { createAuthClient } from 'better-auth/vue'
import { magicLinkClient } from 'better-auth/client/plugins'

// Tokens remain in HttpOnly cookies; Go verifies every application request.
export const authClient = createAuthClient({
  basePath: '/api/auth',
  fetchOptions: { timeout: 10000 },
  plugins: [magicLinkClient()],
})
