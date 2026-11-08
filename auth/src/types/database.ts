import type { ColumnType, Generated } from 'kysely'

// Schema used by operator/seed queries. Migration ownership remains in Go and
// Better Auth; this interface does not create tables or grant database access.
export interface OperatorDatabase {
  users: ApplicationUserTable
  auth_user: AuthUserTable
  auth_account: AuthAccountTable
}

export interface ApplicationUserTable {
  id: string
  hash_id: string
  is_dummy: boolean
  email: string
  password_hash: string
  auth_user_id: string | null
  // node-postgres reads PostgreSQL bigint as a string, but accepts numeric writes.
  user_status: ColumnType<string, number, number>
  user_role: 'user' | 'admin'
  theme: 'light' | 'dark'
  base_currency: string
  locale: string
  created_at: Date
  updated_at: Date
  deleted_at: Generated<Date | null>
}

export interface AuthUserTable {
  id: string
  name: string
  email: string
  emailVerified: boolean
  image: Generated<string | null>
  createdAt: Date
  updatedAt: Date
}

export interface AuthAccountTable {
  id: string
  accountId: string
  providerId: string
  userId: string
  password: string | null
  createdAt: Date
  updatedAt: Date
}
