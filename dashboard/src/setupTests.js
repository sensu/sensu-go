//
// See 'Initializing Test Environment'
// @link https://github.com/facebookincubator/create-react-app/blob/master/packages/react-scripts/template/README.md#initializing-test-environment
//

// Jest enzyme plugin
import "jest-enzyme";
import { configure as configureEnzyme } from "enzyme";
import Adapter from "enzyme-adapter-react-16";

// Mock localStorage
import "jest-localstorage-mock";

// Configure enzyme
configureEnzyme({ adapter: new Adapter() });

// Mock what-wg/fetch
global.fetch = require("jest-fetch-mock");
