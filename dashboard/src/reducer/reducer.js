import { combineReducers } from "redux";
import foundReducer from "found/lib/foundReducer";

import themeReducer from "./themeReducer";

const reducer = combineReducers({
  found: foundReducer,
  theme: themeReducer,
});

export default reducer;
