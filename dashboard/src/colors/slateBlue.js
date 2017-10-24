import { darken, lighten } from "material-ui/styles/colorManipulator";

// TODO: Don't use lighten / darken
const slateBlueColor = "#717CE5";
const slateBlue = {
  50: lighten(slateBlueColor, 0.125),
  100: lighten(slateBlueColor, 0.1),
  200: lighten(slateBlueColor, 0.075),
  300: lighten(slateBlueColor, 0.05),
  400: lighten(slateBlueColor, 0.025),
  500: slateBlueColor,
  600: darken(slateBlueColor, 0.025),
  700: darken(slateBlueColor, 0.05),
  800: darken(slateBlueColor, 0.075),
  900: darken(slateBlueColor, 0.1),
  A100: slateBlueColor,
  A200: slateBlueColor,
  A400: slateBlueColor,
  A700: slateBlueColor,
  contrastDefaultColor: "light",
};

export default slateBlue;
