//
// See 'Initializing Test Environment'
// @link https://github.com/facebookincubator/create-react-app/blob/master/packages/react-scripts/template/README.md#initializing-test-environment
//

// Jest enzyme plugin
import "jest-enzyme";

// Mock localStorage
import "jest-localstorage-mock";

// Mock what-wg/fetch
global.fetch = require("jest-fetch-mock");
