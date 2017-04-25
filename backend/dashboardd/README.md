## Sensu Dashboard

### `npm start`

Starts a webpack server that listens on
[http://localhost:3030](http://localhost:3030) for easier development (the page
will reload if you make edits).

We use webpack proxy mechanism (see `proxy` attribute in `package.json`) to
forward requests destined to the API in order to avoid any CORS issue.

### `npm run build`

Builds the app for production to the build folder. It correctly bundles React in
production mode and optimizes and minifies the build for the best performance.
