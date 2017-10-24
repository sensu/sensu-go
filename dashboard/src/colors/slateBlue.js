import { darken, lighten } from "material-ui/styles/colorManipulator";

// TODO: Don't use lighten / darken
const slateBlueColor = "#717CE5";
const slateBlue = {
  50: lighten(slateBlueColor, 0.25),
  100: lighten(slateBlueColor, 0.2),
  200: lighten(slateBlueColor, 0.15),
  300: lighten(slateBlueColor, 0.1),
  400: lighten(slateBlueColor, 0.05),
  500: slateBlueColor,
  600: darken(slateBlueColor, 0.05),
  700: darken(slateBlueColor, 0.1),
  800: darken(slateBlueColor, 0.15),
  900: darken(slateBlueColor, 0.2),
  A100: slateBlueColor,
  A200: slateBlueColor,
  A400: slateBlueColor,
  A700: slateBlueColor,
  contrastDefaultColor: "light",
};

export default slateBlue;
