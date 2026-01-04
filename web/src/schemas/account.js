import { z } from 'zod'

// Valid currency codes
const currencyCodes = [
    'USD', 'EUR', 'GBP', 'TRY', 'JPY', 'CAD', 'AUD', 'CHF',
    'CNY', 'INR', 'KRW', 'SGD', 'HKD', 'SEK', 'NOK', 'DKK',
    'PLN', 'BRL', 'MXN', 'RUB'
]

// Create Account Schema
export const createAccountSchema = z.object({
    account_name: z
        .string()
        .min(1, 'Account name is required')
        .max(100, 'Account name must be less than 100 characters'),
    currency: z
        .string()
        .min(1, 'Currency is required')
        .refine((val) => currencyCodes.includes(val), {
            message: 'Please select a valid currency'
        }),
    initial_balance: z
        .number()
        .min(0, 'Balance cannot be negative')
        .optional()
        .default(0),
    platform: z
        .string()
        .max(50, 'Platform name must be less than 50 characters')
        .optional()
        .default(''),
    description: z
        .string()
        .max(255, 'Description must be less than 255 characters')
        .optional()
        .default(''),
    is_excluded: z
        .boolean()
        .optional()
        .default(false)
})

// Update Account Schema (all fields optional except what's being updated)
export const updateAccountSchema = z.object({
    account_name: z
        .string()
        .min(1, 'Account name is required')
        .max(100, 'Account name must be less than 100 characters')
        .optional(),
    currency: z
        .string()
        .refine((val) => !val || currencyCodes.includes(val), {
            message: 'Please select a valid currency'
        })
        .optional(),
    initial_balance: z
        .number()
        .min(0, 'Balance cannot be negative')
        .optional(),
    platform: z
        .string()
        .max(50, 'Platform name must be less than 50 characters')
        .optional(),
    description: z
        .string()
        .max(255, 'Description must be less than 255 characters')
        .optional(),
    is_excluded: z
        .boolean()
        .optional()
})

// Currency codes export for reuse
export const CURRENCY_CODES = currencyCodes

/** @typedef {z.infer<typeof createAccountSchema>} CreateAccountInput */
/** @typedef {z.infer<typeof updateAccountSchema>} UpdateAccountInput */
