import { combineReducers } from "redux";

import themeReducer from "./themeReducer";

const reducer = combineReducers({
  theme: themeReducer,
});

export default reducer;
