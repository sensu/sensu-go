import { darken, lighten } from "@material-ui/core/styles/colorManipulator";

// TODO: Don't use lighten / darken
const salmonColor = "#FA8072";
const salmon = {
  50: lighten(salmonColor, 0.25),
  100: lighten(salmonColor, 0.2),
  200: lighten(salmonColor, 0.15),
  300: lighten(salmonColor, 0.1),
  400: lighten(salmonColor, 0.05),
  500: salmonColor,
  600: darken(salmonColor, 0.05),
  700: darken(salmonColor, 0.1),
  800: darken(salmonColor, 0.15),
  900: darken(salmonColor, 0.2),
  A100: salmonColor,
  A200: salmonColor,
  A400: salmonColor,
  A700: salmonColor,
  contrastDefaultColor: "light",
};

export default salmon;
