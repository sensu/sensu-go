import { combineReducers } from "redux";

const initialState = {
  dark: false,
  theme: "sensu",
};

function themeReducer(state = initialState.theme, action) {
  if (action.type === "@@storage/CHANGED") {
    return action.payload.theme || state;
  }
  if (action.type === "theme/CHANGE") {
    return action.payload.theme;
  }
  return state;
}

function darkModeReducer(state = initialState.dark, action) {
  if (action.type === "@@storage/CHANGED") {
    if (typeof action.payload.dark === "undefined") {
      return state;
    }
    return action.payload.dark;
  }
  if (action.type === "theme/TOGGLE_DARK_MODE") {
    return !state;
  }
  return state;
}

export default combineReducers({
  theme: themeReducer,
  dark: darkModeReducer,
});
