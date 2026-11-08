import pg from 'pg'
import { readConfig } from './config.ts'
import { createAuth } from './auth.ts'

export const config = readConfig()
export const pool = new pg.Pool({ connectionString: config.databaseURL, max: 10, connectionTimeoutMillis: 5000 })
export const auth = createAuth(config, pool)
