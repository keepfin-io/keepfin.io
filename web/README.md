# web

Use Node.js 26 (`nvm use` from the repository root). Development Compose and the
production frontend builder use `node:26-alpine`; the auth service uses the same
major version. `npm ci` installs the locked dependencies.

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

### Compiles and minifies for production
```
npm run build
```

### Lints and fixes files
```
npm run lint
```

### Account analysis and asset details

The account page links to `/accounts/:id/analysis`, which reuses the portfolio
report with `account_id` and the account currency. Asset names and the view action
open `/accounts/:id/holdings/:holdingId` for the position, daily closing prices
and filtered, paginated account activities. Missing prices and failed requests
remain explicit; the frontend does not calculate portfolio returns.

Run `npm run build` followed by `npm run test:ui` to check the rendered app in
jsdom, including these routes, account scoping and retry behavior. These checks
simulate viewport sizes; they do not replace browser visual verification.

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).
