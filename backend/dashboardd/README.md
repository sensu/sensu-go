## Sensu Dashboard

### Development

##### `yarn start`

Starts a webpack server that listens on
[http://localhost:3030](http://localhost:3030) for easier development (the page
will reload if you make edits).

We use webpack proxy mechanism (see `proxy` attribute in `package.json`) to
forward requests destined to the API in order to avoid any CORS issue.

### Releasing

##### `yarn run build`

Builds the app for production to the build folder. It correctly bundles React in
production mode and optimizes and minifies the build for the best performance.

##### `yarn run static-assets`

Compiles the dashboard static assets into the **sensu-backend** binary using the
[fileb0x](https://github.com/UnnoTed/fileb0x) utility. These assets must first
be build with the step above. The embedded assets live in the
`backend/dashboardd/ab0x.go` file, which is added to source control so all
developers don't need to install all frontend development tools.
