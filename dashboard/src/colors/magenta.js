import { darken, lighten } from "material-ui/styles/colorManipulator";

// TODO: Don't use lighten / darken
const magentaColor = "#d86cc5";
const magenta = {
  50: lighten(magentaColor, 0.125),
  100: lighten(magentaColor, 0.1),
  200: lighten(magentaColor, 0.075),
  300: lighten(magentaColor, 0.05),
  400: lighten(magentaColor, 0.025),
  500: magentaColor,
  600: darken(magentaColor, 0.025),
  700: darken(magentaColor, 0.05),
  800: darken(magentaColor, 0.075),
  900: darken(magentaColor, 0.1),
  A100: magentaColor,
  A200: magentaColor,
  A400: magentaColor,
  A700: magentaColor,
  contrastDefaultColor: "light",
};

export default magenta;
