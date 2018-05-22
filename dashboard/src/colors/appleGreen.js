import { darken, lighten } from "@material-ui/core/styles/colorManipulator";

const appleGreenColor = "#82C023";
const appleGreen = {
  50: lighten(appleGreenColor, 0.25),
  100: lighten(appleGreenColor, 0.2),
  200: lighten(appleGreenColor, 0.15),
  300: lighten(appleGreenColor, 0.1),
  400: lighten(appleGreenColor, 0.05),
  500: appleGreenColor,
  600: darken(appleGreenColor, 0.05),
  700: darken(appleGreenColor, 0.1),
  800: darken(appleGreenColor, 0.15),
  900: darken(appleGreenColor, 0.2),
  A100: appleGreenColor, // TODO
  A200: appleGreenColor, // TODO
  A400: appleGreenColor, // TODO
  A700: appleGreenColor, // TODO
  contrastDefaultColor: "light",
};

export default appleGreen;
