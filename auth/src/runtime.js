import pg from 'pg'
import { readConfig } from './config.js'
import { createAuth } from './auth.js'

export const config = readConfig()
export const pool = new pg.Pool({ connectionString: config.databaseURL, max: 10, connectionTimeoutMillis: 5000 })
export const auth = createAuth(config, pool)
